package cloudfoundry

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/schema"
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

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	dm := session.DomainManager()
	name := d.Get("name").(string)

	routerGroup, err := dm.FindRouterGroupByName(name)
	if err != nil {
		return
	}
	d.SetId(routerGroup.GUID)
	d.Set("type", routerGroup.Type)
	return
}
