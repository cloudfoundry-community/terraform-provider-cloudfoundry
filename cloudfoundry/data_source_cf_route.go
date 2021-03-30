package cloudfoundry

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoute() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceRouteRead,

		Schema: map[string]*schema.Schema{
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"port": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceRouteRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	dm := session.ClientV2

	filters := []ccv2.Filter{ccv2.FilterEqual(constant.DomainGUIDFilter, d.Get("domain").(string))}
	if v, ok := d.GetOk("hostname"); ok {
		filters = append(filters, ccv2.FilterEqual(constant.HostFilter, v.(string)))
	}
	if v, ok := d.GetOk("org"); ok {
		filters = append(filters, ccv2.FilterByOrg(v.(string)))
	}
	if v, ok := d.GetOk("path"); ok {
		filters = append(filters, ccv2.FilterEqual(constant.PathFilter, v.(string)))
	}
	if v, ok := d.GetOk("port"); ok {
		filters = append(filters, ccv2.FilterEqual(constant.PortFilter, fmt.Sprintf("%d", v.(int))))
	}
	routes, _, err := dm.GetRoutes(filters...)
	if err != nil {
		return err
	}
	if len(routes) == 0 {
		return NotFound
	}
	route := routes[0]

	d.Set("hostname", route.Host)
	d.Set("path", route.Path)
	if route.Port.IsSet {
		d.Set("port", route.Port.Value)
	}
	d.SetId(route.GUID)
	return err
}
