package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrgQuota() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOrgQuotaRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceOrgQuotaRead(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*managers.Session)
	qm := session.ClientV2

	var (
		name   string
		quotas []ccv2.Quota
	)

	name = d.Get("name").(string)
	quotas, _, err = qm.GetQuotas(constant.OrgQuota, ccv2.FilterByName(name))
	if err != nil {
		return err
	}
	if len(quotas) == 0 {
		return NotFound
	}
	d.SetId(quotas[0].GUID)
	return nil
}
