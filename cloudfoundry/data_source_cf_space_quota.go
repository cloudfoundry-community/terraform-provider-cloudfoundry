package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func dataSourceSpaceQuota() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSpaceQuotaRead,
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

func dataSourceSpaceQuotaRead(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*managers.Session)
	qm := session.ClientV2

	var (
		name   string
		quotas []ccv2.Quota
	)

	name = d.Get("name").(string)
	quotas, _, err = qm.GetQuotas(constant.SpaceQuota, ccv2.FilterByName(name), ccv2.FilterByOrg(d.Get("org").(string)))
	if err != nil {
		return err
	}
	if len(quotas) == 0 {
		return NotFound
	}
	d.SetId(quotas[0].GUID)
	return nil
}
