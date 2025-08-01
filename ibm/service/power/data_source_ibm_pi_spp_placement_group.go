// Copyright IBM Corp. 2022 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"
	"log"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceIBMPISPPPlacementGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPISPPPlacementGroupRead,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_SPPPlacementGroupID: {
				Description:  "The ID of the shared processor pool placement group.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_CRN: {
				Computed:    true,
				Description: "The CRN of this resource.",
				Type:        schema.TypeString,
			},
			Attr_Members: {
				Computed:    true,
				Description: "List of shared processor pool IDs that are members of the placement group.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Type:        schema.TypeList,
			},
			Attr_Name: {
				Computed:    true,
				Description: "The name of the shared processor pool placement group.",
				Type:        schema.TypeString,
			},
			Attr_Policy: {
				Computed:    true,
				Description: "The value of the group's affinity policy. Valid values are affinity and anti-affinity.",
				Type:        schema.TypeString,
			},
			Attr_UserTags: {
				Computed:    true,
				Description: "List of user tags attached to the resource.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Type:        schema.TypeSet,
			},
		},
	}
}

func dataSourceIBMPISPPPlacementGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)
	placementGroupID := d.Get(Arg_SPPPlacementGroupID).(string)
	client := instance.NewIBMPISPPPlacementGroupClient(ctx, sess, cloudInstanceID)

	response, err := client.Get(placementGroupID)
	if err != nil || response == nil {
		return diag.Errorf("error fetching the spp placement group: %v", err)
	}

	d.SetId(*response.ID)
	if response.Crn != "" {
		d.Set(Attr_CRN, response.Crn)
		userTags, err := flex.GetGlobalTagsUsingCRN(meta, string(response.Crn), "", UserTagType)
		if err != nil {
			log.Printf("Error on get of spp placement group (%s) user_tags: %s", *response.ID, err)
		}
		d.Set(Attr_UserTags, userTags)
	}
	d.Set(Attr_Members, response.MemberSharedProcessorPools)
	d.Set(Attr_Name, response.Name)
	d.Set(Attr_Policy, response.Policy)

	return nil
}
