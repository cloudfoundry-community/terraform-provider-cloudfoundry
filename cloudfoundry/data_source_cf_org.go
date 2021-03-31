package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrg() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceOrgRead,

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

func dataSourceOrgRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	name := d.Get("name").(string)

	orgs, _, err := session.ClientV2.GetOrganizations(ccv2.FilterByName(name))
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
