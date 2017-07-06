package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccConfig_normal(t *testing.T) {

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckConfigDestroy,
			Steps:        []resource.TestStep{},
		})
}

func testAccCheckConfigDestroy(s *terraform.State) error {
	return nil
}
