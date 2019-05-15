package cloudfoundry

import (
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceUser() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceUserRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	um := session.ClientUAA

	name := d.Get("name").(string)

	users, err := um.GetUsersByUsername(name)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return NotFound
	}
	d.SetId(users[0].ID)
	return err
}
