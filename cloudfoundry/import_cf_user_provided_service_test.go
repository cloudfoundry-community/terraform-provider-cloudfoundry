package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserProvidedService_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_user_provided_service.mq"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserProvidedServiceDestroyed("mq", "cloudfoundry_space.space1"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: userProvidedServiceResourceCreate,
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
}
