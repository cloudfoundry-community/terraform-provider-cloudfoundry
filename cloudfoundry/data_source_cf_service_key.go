package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func dataSourceServiceKey() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceServiceKeyRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"service_instance": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"credentials": &schema.Schema{
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceServiceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	var serviceKeys []resources.ServiceCredentialBinding
	var err error

	serviceKeys, _, err = session.ClientV3.GetServiceCredentialBindings(
		ccv3.Query{
			Key:    ccv3.QueryKey("service_instance_guids"),
			Values: []string{d.Get("service_instance").(string)},
		}, ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{d.Get("name").(string)},
		},
	)

	if err != nil {
		return diag.FromErr(err)
	}
	if len(serviceKeys) == 0 {
		return diag.FromErr(NotFound)
	}
	serviceKey := serviceKeys[0]
	serviceKeyDetails, _, err := session.ClientV3.GetServiceCredentialBindingDetails(serviceKey.GUID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serviceKey.GUID)
	d.Set("name", serviceKey.Name)
	d.Set("service_instance", serviceKey.ServiceInstanceGUID)
	d.Set("credentials", normalizeMap(serviceKeyDetails.Credentials, make(map[string]interface{}), "", "_"))

	return nil
}
