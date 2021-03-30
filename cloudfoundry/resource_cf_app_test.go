package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/appdeployers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const appResourceTemplate = `

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
data "cloudfoundry_service" "s1" {
	name = "%s"
}
data "cloudfoundry_service" "s2" {
	name = "%s"
}

resource "cloudfoundry_route" "dummy-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "dummy-app"
}
resource "cloudfoundry_service_instance" "db" {
  name = "db"
	space = "${data.cloudfoundry_space.space.id}"
	service_plan = "${data.cloudfoundry_service.s1.service_plans.%s}"
}
resource "cloudfoundry_service_instance" "fs1" {
  name = "fs1"
	space = "${data.cloudfoundry_space.space.id}"
	service_plan = "${data.cloudfoundry_service.s2.service_plans.%s}"
}
%%s
resource "cloudfoundry_app" "dummy-app" {
  name = "dummy-app"
  buildpack = "binary_buildpack"
  space = "${data.cloudfoundry_space.space.id}"
  memory = "64"
  disk_quota = "512"
  timeout = 1800
  path = "%s"
	
%%s
}
`

const appResource = `

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
data "cloudfoundry_service" "s1" {
	name = "%s"
}
data "cloudfoundry_service" "s2" {
	name = "%s"
}

resource "cloudfoundry_route" "dummy-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "dummy-app"
}
resource "cloudfoundry_service_instance" "db" {
  name = "db"
	space = "${data.cloudfoundry_space.space.id}"
	service_plan = "${data.cloudfoundry_service.s1.service_plans.%s}"
}
resource "cloudfoundry_service_instance" "fs1" {
  name = "fs1"
	space = "${data.cloudfoundry_space.space.id}"
	service_plan = "${data.cloudfoundry_service.s2.service_plans.%s}"
}
resource "cloudfoundry_app" "dummy-app" {
  name = "dummy-app"
  buildpack = "binary_buildpack"
  space = "${data.cloudfoundry_space.space.id}"
  memory = "64"
  disk_quota = "512"
  timeout = 1800

  path = "%s"

  service_binding {
    service_instance = "${cloudfoundry_service_instance.db.id}"
  }
  service_binding {
    service_instance = "${cloudfoundry_service_instance.fs1.id}"
  }

  routes {
    route = "${cloudfoundry_route.dummy-app.id}"
  }

  environment = {
    TEST_VAR_1 = "testval1"
    TEST_VAR_2 = "testval2"
  }
}
`

const appResourceBlueGreen = `

data "cloudfoundry_domain" "local" {
	name = "%s"
}

resource "cloudfoundry_route" "dummy-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "%s"
  hostname = "dummy-app"
}

resource "cloudfoundry_app" "dummy-app" {
  name = "dummy-app"
  buildpack = "binary_buildpack"
  space = "%s"
  memory = "64"
  disk_quota = "512"
  timeout = 1800
  strategy = "blue-green"
  source_code_hash = "%s"
  path = "%s"

  routes {
    route = "${cloudfoundry_route.dummy-app.id}"
  }
}
`

const appResourceUpdate = `

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
data "cloudfoundry_service" "s1" {
	name = "%s"
}
data "cloudfoundry_service" "s2" {
	name = "%s"
}

resource "cloudfoundry_route" "dummy-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "dummy-app"
}
resource "cloudfoundry_service_instance" "db" {
  name = "db"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.s1.service_plans.%s}"
}
resource "cloudfoundry_service_instance" "fs1" {
  name = "fs1"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.s2.service_plans.%s}"
}
resource "cloudfoundry_service_instance" "fs2" {
  name = "fs2"
    space = "${data.cloudfoundry_space.space.id}"
    service_plan = "${data.cloudfoundry_service.s2.service_plans.%s}"
}
resource "cloudfoundry_app" "dummy-app" {
  name = "dummy-app-updated"
  buildpack = "binary_buildpack"
  space = "${data.cloudfoundry_space.space.id}"
  instances ="2"
  memory = "128"
  disk_quota = "1024"
  timeout = 1800

  path = "%s"

  service_binding {
    service_instance = "${cloudfoundry_service_instance.db.id}"
  }
  service_binding {
    service_instance = "${cloudfoundry_service_instance.fs1.id}"
  }
  service_binding {
    service_instance = "${cloudfoundry_service_instance.fs2.id}"
  }

  routes {
    route = "${cloudfoundry_route.dummy-app.id}"
  }

  environment = {
    TEST_VAR_1 = "testval1"
    TEST_VAR_2 = "testval2"
  }
}
`

const appResourceDocker = `

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

resource "cloudfoundry_route" "test-docker-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "test-docker-app"
}
resource "cloudfoundry_app" "test-docker-app" {
  name = "test-docker-app"
  space = "${data.cloudfoundry_space.space.id}"
  docker_image = "cloudfoundry/diego-docker-app:latest"
  timeout = 900
  routes {
    route = "${cloudfoundry_route.test-docker-app.id}"
  }
}
`

const multipleVersion = `
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

resource "cloudfoundry_route" "test-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "test-app"
  target {
    app = "${cloudfoundry_app.test-app.id}"
  }
}
resource "cloudfoundry_app" "test-app" {
  name = "test-app"
  space = "${data.cloudfoundry_space.space.id}"
  timeout = 1800
  buildpack = "binary_buildpack"
  memory = "64"
  path = "%s"
}
`

const multipleVersionUpdate = `
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

resource "cloudfoundry_route" "test-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "test-app"
  target {
    app = "${cloudfoundry_app.test-app.id}"
  }
}
resource "cloudfoundry_app" "test-app" {
  name = "test-app"
  space = "${data.cloudfoundry_space.space.id}"
  timeout = 1800
  buildpack = "binary_buildpack"
  memory = "128"

  path = "%s"
}
`

var appPath = asset("dummy-app.zip")

func TestAccResAppVersions_app1(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)
	refRoute := "cloudfoundry_route.test-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"test-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(multipleVersion, defaultAppDomain(), orgName, spaceName, appPath),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckRouteExists(refRoute, func() (err error) {

							if err = assertHTTPResponse("https://test-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(multipleVersionUpdate, defaultAppDomain(), orgName, spaceName, appPath),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckRouteExists(refRoute, func() (err error) {

							if err = assertHTTPResponse("https://test-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
					),
				},
			},
		})
}

func TestAccResApp_app1(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.dummy-app"

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
						serviceName1, serviceName2, servicePlan, servicePlan,
						appPath,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://dummy-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "dummy-app"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "64"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "2"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_1", "testval1"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_2", "testval2"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckResourceAttr(refApp, "service_binding.#", "2"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(appResourceUpdate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan, servicePlan,
						appPath,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://dummy-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "dummy-app-updated"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "2"),
						resource.TestCheckResourceAttr(refApp, "memory", "128"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "1024"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "2"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_1", "testval1"),
						resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_2", "testval2"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckResourceAttr(refApp, "service_binding.#", "3"),
					),
				},
			},
		})
}

func TestAccResApp_Routes_updateToAndmore(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.dummy-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"dummy-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan, appPath),
						``,
						`routes {
              route = "${cloudfoundry_route.dummy-app.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://dummy-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "dummy-app"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "64"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan, appPath),
						`resource "cloudfoundry_route" "dummy-app-2" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "dummy-app-2"
            }`,
						`routes {
              route = "${cloudfoundry_route.dummy-app.id}"
            }
            routes {
              route = "${cloudfoundry_route.dummy-app-2.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://dummy-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://dummy-app-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "dummy-app"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "64"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "2"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan, appPath),
						`resource "cloudfoundry_route" "dummy-app-2" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "dummy-app-2"
            }`,
						`routes {
              route = "${cloudfoundry_route.dummy-app.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://dummy-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://dummy-app-2."+defaultAppDomain(), 404, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "dummy-app"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "64"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan, appPath),
						`resource "cloudfoundry_route" "dummy-app-2" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "dummy-app-2"
            }
            resource "cloudfoundry_route" "dummy-app-3" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "dummy-app-3"
            }`,
						`routes {
              route = "${cloudfoundry_route.dummy-app-2.id}"
            }
            routes {
              route = "${cloudfoundry_route.dummy-app-3.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://dummy-app."+defaultAppDomain(), 404, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://dummy-app-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://dummy-app-3."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "dummy-app"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "64"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "2"),
					),
				},
			},
		})
}

func TestAccResApp_dockerApp(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)

	refApp := "cloudfoundry_app.test-docker-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"test-docker-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceDocker, defaultAppDomain(), orgName, spaceName),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://test-docker-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(
							refApp, "name", "test-docker-app"),
						resource.TestCheckResourceAttr(
							refApp, "space", spaceID),
						resource.TestCheckResourceAttr(
							refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(
							refApp, "instances", "1"),
						resource.TestCheckResourceAttrSet(
							refApp, "stack"),
						resource.TestCheckResourceAttr(
							refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(
							refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(
							refApp, "docker_image", "cloudfoundry/diego-docker-app:latest"),
					),
				},
			},
		})
}

func TestAccResApp_app_bluegreen(t *testing.T) {

	spaceID, _ := defaultTestSpace(t)

	refApp := "cloudfoundry_app.dummy-app"
	appDeploy := &appdeployers.AppDeploy{}
	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"dummy-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceBlueGreen,
						defaultAppDomain(),
						spaceID, spaceID,
						"1",
						appPath,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExistsInject(refApp, appDeploy, func() (err error) {

							if err = assertHTTPResponse("https://dummy-app."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(appResourceBlueGreen,
						defaultAppDomain(),
						spaceID, spaceID,
						"2",
						appPath,
					),
					Check: resource.ComposeTestCheckFunc(
						func(s *terraform.State) error {
							err := assertHTTPResponse("https://dummy-app."+defaultAppDomain(), 200, nil)
							if err != nil {
								return err
							}
							rs, ok := s.RootModule().Resources[refApp]
							if !ok {
								return fmt.Errorf("app '%s' not found in terraform state", refApp)
							}

							id := rs.Primary.ID
							if id == appDeploy.App.GUID {
								return fmt.Errorf("After blue green deployment, app must have changed but GUID are the same between previous and update")
							}
							return nil
						},
					),
				},
			},
		})
}

func testAccCheckAppExistsInject(resApp string, appDeploy *appdeployers.AppDeploy, validate func() error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resApp]
		if !ok {
			return fmt.Errorf("app '%s' not found in terraform state", resApp)
		}

		id := rs.Primary.ID

		app, _, err := session.ClientV2.GetApplication(id)
		if err != nil {
			return err
		}

		serviceBindings, _, err := session.ClientV2.GetServiceBindings(ccv2.FilterEqual(constant.AppGUIDFilter, id))
		if err != nil {
			return err
		}

		routeMappings, _, err := session.ClientV2.GetRouteMappings(ccv2.FilterEqual(constant.AppGUIDFilter, id))
		if err != nil {
			return err
		}
		appDeploy.App = app
		appDeploy.ServiceBindings = serviceBindings
		appDeploy.Mappings = routeMappings
		return validate()
	}
}
func testAccCheckAppExists(resApp string, validate func() error) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resApp]
		if !ok {
			return fmt.Errorf("app '%s' not found in terraform state", resApp)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		app, _, err := session.ClientV2.GetApplication(id)
		if err != nil {
			return err
		}

		if err = assertEquals(attributes, "name", app.Name); err != nil {
			return err
		}
		if err = assertEquals(attributes, "space", app.SpaceGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "instances", app.Instances.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "memory", app.Memory.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "disk_quota", app.DiskQuota.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "stack", app.StackGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "buildpack", app.Buildpack.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "command", app.Command.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "enable_ssh", app.EnableSSH.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_http_endpoint", app.HealthCheckHTTPEndpoint); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_type", string(app.HealthCheckType)); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_timeout", app.HealthCheckTimeout); err != nil {
			return err
		}
		envVars := make(map[string]interface{})
		for k, v := range app.EnvironmentVariables {
			envVars[k] = v
		}
		if err = assertMapEquals("environment", attributes, envVars); err != nil {
			return err
		}
		serviceBindings, _, err := session.ClientV2.GetServiceBindings(ccv2.FilterEqual(constant.AppGUIDFilter, id))
		if err != nil {
			return err
		}
		if err = assertListEquals(attributes, "service_binding", len(serviceBindings),
			func(values map[string]string, i int) (match bool) {
				found := false
				for _, b := range serviceBindings {
					if values["service_instance"] == b.ServiceInstanceGUID {
						found = true
						break
					}
				}
				return found

			}); err != nil {
			return err
		}
		routeMappingsC, _, err := session.ClientV2.GetRouteMappings(ccv2.FilterEqual(constant.AppGUIDFilter, id))
		if err != nil {
			return err
		}
		routeMappings := make([]map[string]interface{}, 0)
		for _, mapping := range routeMappingsC {
			curMapping := make(map[string]interface{})
			curMapping["route"] = mapping.RouteGUID
			curMapping["port"] = mapping.AppPort
			routeMappings = append(routeMappings, curMapping)
		}
		if err = validateRouteMappings(attributes, routeMappings); err != nil {
			return err
		}

		return validate()
	}
}

func validateRouteMappings(attributes map[string]string, routeMappings []map[string]interface{}) (err error) {
	entity := resourceApp()
	reader := &schema.MapFieldReader{
		Schema: entity.Schema,
		Map:    schema.BasicMapReader(attributes),
	}
	result, err := reader.ReadField([]string{"routes"})
	if err != nil {
		return err
	}
	routesTf := getListOfStructs(result.Value)
	for _, routeTf := range routesTf {
		match := func(object interface{}) bool {
			routeMapping := object.(map[string]interface{})
			return routeTf["route"] == routeMapping["route"]
		}
		if !isInSlice(routeMappings, match) {
			return fmt.Errorf("unable to find route mapping for route '%s' and port '%d' matching cf mapping",
				routeTf["route"], routeTf["port"],
			)
		}
	}
	return nil
}

func testAccCheckAppDestroyed(apps []string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)
		for _, a := range apps {
			apps, _, err := session.ClientV2.GetApplications(ccv2.FilterByName(a))
			if err != nil {
				return err
			}
			if len(apps) > 0 {
				_, err := session.ClientV2.DeleteApplication(apps[0].GUID)
				return err
			}
		}
		return nil
	}
}
