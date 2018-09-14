package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDefaultRunningAsg_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_default_asg.running"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDefaultRunningAsgDestroy,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: defaultRunningSecurityGroupResource,
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
