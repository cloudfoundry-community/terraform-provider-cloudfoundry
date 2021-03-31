package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func dataSourceSpaceQuota() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSpaceQuotaRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceSpaceQuotaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	qm := session.ClientV2

	name := d.Get("name").(string)
	orgId := d.Get("org").(string)
	quotas, _, err := qm.GetQuotas(constant.SpaceQuota)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, quota := range quotas {
		if quota.Name != name {
			continue
		}
		if orgId != "" && quota.OrganizationGUID != orgId {
			continue
		}
		d.SetId(quota.GUID)
		return nil
	}
	return diag.FromErr(NotFound)
}
