package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccQuota_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_quota.50g"
	quotaname := "50g"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckQuotaResourceDestroy(quotaname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: quotaResource,
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
