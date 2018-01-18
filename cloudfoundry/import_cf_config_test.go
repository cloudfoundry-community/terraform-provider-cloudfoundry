package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccConfig_importBasic(t *testing.T) {
	resourceName := "cf_config.ff"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckConfigDestroy,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: configResource,
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
