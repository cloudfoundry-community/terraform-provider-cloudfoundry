package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrgV3() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceOrgReadV3,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
		},
	}
}

func dataSourceOrgReadV3(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	name := d.Get("name").(string)

	orgs, _, err := session.ClientV3.GetOrganizations(ccv3.Query{
		Key:    "names",
		Values: []string{name},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if len(orgs) == 0 {
		return diag.FromErr(NotFound)
	}
	d.SetId(orgs[0].GUID)

	err = metadataRead(orgMetadata, d, meta, true)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(err)
}
