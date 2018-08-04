package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceIsolationSegment() *schema.Resource {
	return &schema.Resource{
		Create: resourceIsolationSegmentCreate,
		Read:   resourceIsolationSegmentRead,
		Update: resourceIsolationSegmentUpdate,
		Delete: resourceIsolationSegmentDelete,
		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"orgs": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
				Optional: true,
			},
		},
	}
}

func resourceIsolationSegmentCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	name := d.Get("name").(string)

	sm := session.SegmentManager()
	seg, err := sm.CreateSegment(name)
	if err != nil {
		return err
	}
	session.Log.DebugMessage("New segment created: %# v", seg)
	d.SetId(seg.GUID)
	orgs := d.Get("orgs").(*schema.Set).List()
	return sm.SetSegmentOrgs(d.Id(), orgs)
}

func resourceIsolationSegmentRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	sm := session.SegmentManager()
	seg, err := sm.ReadSegment(d.Id())
	if err != nil {
		d.SetId("")
		return err
	}

	orgs, err := sm.GetSegmentOrgs(d.Id())
	if err != nil {
		return err
	}
	d.Set("name", seg.Name)
	d.Set("orgs", schema.NewSet(resourceStringHash, orgs))

	return nil
}

func resourceIsolationSegmentUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	name := d.Get("name").(string)
	sm := session.SegmentManager()
	if name == "" {
		return fmt.Errorf("isolation segment name cannot be empty")
	}

	_, err := sm.UpdateSegment(d.Id(), name)
	if err != nil {
		d.SetId("")
		return err
	}

	if d.HasChange("orgs") {
		orgs := d.Get("orgs").(*schema.Set).List()
		if err = sm.SetSegmentOrgs(d.Id(), orgs); err != nil {
			return err
		}
	}

	return nil
}

func resourceIsolationSegmentDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	sm := session.SegmentManager()
	return sm.DeleteSegment(d.Id())
}
