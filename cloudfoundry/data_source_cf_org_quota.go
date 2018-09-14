package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
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
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		name  string
		quota cfapi.CCQuota
	)

	name = d.Get("name").(string)
	qm := session.QuotaManager()
	quota, err = qm.FindQuotaByName(cfapi.OrgQuota, name, nil)
	if err != nil {
		return err
	}
	d.SetId(quota.ID)
	return nil
}
