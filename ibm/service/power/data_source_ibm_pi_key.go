// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package power

import (
	"context"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/helpers"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/conns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DataSourceIBMPIKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMPIKeyRead,
		Schema: map[string]*schema.Schema{
			// Arguments
			Arg_CloudInstanceID: {
				Description:  "The GUID of the service instance associated with an account.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},
			Arg_KeyName: {
				Description:  "User defined name for the SSH key.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.NoZeroValues,
			},

			// Attributes
			Attr_CreationDate: {
				Computed:    true,
				Description: "Date of SSH Key creation.",
				Type:        schema.TypeString,
			},
			Attr_Description: {
				Computed:    true,
				Description: "Description of the ssh key.",
				Type:        schema.TypeString,
			},
			Attr_KeyName: {
				Computed:    true,
				Description: "Name of SSH key.",
				Type:        schema.TypeString,
			},
			Attr_PrimaryWorkspace: {
				Computed:    true,
				Description: "Indicates if the current workspace owns the ssh key or not.",
				Type:        schema.TypeBool,
			},
			Attr_SSHKey: {
				Computed:    true,
				Description: "SSH RSA key.",
				Sensitive:   true,
				Type:        schema.TypeString,
			},
			Attr_SSHKeyID: {
				Computed:    true,
				Description: "Unique ID of SSH key.",
				Type:        schema.TypeString,
			},
			Attr_Visibility: {
				Computed:    true,
				Description: "Visibility of the ssh key.",
				Type:        schema.TypeString,
			},
		},
	}
}

func dataSourceIBMPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := meta.(conns.ClientSession).IBMPISession()
	if err != nil {
		return diag.FromErr(err)
	}

	cloudInstanceID := d.Get(Arg_CloudInstanceID).(string)

	sshkeyC := instance.NewIBMPISSHKeyClient(ctx, sess, cloudInstanceID)
	sshkeydata, err := sshkeyC.Get(d.Get(helpers.PIKeyName).(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*sshkeydata.Name)
	d.Set(Attr_CreationDate, sshkeydata.CreationDate.String())
	d.Set(Attr_Description, sshkeydata.Description)
	d.Set(Attr_KeyName, sshkeydata.Name)
	d.Set(Attr_PrimaryWorkspace, sshkeydata.PrimaryWorkspace)
	d.Set(Attr_SSHKey, sshkeydata.SSHKey)
	d.Set(Attr_SSHKeyID, sshkeydata.ID)
	d.Set(Attr_Visibility, sshkeydata.Visibility)

	return nil
}
