package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceService() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceServiceRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"service_plans": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceServiceRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	space := d.Get("space").(string)

	filters := []ccv2.Filter{
		ccv2.FilterEqual(constant.LabelFilter, name),
	}
	if space != "" {
		filters = append(filters, ccv2.FilterBySpace(space))
	}
	services, _, err := session.ClientV2.GetServices(filters...)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		return NotFound
	}
	service := services[0]
	d.SetId(service.GUID)

	servicePlans, _, err := session.ClientV2.GetServicePlans(ccv2.FilterEqual(constant.ServiceGUIDFilter, service.GUID))
	if err != nil {
		return err
	}

	servicePlansTf := make(map[string]interface{})
	for _, sp := range servicePlans {
		servicePlansTf[strings.Replace(sp.Name, ".", "_", -1)] = sp.GUID
	}
	d.Set("service_plans", servicePlansTf)

	return nil
}
