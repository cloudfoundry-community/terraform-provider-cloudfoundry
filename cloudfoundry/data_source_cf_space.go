package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
		},
	}
}

func dataSourceSpaceRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	name := d.Get("name").(string)

	if d.Get("org").(string) == "" && d.Get("org_name").(string) == "" {
		return fmt.Errorf("You must provide either 'org' or 'org_name' attribute")
	}

	orgId := d.Get("org").(string)
	orgName := d.Get("org_name").(string)
	if d.Get("org_name").(string) != "" {
		orgs, _, err := session.ClientV2.GetOrganizations(ccv2.FilterByName(orgName))
		if err != nil {
			return err
		}
		if len(orgs) == 0 {
			return fmt.Errorf("Can't found org with name %s", orgName)
		}
		orgId = orgs[0].GUID
	} else {
		org, _, err := session.ClientV2.GetOrganization(orgId)
		if err != nil {
			return err
		}
		orgName = org.Name
	}
	spaces, _, err := session.ClientV2.GetSpaces(ccv2.FilterByName(name), ccv2.FilterByOrg(orgId))
	if err != nil {
		return err
	}
	if len(spaces) == 0 {
		return NotFound
	}
	space := spaces[0]
	d.SetId(space.GUID)
	d.Set("org_name", orgName)
	d.Set("org", orgId)
	d.Set("quota", space.SpaceQuotaDefinitionGUID)

	err = metadataRead(spaceMetadata, d, meta, true)
	if err != nil {
		return err
	}
	return err
}
