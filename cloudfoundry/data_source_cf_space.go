package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSpace() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceSpaceRead,

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

func dataSourceSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	name := d.Get("name").(string)

	if d.Get("org").(string) == "" && d.Get("org_name").(string) == "" {
		return diag.Errorf("You must provide either 'org' or 'org_name' attribute")
	}

	orgId := d.Get("org").(string)
	orgName := d.Get("org_name").(string)
	if d.Get("org_name").(string) != "" {
		orgs, _, err := session.ClientV3.GetOrganizations(ccv3.Query{
			Key:    "names",
			Values: []string{orgName},
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if len(orgs) == 0 {
			return diag.Errorf("Can't found org with name %s", orgName)
		}
		orgId = orgs[0].GUID
	} else {
		org, _, err := session.ClientV3.GetOrganization(orgId)
		if err != nil {
			return diag.FromErr(err)
		}
		orgName = org.Name
	}
	nameQuery := ccv3.Query{
		Key:    ccv3.NameFilter,
		Values: []string{name},
	}
	orgQuery := ccv3.Query{
		Key:    ccv3.OrganizationGUIDFilter,
		Values: []string{orgId},
	}
	spaces, _, _, err := session.ClientV3.GetSpaces(nameQuery, orgQuery)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(spaces) == 0 {
		return diag.FromErr(NotFound)
	}
	space := spaces[0]
	d.SetId(space.GUID)
	d.Set("org_name", orgName)
	d.Set("org", orgId)
	d.Set("quota", space.Relationships["quota"].GUID)

	err = metadataRead(spaceMetadata, d, meta, true)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(err)
}
