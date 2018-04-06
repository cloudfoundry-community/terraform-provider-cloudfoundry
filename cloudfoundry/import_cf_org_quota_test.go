package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccOrgQuota_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_org_quota.50g-org"
	quotaname := "50g-org"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckOrgQuotaResourceDestroy(quotaname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: orgQuotaResource,
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
