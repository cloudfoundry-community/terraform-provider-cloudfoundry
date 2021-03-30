package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSpace_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_space.space1"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSpaceDestroyed("space-one"),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: spaceResource,
				},
				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"staging_asgs", "asgs"},
				},
			},
		})
}
