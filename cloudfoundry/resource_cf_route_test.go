package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

const routeResource = `

data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_app" "test-app-8080" {
	name = "test-app"
	space_id = "${data.cloudfoundry_space.space.id}"
	command = "test-app --ports=8080"
	timeout = 1800

	git {
		url = "https://github.com/mevansam/test-app.git"
	}
}
resource "cloudfoundry_route" "test-app-route" {
	domain_id = "${data.cloudfoundry_domain.local.id}"
	space_id = "${data.cloudfoundry_space.space.id}"
	hostname = "test-app-single"

	target {
		app_id = "${cloudfoundry_app.test-app-8080.id}"
	}
}
`

const routeResourceUpdate = `

data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "pcfdev-org"
}
data "cloudfoundry_space" "space" {
    name = "pcfdev-space"
	org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_app" "test-app-8080" {
	name = "test-app-8080"
	space_id = "${data.cloudfoundry_space.space.id}"
	command = "test-app --ports=8080"
	timeout = 1800

	git {
		url = "https://github.com/mevansam/test-app.git"
	}
}
resource "cloudfoundry_app" "test-app-8888" {
	name = "test-app-8888"
	space_id = "${data.cloudfoundry_space.space.id}"
	ports = [ 8888 ]
	command = "test-app --ports=8888"
	timeout = 1800

	git {
		url = "https://github.com/mevansam/test-app.git"
	}
}
resource "cloudfoundry_app" "test-app-9999" {
	name = "test-app-9999"
	space_id = "${data.cloudfoundry_space.space.id}"
	ports = [ 9999 ]
	command = "test-app --ports=9999"
	timeout = 1800

	git {
		url = "https://github.com/mevansam/test-app.git"
	}
}
resource "cloudfoundry_route" "test-app-route" {
	domain_id = "${data.cloudfoundry_domain.local.id}"
	space_id = "${data.cloudfoundry_space.space.id}"
	hostname = "test-app-multi"

	target {
		app_id = "${cloudfoundry_app.test-app-9999.id}"
		port = 9999
	}
	target {
		app_id = "${cloudfoundry_app.test-app-8888.id}"
		port = 8888
	}
	target {
		app_id = "${cloudfoundry_app.test-app-8080.id}"
	}
}
`

func TestAccRoute_normal(t *testing.T) {

	refRoute := "cloudfoundry_route.test-app-route"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckRouteDestroyed([]string{"test-app-single", "test-app-multi"}, defaultAppDomain()),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(routeResource, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckRouteExists(refRoute, func() (err error) {

							responses := []string{"8080"}
							if err = assertHTTPResponse("http://test-app-single."+defaultAppDomain()+"/port", 200, &responses); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(
							refRoute, "hostname", "test-app-single"),
						resource.TestCheckResourceAttr(
							refRoute, "target.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(routeResourceUpdate, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckRouteExists(refRoute, func() (err error) {

							responses := []string{"8080", "8888", "9999"}
							for i := 1; i <= 9; i++ {
								if err = assertHTTPResponse("http://test-app-multi."+defaultAppDomain()+"/port", 200, &responses); err != nil {
									return err
								}
							}
							return
						}),
						resource.TestCheckResourceAttr(
							refRoute, "hostname", "test-app-multi"),
						resource.TestCheckResourceAttr(
							refRoute, "target.#", "3"),
					),
				},
			},
		})
}

func testAccCheckRouteExists(resRoute string, validate func() error) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resRoute]
		if !ok {
			return fmt.Errorf("route '%s' not found in terraform state", resRoute)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resRoute, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var route cfapi.CCRoute
		rm := session.RouteManager()
		if route, err = rm.ReadRoute(id); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved route for resource '%s' with id '%s': %# v",
			resRoute, id, route)

		if err = assertEquals(attributes, "domain_id", route.DomainGUID); err != nil {
			return
		}
		if err = assertEquals(attributes, "space_id", route.SpaceGUID); err != nil {
			return
		}
		if err = assertEquals(attributes, "hostname", route.Hostname); err != nil {
			return
		}
		if err = assertEquals(attributes, "port", route.Port); err != nil {
			return
		}
		if err = assertEquals(attributes, "path", route.Path); err != nil {
			return
		}

		err = validate()
		return
	}
}

func testAccCheckRouteDestroyed(hostnames []string, domain string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)
		for _, h := range hostnames {
			if _, err := session.RouteManager().FindRoute(domain, &h, nil, nil); err != nil {
				switch err.(type) {
				case *errors.ModelNotFoundError:
					continue
				default:
					return err
				}
			}
			return fmt.Errorf("route with hostname '%s' and domain '%s' still exists in cloud foundry", h, domain)
		}
		return nil
	}
}
