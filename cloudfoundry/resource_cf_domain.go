package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDomain() *schema.Resource {

	return &schema.Resource{

		Create: resourceDomainCreate,
		Read:   resourceDomainRead,
		Delete: resourceDomainDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceDomainRead),
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"sub_domain": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"internal": &schema.Schema{
				Type:        schema.TypeBool,
				ForceNew:    true,
				Optional:    true,
				Description: "Flag that sets the domain as an internal domain. Internal domains are used for internal app to app networking only.",
				Default:     false,
			},
			"router_group": &schema.Schema{
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"org"},
			},
			"router_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"org": &schema.Schema{
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"router_group"},
			},
			// "shared-with": &schema.Schema{
			// 	Type:     schema.TypeSet,
			// 	Optional: true,
			// 	Elem:     &schema.Schema{Type: schema.TypeString},
			// 	Set:      resourceStringHash,
			// },
		},
	}
}

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)

	nameAttr, nameOk := d.GetOk("name")
	subDomainAttr, subDomainOk := d.GetOk("sub_domain")
	domainAttr, domainOk := d.GetOk("domain")
	org, orgOk := d.GetOk("org")
	routerGroup := d.Get("router_group")

	if nameOk {

		domainParts := strings.Split(nameAttr.(string), ".")
		if len(domainParts) <= 1 {
			return fmt.Errorf("the 'name' attribute does not contain a sub-domain")
		}
		sd := domainParts[0]
		dn := strings.Join(domainParts[1:], ".")

		if subDomainOk {
			return fmt.Errorf("the 'sub_domain' will be computed from the 'name' attribute, so it is not needed here")
		}
		if domainOk {
			return fmt.Errorf("the 'domain' will be computed from the 'name' attribute, so it is not needed here")
		}
		d.Set("sub_domain", sd)
		d.Set("domain", dn)
	} else {
		if !subDomainOk || !domainOk {
			return fmt.Errorf("to compute the 'name' both the 'sub_domain' and 'domain' attributes need to be provided")
		}
		d.Set("name", subDomainAttr.(string)+"."+domainAttr.(string))
	}

	var (
		ccDomain ccv2.Domain
		err      error
	)
	name := d.Get("name").(string)

	dm := session.ClientV2
	if orgOk {
		ccDomain, _, err = dm.CreatePrivateDomain(name, org.(string))
	} else {
		ccDomain, _, err = dm.CreateSharedDomain(name, routerGroup.(string), d.Get("internal").(bool))
	}
	if err != nil {
		return err
	}
	d.Set("router_type", ccDomain.RouterGroupType)
	d.SetId(ccDomain.GUID)
	return nil
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	dm := session.ClientV2
	id := d.Id()

	ccDomain, _, err := dm.GetSharedDomain(id)
	if err != nil {
		ccDomain, _, err = dm.GetPrivateDomain(id)
	}
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	domainParts := strings.Split(ccDomain.Name, ".")
	subDomain := domainParts[0]
	domain := strings.Join(domainParts[1:], ".")

	d.Set("name", ccDomain.Name)
	d.Set("sub_domain", subDomain)
	d.Set("domain", domain)
	d.Set("route_group", ccDomain.RouterGroupGUID)
	d.Set("router_type", ccDomain.RouterGroupType)
	d.Set("internal", ccDomain.Internal)
	d.Set("org", ccDomain.OwningOrganizationGUID)

	return nil
}

func resourceDomainDelete(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	dm := session.ClientV2
	id := d.Id()

	var err error
	if _, orgOk := d.GetOk("org"); orgOk {
		_, err = dm.DeletePrivateDomain(id)
	} else {
		_, err = dm.DeleteSharedDomain(id)
	}
	return err
}
