package cloudfoundry

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceStack() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceStackRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceStackRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	sm := session.StackManager()

	var (
		name  string
		stack cfapi.CCStack
	)

	name = d.Get("name").(string)

	stack, err = sm.FindStackByName(name)
	if err != nil {
		return
	}

	d.SetId(stack.ID)
	d.Set("description", stack.Description)
	return
}
