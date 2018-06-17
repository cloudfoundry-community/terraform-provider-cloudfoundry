package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAsg_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_asg.rmq"
	asgname := "rmq-dev"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckASGDestroy(asgname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: securityGroup,
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
