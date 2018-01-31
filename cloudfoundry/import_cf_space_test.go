package cloudfoundry

import (
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccSpace_importBasic(t *testing.T) {
	resourceName := "cf_space.space1"

	toIgnore := []string{"staging_asgs.#"}

	session := testSession()
	asm := session.ASGManager()
	secGroup, err := asm.Read("public_networks")
	if err == nil && secGroup.GUID != "" {
		toIgnore = append(toIgnore, fmt.Sprintf("staging_asgs.%d", resourceStringHash(secGroup.GUID)))
	}

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSpaceDestroyed("space-one"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: spaceResource,
				},

				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: toIgnore,
				},
			},
		})
}
