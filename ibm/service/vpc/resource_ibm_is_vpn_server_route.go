// Copyright IBM Corp. 2022 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package vpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/IBM/vpc-go-sdk/vpcv1"
)

const (
	isVPNServerRouteStatusPending  = "pending"
	isVPNServerRouteStatusUpdating = "updating"
	isVPNServerRouteStatusStable   = "stable"
	isVPNServerRouteStatusFailed   = "failed"

	isVPNServerRouteStatusDeleting = "deleting"
	isVPNServerRouteStatusDeleted  = "deleted"
)

func ResourceIBMIsVPNServerRoute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIBMIsVPNServerRouteCreate,
		ReadContext:   resourceIBMIsVPNServerRouteRead,
		UpdateContext: resourceIBMIsVPNServerRouteUpdate,
		DeleteContext: resourceIBMIsVPNServerRouteDelete,
		Importer:      &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"vpn_server": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The VPN server identifier.",
			},
			"vpn_route": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VPN route identifier.",
			},
			"destination": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.InvokeValidator("ibm_is_vpn_server_route", "destination"),
				Description:  "The destination to use for this VPN route in the VPN server. Must be unique within the VPN server. If an incoming packet does not match any destination, it will be dropped.",
			},
			"action": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "deliver",
				ValidateFunc: validate.InvokeValidator("ibm_is_vpn_server_route", "action"),
				Description:  "The action to perform with a packet matching the VPN route:- `translate`: translate the source IP address to one of the private IP addresses of the VPN server, then deliver the packet to target.- `deliver`: deliver the packet to the target.- `drop`: drop the packetThe enumerated values for this property are expected to expand in the future. When processing this property, check for and log unknown values. Optionally halt processing and surface the error, or bypass the VPN route on which the unexpected property value was encountered.",
			},
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.InvokeValidator("ibm_is_vpn_server_route", "name"),
				Description:  "The user-defined name for this VPN route. If unspecified, the name will be a hyphenated list of randomly-selected words. Names must be unique within the VPN server the VPN route resides in.",
			},
			"created_at": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time that the VPN route was created.",
			},
			"health_state": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The health of this resource.- `ok`: Healthy- `degraded`: Suffering from compromised performance, capacity, or connectivity- `faulted`: Completely unreachable, inoperative, or otherwise entirely incapacitated- `inapplicable`: The health state does not apply because of the current lifecycle state. A resource with a lifecycle state of `failed` or `deleting` will have a health state of `inapplicable`. A `pending` resource may also have this state.",
			},
			"health_reasons": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A snake case string succinctly identifying the reason for this health state.",
						},

						"message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "An explanation of the reason for this health state.",
						},

						"more_info": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Link to documentation about the reason for this health state.",
						},
					},
				},
			},
			"href": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL for this VPN route.",
			},
			"lifecycle_state": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The lifecycle state of the VPN route.",
			},
			"lifecycle_reasons": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The reasons for the current lifecycle_state (if any).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A snake case string succinctly identifying the reason for this lifecycle state.",
						},

						"message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "An explanation of the reason for this lifecycle state.",
						},

						"more_info": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Link to documentation about the reason for this lifecycle state.",
						},
					},
				},
			},
			"resource_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource type.",
			},
		},
	}
}

func ResourceIBMIsVPNServerRouteValidator() *validate.ResourceValidator {
	validateSchema := make([]validate.ValidateSchema, 0)
	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 "destination",
			ValidateFunctionIdentifier: validate.ValidateRegexp,
			Type:                       validate.TypeString,
			Required:                   true,
			Regexp:                     `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/(3[0-2]|[1-2][0-9]|[0-9]))$`,
		},
		validate.ValidateSchema{
			Identifier:                 "action",
			ValidateFunctionIdentifier: validate.ValidateAllowedStringValue,
			Type:                       validate.TypeString,
			Optional:                   true,
			AllowedValues:              "deliver, drop, translate",
		},
		validate.ValidateSchema{
			Identifier:                 "name",
			ValidateFunctionIdentifier: validate.ValidateRegexpLen,
			Type:                       validate.TypeString,
			Optional:                   true,
			Regexp:                     `^([a-z]|[a-z][-a-z0-9]*[a-z0-9])$`,
			MinValueLength:             1,
			MaxValueLength:             63,
		},
	)

	resourceValidator := validate.ResourceValidator{ResourceName: "ibm_is_vpn_server_route", Schema: validateSchema}
	return &resourceValidator
}

func resourceIBMIsVPNServerRouteCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcClient, err := meta.(conns.ClientSession).VpcV1API()
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "create", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	createVPNServerRouteOptions := &vpcv1.CreateVPNServerRouteOptions{}

	createVPNServerRouteOptions.SetVPNServerID(d.Get("vpn_server").(string))
	createVPNServerRouteOptions.SetDestination(d.Get("destination").(string))
	if _, ok := d.GetOk("action"); ok {
		createVPNServerRouteOptions.SetAction(d.Get("action").(string))
	}
	if _, ok := d.GetOk("name"); ok {
		createVPNServerRouteOptions.SetName(d.Get("name").(string))
	}

	vpnServerRoute, _, err := vpcClient.CreateVPNServerRouteWithContext(context, createVPNServerRouteOptions)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("CreateVPNServerRouteWithContext failed: %s", err.Error()), "ibm_is_vpn_server_route", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	d.SetId(fmt.Sprintf("%s/%s", *createVPNServerRouteOptions.VPNServerID, *vpnServerRoute.ID))

	_, err = isWaitForVPNServerRouteStable(context, vpcClient, d, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForVPNServerRouteStable failed: %s", err.Error()), "ibm_is_vpn_server_route", "create")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	return resourceIBMIsVPNServerRouteRead(context, d, meta)
}

func isWaitForVPNServerRouteStable(context context.Context, vpcClient *vpcv1.VpcV1, d *schema.ResourceData, timeout time.Duration) (interface{}, error) {

	log.Printf("Waiting for VPN Server  Route(%s) to be stable.", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{isVPNServerStatusPending, isVPNServerRouteStatusUpdating},
		Target:  []string{isVPNServerStatusStable, isVPNServerRouteStatusFailed},
		Refresh: func() (interface{}, string, error) {
			getVPNServerRouteOptions := &vpcv1.GetVPNServerRouteOptions{}

			parts, err := flex.SepIdParts(d.Id(), "/")
			if err != nil {
				return nil, "", flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "wait", "sep-id-parts")
			}

			getVPNServerRouteOptions.SetVPNServerID(parts[0])
			getVPNServerRouteOptions.SetID(parts[1])

			vpnServerRoute, _, err := vpcClient.GetVPNServerRouteWithContext(context, getVPNServerRouteOptions)
			if err != nil {
				tfErr := flex.TerraformErrorf(err, fmt.Sprintf("GetVPNServerRouteWithContext failed: %s", err.Error()), "ibm_is_vpn_server_route", "wait")
				log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
				return vpnServerRoute, "", tfErr
			}

			if *vpnServerRoute.LifecycleState == "stable" || *vpnServerRoute.LifecycleState == "failed" {
				return vpnServerRoute, *vpnServerRoute.LifecycleState, nil
			}
			return vpnServerRoute, *vpnServerRoute.LifecycleState, nil
		},
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}
func resourceIBMIsVPNServerRouteRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := vpcClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	getVPNServerRouteOptions := &vpcv1.GetVPNServerRouteOptions{}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		return diag.FromErr(err)
	}

	getVPNServerRouteOptions.SetVPNServerID(parts[0])
	getVPNServerRouteOptions.SetID(parts[1])

	vpnServerRoute, response, err := sess.GetVPNServerRouteWithContext(context, getVPNServerRouteOptions)
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "sep-id-parts").GetDiag()
	}

	if err = d.Set("vpn_server", getVPNServerRouteOptions.VPNServerID); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-vpn_server").GetDiag()
	}
	if err = d.Set("vpn_route", getVPNServerRouteOptions.ID); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-vpn_route").GetDiag()
	}

	if err = d.Set("destination", vpnServerRoute.Destination); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-destination").GetDiag()
	}
	if err = d.Set("action", vpnServerRoute.Action); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-action").GetDiag()
	}
	if err = d.Set("name", vpnServerRoute.Name); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-name").GetDiag()
	}
	if err = d.Set("created_at", flex.DateTimeToString(vpnServerRoute.CreatedAt)); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-created_at").GetDiag()
	}
	if err = d.Set("health_state", vpnServerRoute.HealthState); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-health_state").GetDiag()
	}
	if vpnServerRoute.HealthReasons != nil {
		if err := d.Set("health_reasons", resourceVPNServerRouteFlattenHealthReasons(vpnServerRoute.HealthReasons)); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-health_reasons").GetDiag()
		}
	}
	if err = d.Set("href", vpnServerRoute.Href); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-href").GetDiag()
	}
	if err = d.Set("lifecycle_state", vpnServerRoute.LifecycleState); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-lifecycle_state").GetDiag()
	}
	if vpnServerRoute.LifecycleReasons != nil {
		if err := d.Set("lifecycle_reasons", resourceVPNServerRouteFlattenLifecycleReasons(vpnServerRoute.LifecycleReasons)); err != nil {
			return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-lifecycle_reasons").GetDiag()
		}
	}
	if err = d.Set("resource_type", vpnServerRoute.ResourceType); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "read", "set-resource_type").GetDiag()
	}

	return nil
}

func resourceIBMIsVPNServerRouteUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcClient, err := meta.(conns.ClientSession).VpcV1API()
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "update", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	updateVPNServerRouteOptions := &vpcv1.UpdateVPNServerRouteOptions{}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "update", "sep-id-parts").GetDiag()
	}

	updateVPNServerRouteOptions.SetVPNServerID(parts[0])
	updateVPNServerRouteOptions.SetID(parts[1])

	hasChange := false

	patchVals := &vpcv1.VPNServerRoutePatch{}

	if d.HasChange("name") {
		patchVals.Name = core.StringPtr(d.Get("name").(string))
		hasChange = true
	}

	if hasChange {
		updateVPNServerRouteOptions.VPNServerRoutePatch, _ = patchVals.AsPatch()
		_, _, err := vpcClient.UpdateVPNServerRouteWithContext(context, updateVPNServerRouteOptions)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("UpdateVPNServerRouteWithContext failed: %s", err.Error()), "ibm_is_vpn_server_route", "update")
			log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
	}
	_, err = isWaitForVPNServerRouteStable(context, vpcClient, d, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForVPNServerRouteStable failed: %s", err.Error()), "ibm_is_vpn_server_route", "update")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	return resourceIBMIsVPNServerRouteRead(context, d, meta)
}

func resourceIBMIsVPNServerRouteDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcClient, err := meta.(conns.ClientSession).VpcV1API()
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "delete", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}

	deleteVPNServerRouteOptions := &vpcv1.DeleteVPNServerRouteOptions{}

	parts, err := flex.SepIdParts(d.Id(), "/")
	if err != nil {
		return flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "delete", "sep-id-parts").GetDiag()
	}

	deleteVPNServerRouteOptions.SetVPNServerID(parts[0])
	deleteVPNServerRouteOptions.SetID(parts[1])

	_, err = vpcClient.DeleteVPNServerRouteWithContext(context, deleteVPNServerRouteOptions)

	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("DeleteVPNServerRouteWithContext failed: %s", err.Error()), "ibm_is_vpn_server_route", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	_, err = isWaitForVPNServerRouteDeleted(context, vpcClient, d)
	if err != nil {
		tfErr := flex.TerraformErrorf(err, fmt.Sprintf("isWaitForVPNServerRouteStable failed: %s", err.Error()), "ibm_is_vpn_server_route", "delete")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	d.SetId("")

	return nil
}

func isWaitForVPNServerRouteDeleted(context context.Context, sess *vpcv1.VpcV1, d *schema.ResourceData) (interface{}, error) {
	log.Printf("Waiting for VPN Server Route(%s) to be deleted.", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending: []string{"retry", isVPNServerRouteStatusDeleting},
		Target:  []string{isVPNServerStatusDeleted, isVPNServerRouteStatusFailed},
		Refresh: func() (interface{}, string, error) {
			getVPNServerRouteOptions := &vpcv1.GetVPNServerRouteOptions{}

			parts, err := flex.SepIdParts(d.Id(), "/")
			if err != nil {
				return nil, "", flex.DiscriminatedTerraformErrorf(err, err.Error(), "ibm_is_vpn_server_route", "delete-wait", "sep-id-parts")
			}

			getVPNServerRouteOptions.SetVPNServerID(parts[0])
			getVPNServerRouteOptions.SetID(parts[1])

			vpnServerRoute, response, err := sess.GetVPNServerRouteWithContext(context, getVPNServerRouteOptions)
			if err != nil {
				if response != nil && response.StatusCode == 404 {
					return vpnServerRoute, isVPNServerRouteStatusDeleted, nil
				}
				return vpnServerRoute, isVPNServerRouteStatusFailed, flex.TerraformErrorf(err, fmt.Sprintf("GetVPNServerRouteWithContext failed: %s", err.Error()), "ibm_is_vpn_server_route", "delete-wait")
			}
			return vpnServerRoute, *vpnServerRoute.LifecycleState, nil

		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	return stateConf.WaitForState()
}

func resourceVPNServerRouteFlattenLifecycleReasons(lifecycleReasons []vpcv1.VPNServerRouteLifecycleReason) (lifecycleReasonsList []map[string]interface{}) {
	lifecycleReasonsList = make([]map[string]interface{}, 0)
	for _, lr := range lifecycleReasons {
		currentLR := map[string]interface{}{}
		if lr.Code != nil && lr.Message != nil {
			currentLR[isInstanceLifecycleReasonsCode] = *lr.Code
			currentLR[isInstanceLifecycleReasonsMessage] = *lr.Message
			if lr.MoreInfo != nil {
				currentLR[isInstanceLifecycleReasonsMoreInfo] = *lr.MoreInfo
			}
			lifecycleReasonsList = append(lifecycleReasonsList, currentLR)
		}
	}
	return lifecycleReasonsList
}
func resourceVPNServerRouteFlattenHealthReasons(healthReasons []vpcv1.VPNServerRouteHealthReason) (healthReasonsList []map[string]interface{}) {
	healthReasonsList = make([]map[string]interface{}, 0)
	for _, hr := range healthReasons {
		currentHR := map[string]interface{}{}
		if hr.Code != nil && hr.Message != nil {
			currentHR[isVolumeHealthReasonsCode] = *hr.Code
			currentHR[isVolumeHealthReasonsMessage] = *hr.Message
			if hr.MoreInfo != nil {
				currentHR[isVolumeHealthReasonsMoreInfo] = *hr.MoreInfo
			}
			healthReasonsList = append(healthReasonsList, currentHR)
		}
	}
	return healthReasonsList
}
