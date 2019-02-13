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
			"segment": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"orgs": &schema.Schema{
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
	name := d.Get("segment").(string)
	orgs := d.Get("orgs").(*schema.Set).List()
	d.SetId(name)

	sm := session.SegmentManager()
	if err := sm.SetSegmentOrgs(d.Id(), orgs); err != nil {
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
	if d.HasChange("orgs") {
		orgs := d.Get("orgs").(*schema.Set).List()
		if err := sm.SetSegmentOrgs(d.Id(), orgs); err != nil {
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
	orgs, err := sm.GetSegmentOrgs(d.Id())
	if err != nil {
		d.SetId("")
		return err
	}

	resourceOrgs := d.Get("orgs").(*schema.Set).List()
	finalOrgs := []interface{}{}
	for _, currentOrg := range orgs {
		found := false
		for _, requestOrg := range resourceOrgs {
			if requestOrg == currentOrg.(string) {
				found = true
				break
			}
		}
		if found {
			finalOrgs = append(finalOrgs, currentOrg)
		}
	}

	d.Set("orgs", schema.NewSet(resourceStringHash, finalOrgs))
	return nil
}

func resourceSegmentEntitlementDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	sm := session.SegmentManager()
	orgs := d.Get("orgs").(*schema.Set).List()
	return sm.DeleteSegmentOrgs(d.Id(), orgs)
}
