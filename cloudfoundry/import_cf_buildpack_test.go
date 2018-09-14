package cloudfoundry

import (
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"os"
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
					Config: fmt.Sprintf(buildpackResource, os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_TOKEN")),
				},

				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"github_release"},
				},
			},
		})
}
