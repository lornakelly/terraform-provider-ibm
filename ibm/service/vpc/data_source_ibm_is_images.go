// Copyright IBM Corp. 2017, 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package vpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/flex"
	"github.com/IBM-Cloud/terraform-provider-ibm/ibm/validate"
	"github.com/IBM/vpc-go-sdk/vpcv1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	isImages                = "images"
	isImagesResourceGroupID = "resource_group"
	isImageCatalogManaged   = "catalog_managed"
	isImageRemoteAccountId  = "remote_account_id"
)

func DataSourceIBMISImages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIBMISImagesRead,

		Schema: map[string]*schema.Schema{
			isImagesResourceGroupID: {
				Type:        schema.TypeString,
				Description: "The id of the resource group",
				Optional:    true,
			},
			isImageCatalogManaged: {
				Type:        schema.TypeBool,
				Description: "Lists images managed as part of a catalog offering. If an image is managed, accounts in the same enterprise with access to that catalog can specify the image's catalog offering version CRN to provision virtual server instances using the image",
				Optional:    true,
			},
			isImageName: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.InvokeValidator("ibm_is_image", isImageName),
				Description:  "The name of the image",
			},
			isImageStatus: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.InvokeDataSourceValidator("ibm_is_images", isImageStatus),
				Description:  "The status of the image",
			},
			isImageVisibility: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Whether the image is publicly visible or private to the account",
			},
			isImageRemoteAccountId: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filters the collection to images with a remote.account.id property matching the specified account identifier.",
			},
			isImageUserDataFormat: {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "Filters the collection to images with a user_data_format property matching one of the specified comma-separated values.",
			},

			isImages: {
				Type:        schema.TypeList,
				Description: "List of images",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Image name",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier for this image",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of this image",
						},
						"status_reasons": &schema.Schema{
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The reasons for the current status (if any).",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"code": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "A snake case string succinctly identifying the status reason.",
									},
									"message": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "An explanation of the status reason.",
									},
									"more_info": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Link to documentation about this status reason.",
									},
								},
							},
						},
						"visibility": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Whether the image is publicly visible or private to the account",
						},
						"operating_system": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									isOperatingSystemAllowUserImageCreation: {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Users may create new images with this operating system",
									},
									"architecture": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The operating system architecture",
									},
									"dedicated_host_only": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Images with this operating system can only be used on dedicated hosts or dedicated host groups",
									},
									"display_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "A unique, display-friendly name for the operating system",
									},
									"family": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The software family for this operating system",
									},
									"href": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The URL for this operating system",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The globally unique name for this operating system",
									},
									"vendor": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The vendor of the operating system",
									},
									"version": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The major release version of this operating system",
									},
									isOperatingSystemUserDataFormat: {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The user data format for this image",
									},
								},
							},
						},
						"os": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Image Operating system",
						},
						"architecture": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The operating system architecture",
						},
						"remote": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "If present, this property indicates that the resource associated with this reference is remote and therefore may not be directly retrievable.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "If present, this property indicates that the referenced resource is remote to this account, and identifies the owning account.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"id": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The unique identifier for this resource group.",
												},
												"resource_type": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The resource type.",
												},
											},
										},
									},
								},
							},
						},
						"resource_group": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The resource group for this IPsec policy.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"href": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The URL for this resource group.",
									},
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The unique identifier for this resource group.",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The user-defined name for this resource group.",
									},
								},
							},
						},
						"crn": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The CRN for this image",
						},
						isImageCheckSum: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SHA256 Checksum for this image",
						},
						isImageEncryptionKey: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The CRN of the Key Protect Root Key or Hyper Protect Crypto Service Root Key for this resource",
						},
						isImageEncryption: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of encryption used on the image",
						},
						"source_volume": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Source volume id of the image",
						},
						isImageCatalogOffering: {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									isImageCatalogOfferingManaged: {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indicates whether this image is managed as part of a catalog offering. A managed image can be provisioned using its catalog offering CRN or catalog offering version CRN.",
									},
									isImageCatalogOfferingVersion: {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The catalog offering version associated with this image.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												isImageCatalogOfferingDeleted: {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "If present, this property indicates the referenced resource has been deleted and providessome supplementary information.",
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															isImageCatalogOfferingMoreInfo: {
																Type:        schema.TypeString,
																Computed:    true,
																Description: "Link to documentation about deleted resources.",
															},
														},
													},
												},
												isImageCatalogOfferingCrn: {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The CRN for this version of the IBM Cloud catalog offering.",
												},
											},
										},
									},
								},
							},
						},
						"allowed_use": &schema.Schema{
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The usage constraints to match against the requested instance or bare metal server properties to determine compatibility.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"api_version": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The API version with which to evaluate the expressions.",
									},
									"bare_metal_server": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The expression that must be satisfied by the properties of a bare metal server provisioned using this image.",
									},
									"instance": &schema.Schema{
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The expression that must be satisfied by the properties of a virtual server instance provisioned using this image.",
									},
								},
							},
						},
						isImageUserDataFormat: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user data format for this image",
						},
						isImageAccessTags: {
							Type:        schema.TypeSet,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         flex.ResourceIBMVPCHash,
							Description: "List of access tags",
						},
					},
				},
			},
		},
	}
}

func DataSourceIBMISImagesValidator() *validate.ResourceValidator {

	status := "available, deleting, deprecated, failed, obsolete, pending, tentative, unusable"
	validateSchema := make([]validate.ValidateSchema, 0)
	validateSchema = append(validateSchema,
		validate.ValidateSchema{
			Identifier:                 isImageStatus,
			ValidateFunctionIdentifier: validate.ValidateAllowedStringValue,
			Type:                       validate.TypeString,
			Optional:                   true,
			AllowedValues:              status})
	ibmISImageResourceValidator := validate.ResourceValidator{ResourceName: "ibm_is_images", Schema: validateSchema}
	return &ibmISImageResourceValidator
}

func dataSourceIBMISImagesRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	err := imageList(context, d, meta)
	if err != nil {
		return err
	}
	return nil
}

func imageList(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sess, err := vpcClient(meta)
	if err != nil {
		tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "(Data) ibm_is_images", "read", "initialize-client")
		log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
		return tfErr.GetDiag()
	}
	start := ""
	allrecs := []vpcv1.Image{}

	var resourceGroupID string
	if v, ok := d.GetOk(isImagesResourceGroupID); ok {
		resourceGroupID = v.(string)
	}

	var imageName string
	if v, ok := d.GetOk(isImageName); ok {
		imageName = v.(string)
	}

	var visibility string
	if v, ok := d.GetOk(isImageVisibility); ok {
		visibility = v.(string)
	}

	var remoteAccountId string
	if v, ok := d.GetOk(isImageRemoteAccountId); ok {
		remoteAccountId = v.(string)
	}
	var status string
	if v, ok := d.GetOk(isImageStatus); ok {
		status = v.(string)
	}
	var catalogManaged bool
	if v, ok := d.GetOk(isImageCatalogManaged); ok {
		catalogManaged = v.(bool)
	}

	listImagesOptions := &vpcv1.ListImagesOptions{}
	if resourceGroupID != "" {
		listImagesOptions.SetResourceGroupID(resourceGroupID)
	}
	if imageName != "" {
		listImagesOptions.SetName(imageName)
	}

	if remoteAccountId != "" {
		if remoteAccountId == "user" {
			remoteAccountId = "null"
			listImagesOptions.SetRemoteAccountID(remoteAccountId)
		} else if remoteAccountId == "provider" {
			remoteAccountId = "not:null"
			listImagesOptions.SetRemoteAccountID(remoteAccountId)
		} else {
			listImagesOptions.SetRemoteAccountID(remoteAccountId)
		}
	}

	if visibility != "" {
		listImagesOptions.SetVisibility(visibility)
	}

	if userDataFormat, ok := d.GetOk(isImageUserDataFormat); ok {
		userDataFormats := userDataFormat.(*schema.Set)
		if userDataFormats.Len() != 0 {
			userDataFormatsArray := make([]string, userDataFormats.Len())
			for i, key := range userDataFormats.List() {
				userDataFormatsArray[i] = key.(string)
			}
			listImagesOptions.SetUserDataFormat(userDataFormatsArray)
		}
	}

	for {
		if start != "" {
			listImagesOptions.Start = &start
		}
		availableImages, _, err := sess.ListImagesWithContext(context, listImagesOptions)
		if err != nil {
			tfErr := flex.TerraformErrorf(err, fmt.Sprintf("ListImagesWithContext failed %s", err), "(Data) ibm_is_images", "read")
			log.Printf("[DEBUG] %s", tfErr.GetDebugMessage())
			return tfErr.GetDiag()
		}
		start = flex.GetNext(availableImages.Next)
		allrecs = append(allrecs, availableImages.Images...)
		if start == "" {
			break
		}
	}

	if status != "" {
		allrecsTemp := []vpcv1.Image{}
		for _, image := range allrecs {
			if status == *image.Status {
				allrecsTemp = append(allrecsTemp, image)
			}
		}
		allrecs = allrecsTemp
	}

	if catalogManaged {
		allrecsTemp := []vpcv1.Image{}
		for _, image := range allrecs {
			if image.CatalogOffering != nil && catalogManaged == *image.CatalogOffering.Managed {
				allrecsTemp = append(allrecsTemp, image)
			}
		}
		allrecs = allrecsTemp
	}

	imagesInfo := make([]map[string]interface{}, 0)
	for _, image := range allrecs {

		l := map[string]interface{}{
			"name":         *image.Name,
			"id":           *image.ID,
			"status":       *image.Status,
			"crn":          *image.CRN,
			"visibility":   *image.Visibility,
			"os":           *image.OperatingSystem.Name,
			"architecture": *image.OperatingSystem.Architecture,
		}
		if image.UserDataFormat != nil {
			l["user_data_format"] = *image.UserDataFormat
		}
		if len(image.StatusReasons) > 0 {
			l["status_reasons"] = dataSourceIBMIsImageFlattenStatusReasons(image.StatusReasons)
		}
		if image.ResourceGroup != nil {
			resourceGroupList := []map[string]interface{}{}
			resourceGroupMap := dataSourceImageResourceGroupToMap(*image.ResourceGroup)
			resourceGroupList = append(resourceGroupList, resourceGroupMap)
			l["resource_group"] = resourceGroupList
		}
		if image.OperatingSystem != nil {
			operatingSystemList := []map[string]interface{}{}
			operatingSystemMap := dataSourceIBMISImageOperatingSystemToMap(*image.OperatingSystem)
			operatingSystemList = append(operatingSystemList, operatingSystemMap)
			l["operating_system"] = operatingSystemList
		}
		if image.File != nil && image.File.Checksums != nil {
			l[isImageCheckSum] = *image.File.Checksums.Sha256
		}
		if image.Encryption != nil {
			l["encryption"] = *image.Encryption
		}
		if image.EncryptionKey != nil {
			l["encryption_key"] = *image.EncryptionKey.CRN
		}
		if image.SourceVolume != nil {
			l["source_volume"] = *image.SourceVolume.ID
		}
		if image.CatalogOffering != nil {
			catalogOfferingList := []map[string]interface{}{}
			catalogOfferingMap := dataSourceImageCollectionCatalogOfferingToMap(*image.CatalogOffering)
			catalogOfferingList = append(catalogOfferingList, catalogOfferingMap)
			l[isImageCatalogOffering] = catalogOfferingList
		}

		if image.Remote != nil {
			imageRemoteMap, err := dataSourceImageRemote(image)
			if err != nil {
				if err != nil {
					tfErr := flex.DiscriminatedTerraformErrorf(err, err.Error(), "(Data) ibm_is_image", "read", "initialize-client")
					log.Printf("[DEBUG]\n%s", tfErr.GetDebugMessage())
					return tfErr.GetDiag()
				}
			}
			if len(imageRemoteMap) > 0 {
				l["remote"] = []interface{}{imageRemoteMap}
			}
		}

		if image.AllowedUse != nil {
			usageConstraintList := []map[string]interface{}{}
			modelMap, err := DataSourceIBMIsImageAllowedUseToMap(image.AllowedUse)
			if err != nil {
				tfErr := flex.TerraformErrorf(err, err.Error(), "(Data) ibm_is_image", "read")
				log.Println(tfErr.GetDiag())
			}
			usageConstraintList = append(usageConstraintList, modelMap)
			l["allowed_use"] = usageConstraintList
		}

		accesstags, err := flex.GetGlobalTagsUsingCRN(meta, *image.CRN, "", isImageAccessTagType)
		if err != nil {
			log.Printf(
				"Error on get of resource image (%s) access tags: %s", d.Id(), err)
		}
		l[isImageAccessTags] = accesstags
		imagesInfo = append(imagesInfo, l)
	}
	d.SetId(dataSourceIBMISImagesID(d))
	if err = d.Set("images", imagesInfo); err != nil {
		return flex.DiscriminatedTerraformErrorf(err, fmt.Sprintf("Error setting images %s", err), "(Data) ibm_is_images", "read", "images-set").GetDiag()
	}
	return nil
}

// dataSourceIBMISImagesId returns a reasonable ID for a image list.
func dataSourceIBMISImagesID(d *schema.ResourceData) string {
	return time.Now().UTC().String()
}
