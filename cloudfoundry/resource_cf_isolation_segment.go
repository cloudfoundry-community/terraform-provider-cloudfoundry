package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

func resourceSegment() *schema.Resource {
	return &schema.Resource{
		Create: resourceSegmentCreate,
		Read:   resourceSegmentRead,
		Update: resourceSegmentUpdate,
		Delete: resourceSegmentDelete,
		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceSegmentEntitlement() *schema.Resource {
	return &schema.Resource{
		Create: resourceSegmentEntitlementCreate,
		Read:   resourceSegmentEntitlementRead,
		Update: resourceSegmentEntitlementUpdate,
		Delete: resourceSegmentEntitlementDelete,
		// Importer: &schema.ResourceImporter{
		// 	State: ImportStatePassthrough,
		// },

		Schema: map[string]*schema.Schema{
			"segment_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
				MinItems: 1,
				Required: true,
			},
		},
	}
}

func resourceSegmentCreate(d *schema.ResourceData, meta interface{}) error {
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
	return nil
}

func resourceSegmentRead(d *schema.ResourceData, meta interface{}) error {
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
	d.Set("name", seg.Name)
	return nil
}

func resourceSegmentUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	name := d.Get("name").(string)
	sm := session.SegmentManager()
	if name == "" {
		return fmt.Errorf(" segment name cannot be empty")
	}

	_, err := sm.UpdateSegment(d.Id(), name)
	if err != nil {
		d.SetId("")
		return err
	}
	return nil
}

func resourceSegmentDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	sm := session.SegmentManager()
	return sm.DeleteSegment(d.Id())
}

func resourceSegmentEntitlementCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	name := d.Get("segment_id").(string)
	orgIDs := d.Get("org_ids").(*schema.Set).List()
	d.SetId(name)

	sm := session.SegmentManager()
	if err := sm.SetSegmentOrgs(d.Id(), orgIDs); err != nil {
		d.SetId("")
		return err
	}
	return nil
}

func resourceSegmentEntitlementUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	sm := session.SegmentManager()
	if d.HasChange("org_ids") {
		orgIDs := d.Get("org_ids").(*schema.Set).List()
		if err := sm.SetSegmentOrgs(d.Id(), orgIDs); err != nil {
			return err
		}
	}
	return nil
}

func resourceSegmentEntitlementRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	sm := session.SegmentManager()
	orgIDs, err := sm.GetSegmentOrgs(d.Id())
	if err != nil {
		d.SetId("")
		return err
	}
	d.Set("org_ids", schema.NewSet(resourceStringHash, orgIDs))
	return nil
}

func resourceSegmentEntitlementDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	sm := session.SegmentManager()
	orgIDs := d.Get("org_ids").(*schema.Set).List()
	return sm.DeleteSegmentOrgs(d.Id(), orgIDs)
}
