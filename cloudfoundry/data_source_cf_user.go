package cloudfoundry

import (
	"context"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceUserRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Optionally scope the lookup to organization",
				Optional:    true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	um := session.ClientV2

	name := strings.ToLower(d.Get("name").(string))

	users, _, err := um.GetUsers()
	isNotAuthorized := IsErrNotAuthorized(err)
	if err != nil && !isNotAuthorized {
		return diag.FromErr(err)
	}
	if isNotAuthorized { // Fallback for OrgManagers
		orgID := d.Get("org_id").(string)
		if orgID == "" {
			return diag.FromErr(err)
		}
		users, _, err := um.GetOrganizationUsers(orgID)
		if err != nil {
			return diag.FromErr(err)
		}
		if isInSlice(users, func(object interface{}) bool {
			if user, ok := object.(ccv2.User); ok && user.Username == name {
				d.SetId(user.GUID)
				return true
			}
			return false
		}) {
			return nil
		}
		return diag.FromErr(NotFound)
	}

	for _, user := range users {
		if strings.ToLower(user.Username) == name {
			d.SetId(user.GUID)
			return nil
		}
	}
	return diag.FromErr(NotFound)
}
