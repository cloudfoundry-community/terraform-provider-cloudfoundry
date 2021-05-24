package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrgQuota() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOrgQuotaRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceOrgQuotaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	qm := session.ClientV2

	name := d.Get("name").(string)
	quotas, _, err := qm.GetQuotas(constant.OrgQuota, ccv2.FilterByName(name))
	if err != nil {
		return diag.FromErr(err)
	}
	if len(quotas) == 0 {
		return diag.FromErr(NotFound)
	}
	d.SetId(quotas[0].GUID)
	return nil
}
