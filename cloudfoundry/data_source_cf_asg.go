package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAsg() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceAsgRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAsgRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	asgs, _, err := session.ClientV2.GetSecurityGroups(ccv2.FilterByName(d.Get("name").(string)))
	if err != nil {
		return err
	}
	if len(asgs) == 0 {
		return NotFound
	}
	d.SetId(asgs[0].GUID)
	return nil
}
