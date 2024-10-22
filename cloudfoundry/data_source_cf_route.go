package cloudfoundry

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoute() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceRouteRead,

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

func dataSourceRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	queries := []ccv3.Query{{
		Key:    "domain_guids",
		Values: []string{d.Get("domain").(string)},
	}}

	if hostname, ok := d.GetOk("hostname"); ok {
		queries = append(queries,
			ccv3.Query{
				Key:    ccv3.HostsFilter,
				Values: []string{hostname.(string)},
			},
		)
	}
	if org, ok := d.GetOk("org"); ok {
		queries = append(queries,
			ccv3.Query{
				Key:    ccv3.OrganizationGUIDFilter,
				Values: []string{org.(string)},
			},
		)
	}
	if path, ok := d.GetOk("path"); ok {
		queries = append(queries,
			ccv3.Query{
				Key:    "paths",
				Values: []string{path.(string)},
			},
		)
	}
	if port, ok := d.GetOk("port"); ok {
		queries = append(queries,
			ccv3.Query{
				Key:    "ports",
				Values: []string{port.(string)},
			},
		)
	}

	routes, _, err := session.ClientV3.GetRoutes(queries...)

	if err != nil {
		return diag.FromErr(err)
	}
	if len(routes) == 0 {
		return diag.FromErr(NotFound)
	}
	if len(routes) > 1 {
		return diag.FromErr(fmt.Errorf("Unexpected error reading route (more than 1 match)"))
	}

	route := routes[0]

	d.Set("hostname", route.Host)
	d.Set("path", route.Path)
	if route.Port == 0 {
		d.Set("port", route.Port)
	}
	d.SetId(route.GUID)
	return diag.FromErr(err)
}
