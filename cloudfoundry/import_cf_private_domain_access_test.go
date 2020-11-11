package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPrivateDomainAccess_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_private_domain_access.access-to-org"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(privateDomainAccessResourceCreate, defaultAppDomain()),
				},
				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
