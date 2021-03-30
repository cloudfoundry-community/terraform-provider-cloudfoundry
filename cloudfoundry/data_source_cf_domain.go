package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomain() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceDomainRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"sub_domain": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
			"domain": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"internal": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceDomainRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	dm := session.ClientV2

	var (
		name, prefix string
	)

	sharedDomains, _, err := dm.GetSharedDomains()
	if err != nil {
		return err
	}
	privateDomains, _, err := dm.GetPrivateDomains()
	if err != nil {
		return err
	}
	domains := append(sharedDomains, privateDomains...)

	if v, ok := d.GetOk("sub_domain"); ok {
		prefix = v.(string) + "."
		if v, ok = d.GetOk("domain"); ok {
			name = prefix + v.(string)
		}
	} else if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		return fmt.Errorf("neither a full name or sub-domain was provided to do an effective domain search")
	}

	var domain *ccv2.Domain
	if len(name) == 0 {
		for _, d := range domains {
			if strings.HasPrefix(d.Name, prefix) {
				domain = &d
				break
			}
		}
		if domain == nil {
			return fmt.Errorf("no domain found with sub-domain '%s'", prefix)
		}
	} else {
		for _, d := range domains {
			if name == d.Name {
				domain = &d
				break
			}
		}
		if domain == nil {
			return fmt.Errorf("no domain found with name '%s'", name)
		}
	}

	domainParts := strings.Split(domain.Name, ".")

	d.Set("name", domain.Name)
	d.Set("sub_domain", domainParts[0])
	d.Set("domain", strings.Join(domainParts[1:], "."))
	d.Set("org", domain.OwningOrganizationGUID)
	d.Set("internal", domain.Internal)
	d.SetId(domain.GUID)
	return err
}
