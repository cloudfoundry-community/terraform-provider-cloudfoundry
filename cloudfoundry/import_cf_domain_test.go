package cloudfoundry

import (
	"testing"

	"fmt"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDomain_importBasic(t *testing.T) {
	resourceName := "cf_domain.shared"
	domainname := "dev." + defaultAppDomain()

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSharedDomainDestroy(domainname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(domainResourceShared, defaultAppDomain()),
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
