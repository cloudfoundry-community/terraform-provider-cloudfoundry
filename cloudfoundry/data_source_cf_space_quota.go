package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func dataSourceSpaceQuotaRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	qm := session.ClientV2

	name := d.Get("name").(string)
	orgId := d.Get("org").(string)
	quotas, _, err := qm.GetQuotas(constant.SpaceQuota)
	if err != nil {
		return err
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
	return NotFound
}
