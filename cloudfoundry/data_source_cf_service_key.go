package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func dataSourceServiceKey() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceServiceKeyRead,

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
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceKeyRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	serviceKeys, _, err := session.ClientV2.GetServiceKeys(
		ccv2.FilterByName(d.Get("name").(string)),
		ccv2.FilterEqual(constant.ServiceInstanceGUIDFilter, d.Get("service_instance").(string)),
	)
	if err != nil {
		return err
	}
	if len(serviceKeys) == 0 {
		return NotFound
	}
	serviceKey := serviceKeys[0]
	d.SetId(serviceKey.GUID)
	d.Set("name", serviceKey.Name)
	d.Set("service_instance", serviceKey.ServiceInstanceGUID)
	d.Set("credentials", normalizeMap(serviceKey.Credentials, make(map[string]interface{}), "", "_"))

	return nil
}
