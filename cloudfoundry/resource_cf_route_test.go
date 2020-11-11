package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const routeResource = `

data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "%s"
}
data "cloudfoundry_space" "space" {
    name = "%s"
	org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_app" "test-app-8080" {
	name = "test-app"
	space = "${data.cloudfoundry_space.space.id}"
	timeout = 1800
	buildpack = "binary_buildpack"
	path = "%s"
}
resource "cloudfoundry_route" "test-app-route" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "test-app-single"

	target {
		app = "${cloudfoundry_app.test-app-8080.id}"
	}
}
`

const routeResourceUpdate = `

data "cloudfoundry_domain" "local" {
    name = "%s"
}
data "cloudfoundry_org" "org" {
    name = "%s"
}
data "cloudfoundry_space" "space" {
    name = "%s"
	org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_app" "test-app-8080" {
	name = "test-app-8080"
	space = "${data.cloudfoundry_space.space.id}"
	timeout = 1800
	buildpack = "binary_buildpack"
	path = "%s"
}
resource "cloudfoundry_app" "test-app-8888" {
	name = "test-app-8888"
	space = "${data.cloudfoundry_space.space.id}"
	ports = [ 8888 ]
	timeout = 1800
	buildpack = "binary_buildpack"
	path = "%s"
}
resource "cloudfoundry_app" "test-app-9999" {
	name = "test-app-9999"
	space = "${data.cloudfoundry_space.space.id}"
	ports = [ 9999 ]
	timeout = 1800
	buildpack = "binary_buildpack"
	path = "%s"
}
resource "cloudfoundry_route" "test-app-route" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "test-app-multi"

	target {
		app = "${cloudfoundry_app.test-app-9999.id}"
		port = 9999
	}
	target {
		app = "${cloudfoundry_app.test-app-8888.id}"
		port = 8888
	}
	target {
		app = "${cloudfoundry_app.test-app-8080.id}"
	}
}
`

func TestAccResRoute_normal(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)

	refRoute := "cloudfoundry_route.test-app-route"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckRouteDestroyed([]string{"test-app-single", "test-app-multi"}, defaultAppDomain()),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(routeResource,
						defaultAppDomain(),
						orgName, spaceName,
						appPath),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckRouteExists(refRoute, func() (err error) {
							if err = assertHTTPResponse("http://test-app-single."+defaultAppDomain(), 200, nil); err != nil {
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
					Config: fmt.Sprintf(routeResourceUpdate,
						defaultAppDomain(),
						orgName, spaceName,
						appPath, appPath, appPath),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckRouteExists(refRoute, func() (err error) {

							for i := 1; i <= 9; i++ {
								if err = assertHTTPResponse("http://test-app-multi."+defaultAppDomain(), 200, nil); err != nil {
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

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resRoute]
		if !ok {
			return fmt.Errorf("route '%s' not found in terraform state", resRoute)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		route, _, err := session.ClientV2.GetRoute(id)
		if err != nil {
			return err
		}

		if err = assertEquals(attributes, "domain", route.DomainGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "space", route.SpaceGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "hostname", route.Host); err != nil {
			return err
		}
		if err = assertEquals(attributes, "port", route.Port); err != nil {
			return err
		}
		if err = assertEquals(attributes, "path", route.Path); err != nil {
			return err
		}

		return validate()
	}
}

func testAccCheckRouteDestroyed(hostnames []string, domain string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
		domains, _, err := session.ClientV2.GetSharedDomains(ccv2.FilterByName(domain))
		if err != nil || len(domains) == 0 {
			domains, _, err = session.ClientV2.GetPrivateDomains(ccv2.FilterByName(domain))
			if err != nil {
				return err
			}
		}
		if len(domains) == 0 {
			return fmt.Errorf("Domain %s not found", domain)
		}
		domainGuid := domains[0].GUID
		for _, h := range hostnames {
			routes, _, err := session.ClientV2.GetRoutes(ccv2.FilterEqual(constant.HostFilter, h), ccv2.FilterEqual(constant.DomainGUIDFilter, domainGuid))
			if err != nil {
				return err
			}
			if len(routes) > 0 {
				return fmt.Errorf("route with hostname '%s' and domain '%s' still exists in cloud foundry", h, domain)
			}
		}
		return nil
	}
}
