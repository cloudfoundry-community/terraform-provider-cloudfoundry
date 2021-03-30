package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUser_WithGroups_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_user.admin-service-user"
	username := "cf-admin"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserDestroy(username),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: userResourceWithGroups,
				},

				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"password"},
				},
			},
		})
}
