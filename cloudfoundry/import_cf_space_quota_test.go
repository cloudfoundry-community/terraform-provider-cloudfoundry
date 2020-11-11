package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSpaceQuota_importBasic(t *testing.T) {

	orgID, _ := defaultTestOrg(t)

	resourceName := "cloudfoundry_space_quota.quota10g-space"
	quotaname := "10g-space"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSpaceQuotaResourceDestroy(quotaname, orgID),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: spaceQuotaResource,
				},
				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
