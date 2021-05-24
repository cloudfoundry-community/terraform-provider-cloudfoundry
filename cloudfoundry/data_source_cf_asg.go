package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAsg() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceAsgRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAsgRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}
	asgs, _, err := session.ClientV2.GetSecurityGroups(ccv2.FilterByName(d.Get("name").(string)))
	if err != nil {
		return diag.FromErr(err)
	}
	if len(asgs) == 0 {
		return diag.FromErr(NotFound)
	}
	d.SetId(asgs[0].GUID)
	return nil
}
