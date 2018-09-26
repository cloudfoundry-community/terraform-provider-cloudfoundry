package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccSpaceQuota_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_space_quota.10g-space"
	quotaname := "10g-space"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSpaceQuotaResourceDestroy(quotaname),
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
