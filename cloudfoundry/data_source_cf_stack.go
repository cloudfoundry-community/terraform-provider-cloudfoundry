package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
		},
	}
}

func dataSourceStackRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	sm := session.ClientV2
	name := d.Get("name").(string)

	stacks, _, err := sm.GetStacks(ccv2.FilterByName(name))
	if err != nil {
		return err
	}
	if len(stacks) == 0 {
		return NotFound
	}
	d.SetId(stacks[0].GUID)
	d.Set("description", stacks[0].Description)
	err = metadataRead(stackMetadata, d, meta, true)
	if err != nil {
		return err
	}
	return err
}
