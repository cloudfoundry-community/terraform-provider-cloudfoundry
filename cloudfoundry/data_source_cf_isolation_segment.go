package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceIsolationSegment() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceIsolationSegmentRead,

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

func dataSourceIsolationSegmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	segments, _, err := session.ClientV3.GetIsolationSegments(ccv3.Query{
		Key:    ccv3.NameFilter,
		Values: []string{d.Get("name").(string)},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if len(segments) == 0 {
		return diag.FromErr(NotFound)
	}
	d.SetId(segments[0].GUID)
	err = metadataRead(segmentMetadata, d, meta, true)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
