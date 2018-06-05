package cloudfoundry

import (
	"testing"

	"fmt"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccApp_importBasic(t *testing.T) {
	resourceName := "cf_app.spring-music"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"spring-music"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceSpringMusic, defaultAppDomain()),
				},

				resource.TestStep{
					ResourceName:      resourceName,
					ImportState:       true,
					ImportStateVerify: true,
					ImportStateVerifyIgnore: []string{
						"timeout",
						"route",
						"url",
						"service_binding.0.credentials",
						"service_binding.1.credentials",
						"buildpack",
						"command",
						"health_check_http_endpoint",
						"health_check_timeout",
					},
				},
			},
		})
}
