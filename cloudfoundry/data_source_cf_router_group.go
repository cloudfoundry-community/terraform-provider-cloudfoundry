package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/router/routererror"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRouterGroup() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceRouterGroupRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRouterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	dm := session.RouterClient
	name := d.Get("name").(string)

	routerGroup, err := dm.GetRouterGroupByName(name)
	if err != nil {
		if err == (routererror.ResourceNotFoundError{}) {
			return diag.FromErr(NotFound)
		}
		return diag.FromErr(err)
	}
	d.SetId(routerGroup.GUID)
	d.Set("type", routerGroup.Type)
	return nil
}
