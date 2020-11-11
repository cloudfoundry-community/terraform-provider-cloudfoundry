package cloudfoundry

import (
	"testing"

	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccApp_importBasic(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	resourceName := "cloudfoundry_app.dummy-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"dummy-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResource,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan, asset("dummy-app.zip")),
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateVerifyIgnore: []string{
						"timeout",
						"routes",
						"path",
						"strategy",
						"service_binding",
					},
				},
			},
		})
}
