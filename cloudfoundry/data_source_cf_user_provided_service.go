package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func dataSourceUserProvidedService() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceUserProvidedServiceRead,

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

func dataSourceUserProvidedServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	var (
		name            string
		space           string
		serviceInstance resources.ServiceInstance
		credentials     types.JSONObject
	)

	name = d.Get("name").(string)
	space = d.Get("space").(string)
	serviceInstanceV3, _, _, err := session.ClientV3.GetServiceInstanceByNameAndSpace(name, space)

	if err != nil {
		return diag.FromErr(err)
	}

	serviceInstance = serviceInstanceV3
	credentials, _, err = session.ClientV3.GetUserProvidedServiceInstanceCredentails(serviceInstanceV3.GUID)
	
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serviceInstance.GUID)
	d.Set("name", serviceInstance.Name)
	d.Set("credentials", credentials)
	if serviceInstanceV3.RouteServiceURL.IsSet {
		d.Set("route_service_url", serviceInstanceV3.RouteServiceURL.Value)
	} else {
		d.Set("route_service_url", "")
	}
	if serviceInstanceV3.SyslogDrainURL.IsSet {
		d.Set("syslog_drain_url", serviceInstanceV3.SyslogDrainURL.Value)
	} else {
		d.Set("syslog_drain_url", "")
	}
	if serviceInstanceV3.Tags.IsSet {
		tags := make([]interface{}, len(serviceInstanceV3.Tags.Value))
		for i, v := range serviceInstanceV3.Tags.Value {
			tags[i] = v
		}
		d.Set("tags", tags)
	} else {
		d.Set("tags", nil)
	}

	return nil
}
