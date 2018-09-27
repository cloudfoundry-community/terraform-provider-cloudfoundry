package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

func dataSourceSegment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSegmentRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceSegmentRead(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	sm := session.SegmentManager()
	name := d.Get("name").(string)
	seg, err := sm.FindSegment(name)
	if err != nil {
		return err
	}

	d.SetId(seg.GUID)
	return err
}
