package cloudfoundry

import (
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
	um := session.ClientV2

	name := d.Get("name").(string)

	users, _, err := um.GetUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Username == name {
			d.SetId(user.GUID)
			return nil
		}
	}
	return NotFound
}
