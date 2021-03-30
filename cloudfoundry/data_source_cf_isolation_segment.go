package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceIsolationSegment() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceIsolationSegmentRead,

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

func dataSourceIsolationSegmentRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	segments, _, err := session.ClientV3.GetIsolationSegments(ccv3.Query{
		Key:    ccv3.NameFilter,
		Values: []string{d.Get("name").(string)},
	})
	if err != nil {
		return err
	}
	if len(segments) == 0 {
		return NotFound
	}
	d.SetId(segments[0].GUID)
	err = metadataRead(segmentMetadata, d, meta, true)
	if err != nil {
		return err
	}
	return nil
}
