package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

func dataSourceSpace() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceSpaceRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org_name": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"org"},
			},
			"org": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"org_name"},
			},
			"quota": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSpaceRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	om := session.OrgManager()
	sm := session.SpaceManager()

	var (
		v  interface{}
		ok bool

		name string

		org   cfapi.CCOrg
		space cfapi.CCSpace
	)

	name = d.Get("name").(string)

	if v, ok = d.GetOk("org"); ok {
		if org, err = om.ReadOrg(v.(string)); err != nil {
			return
		}
	} else if v, ok = d.GetOk("org_name"); ok {
		if org, err = om.FindOrg(v.(string)); err != nil {
			return
		}
	}
	space, err = sm.FindSpaceInOrg(name, org.ID)
	if err != nil {
		return
	}

	d.SetId(space.ID)
	d.Set("org_name", org.Name)
	d.Set("org", org.ID)
	d.Set("quota", space.QuotaGUID)

	return
}
