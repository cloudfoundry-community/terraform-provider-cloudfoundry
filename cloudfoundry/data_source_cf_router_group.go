package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/router/routererror"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRouterGroup() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceRouterGroupRead,

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

func dataSourceRouterGroupRead(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	dm := session.RouterClient
	name := d.Get("name").(string)

	routerGroup, err := dm.GetRouterGroupByName(name)
	if err != nil {
		if err == (routererror.ResourceNotFoundError{}) {
			return NotFound
		}
		return err
	}
	d.SetId(routerGroup.GUID)
	d.Set("type", routerGroup.Type)
	return nil
}
