package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccEvg_importBasic(t *testing.T) {
	resourceName := "cf_evg.running"
	name := "running"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckEvgDestroy(name),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: evgRunningResource,
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
