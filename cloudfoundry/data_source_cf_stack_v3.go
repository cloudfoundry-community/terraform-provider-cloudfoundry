package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStackV3() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceStackRead,

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

func dataSourceStackV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	sm := session.ClientV3

	query := ccv3.Query{
		Key:    ccv3.NameFilter,
		Values: []string{d.Get("name").(string)},
	}

	stacks, _, err := sm.GetStacks(query)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(stacks) == 0 {
		return diag.FromErr(NotFound)
	}
	d.SetId(stacks[0].GUID)
	d.Set("description", stacks[0].Description)
	err = metadataRead(stackMetadata, d, meta, true)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(err)
}
