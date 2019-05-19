package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccBuildpack_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_buildpack.tomee"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckBuildpackDestroyed("tomee-buildpack"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: buildpackResource,
				},

				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"path"},
				},
			},
		})
}
