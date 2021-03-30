package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func dataSourceUserProvidedService() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceUserProvidedServiceRead,

		Schema: map[string]*schema.Schema{
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"credentials": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
			"route_service_url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"syslog_drain_url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceUserProvidedServiceRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	var (
		name            string
		space           string
		serviceInstance ccv2.UserProvidedServiceInstance
	)

	name = d.Get("name").(string)
	space = d.Get("space").(string)
	serviceInstances, _, err := session.ClientV2.GetUserProvServiceInstances(ccv2.FilterByName(name), ccv2.FilterEqual(constant.SpaceGUIDFilter, space))
	if err != nil {
		return err
	}
	if len(serviceInstances) == 0 {
		return NotFound
	}
	serviceInstance = serviceInstances[0]

	d.SetId(serviceInstance.GUID)
	d.Set("name", serviceInstance.Name)
	d.Set("credentials", serviceInstance.Credentials)
	d.Set("route_service_url", serviceInstance.RouteServiceUrl)
	d.Set("syslog_drain_url", serviceInstance.SyslogDrainUrl)
	d.Set("tags", serviceInstance.Tags)

	return nil
}
