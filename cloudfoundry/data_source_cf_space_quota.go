package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
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
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		org    *string
		name   string
		quota  cfapi.CCQuota
	)

	name = d.Get("name").(string)
	if val, ok := d.GetOk("org"); ok {
		name := val.(string)
		org = &name
	}

	qm := session.QuotaManager()
	quota, err = qm.FindQuotaByName(cfapi.SpaceQuota, name, org)
	if err != nil {
		return err
	}
	d.SetId(quota.ID)
	return nil
}
