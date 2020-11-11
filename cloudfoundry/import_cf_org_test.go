package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccOrg_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_org.org1"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckOrgDestroyed("organization-one-updated"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: orgResource,
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
