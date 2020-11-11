package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccOrgQuota_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_org_quota.quota50g-org"
	quotaname := "quota50g-org"

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
