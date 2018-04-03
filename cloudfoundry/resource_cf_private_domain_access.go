package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourcePrivateDomainAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourcePrivateDomainAccessCreate,
		Read:   resourcePrivateDomainAccessRead,
		Delete: resourcePrivateDomainAccessDelete,
		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

// PrivateDomainAccessImport -
// Checks that user-given ID matches <guid>/<guid> format
func PrivateDomainAccessImport(d *schema.ResourceData, meta interface{}) (res []*schema.ResourceData, err error) {
	// session := meta.(*cfapi.Session)
	// if session == nil {
	// 	err = fmt.Errorf("client is nil")
	// 	return
	// }
	// dm := session.DomainManager()
	id := d.Id()

	// var org, domain string
	// if org, domain, err = parseID(id); err != nil {
	if _, _, err = parseID(id); err != nil {
		return
	}

	// var found bool
	// found, err = dm.HasPrivateDomainAccess(org, domain)
	// if err != nil {
	// 	return
	// }

	// if !found {
	// 	err = fmt.Errorf("organization '%s' has no access to private domain '%s'", org, domain)
	// 	return
	// }
	return schema.ImportStatePassthrough(d, meta)
}

func resourcePrivateDomainAccessCreate(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	domain := d.Get("domain").(string)
	org := d.Get("org").(string)

	dm := session.DomainManager()
	if err = dm.CreatePrivateDomainAccess(org, domain); err != nil {
		return
	}

	d.SetId(computeID(org, domain))
	return nil
}

func resourcePrivateDomainAccessRead(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id := d.Id()
	// id in read hook comes from create or import callback which ensure id's validity
	var org, domain string
	org, domain, _ = parseID(id)

	dm := session.DomainManager()
	var found bool
	if found, err = dm.HasPrivateDomainAccess(org, domain); err != nil || !found {
		d.SetId("")
		return err
	}

	d.Set("org", org)
	d.Set("domain", domain)
	return nil
}

func resourcePrivateDomainAccessDelete(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	dm := session.DomainManager()
	id := d.Id()

	// id in read hook comes from create or import callback which ensure id's validity
	var org, domain string
	org, domain, _ = parseID(id)

	return dm.DeletePrivateDomainAccess(org, domain)
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
