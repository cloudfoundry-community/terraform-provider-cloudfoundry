package cloudfoundry

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"code.cloudfoundry.org/cli/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomain() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceDomainRead,

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

func dataSourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	dm := session.ClientV3

	var (
		name, prefix string
	)

	/* 	sharedDomains, _, err := dm.GetSharedDomains()
	   	if err != nil {
	   		return diag.FromErr(err)
	   	}
	   	privateDomains, _, err := dm.GetPrivateDomains()
	   	if err != nil {
	   		return diag.FromErr(err)
	   	} */

	domains, _, err := dm.GetDomains()

	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.GetOk("sub_domain"); ok {
		prefix = v.(string) + "."
		if v, ok = d.GetOk("domain"); ok {
			name = prefix + v.(string)
		}
	} else if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		return diag.Errorf("neither a full name or sub-domain was provided to do an effective domain search")
	}

	var domain *resources.Domain
	if len(name) == 0 {
		for _, d := range domains {
			if strings.HasPrefix(d.Name, prefix) {
				domain = &d
				break
			}
		}
		if domain == nil {
			return diag.Errorf("no domain found with sub-domain '%s'", prefix)
		}
	} else {
		for _, d := range domains {
			if name == d.Name {
				domain = &d
				break
			}
		}
		if domain == nil {
			return diag.Errorf("no domain found with name '%s'", name)
		}
	}

	domainParts := strings.Split(domain.Name, ".")

	d.Set("name", domain.Name)
	d.Set("sub_domain", domainParts[0])
	d.Set("domain", strings.Join(domainParts[1:], "."))
	d.Set("org", domain.OrganizationGUID)
	d.Set("internal", domain.Internal)
	d.SetId(domain.GUID)
	return diag.FromErr(err)
}
