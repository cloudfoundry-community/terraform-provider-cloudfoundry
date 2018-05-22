package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccPrivateDomainAccess_importBasic(t *testing.T) {
	resourceName := "cf_private_domain_access.access-to-org"

	// toIgnore := []string{"staging_asgs.#"}
	// session := testSession()
	// asm := session.ASGManager()
	// secGroup, err := asm.Read("public_networks")
	// if err == nil && secGroup.GUID != "" {
	// 	toIgnore = append(toIgnore, fmt.Sprintf("staging_asgs.%d", resourceStringHash(secGroup.GUID)))
	// }

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
