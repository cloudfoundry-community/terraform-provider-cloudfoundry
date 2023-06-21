package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	constantV3 "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

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
  hostname = "dummy-app-tf"
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
  enable_ssh = true
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
    space = "${data.cloudfoundry_space.space.id}"
}
data "cloudfoundry_service" "s2" {
	name = "%s"
    space = "${data.cloudfoundry_space.space.id}"
}

resource "cloudfoundry_route" "dummy-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "dummy-app-tf"
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
  enable_ssh = true

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
  hostname = "dummy-app-tf"
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
  enable_ssh = true

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
    space = "${data.cloudfoundry_space.space.id}"
}
data "cloudfoundry_service" "s2" {
	name = "%s"
    space = "${data.cloudfoundry_space.space.id}"
}

resource "cloudfoundry_route" "dummy-app" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "dummy-app-tf"
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
  enable_ssh = true

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
  enable_ssh = true

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
  hostname = "test-app-tf"
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
  enable_ssh = true

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
  hostname = "test-app-tf"
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
  enable_ssh = true

}
`

const defaultValues = `
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

resource "cloudfoundry_route" "app_1" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "app-1-tf"
	target {
		app = "${cloudfoundry_app.app_1.id}"
	}
}
resource "cloudfoundry_app" "app_1" {
	name = "app-1"
	space = "${data.cloudfoundry_space.space.id}"
	buildpack = "binary_buildpack"

	path = "%s"
	strategy = "%s"
}
`

const overrideDefaultValues = `
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

resource "cloudfoundry_route" "app_1" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "app-1-tf"
	target {
		app = "${cloudfoundry_app.app_1.id}"
	}
}
resource "cloudfoundry_app" "app_1" {
	name = "app-1-update"
	space = "${data.cloudfoundry_space.space.id}"
	buildpack = "binary_buildpack"
	memory = 128
	disk_quota = 64
	instances = 2
	enable_ssh = true
	
	path = "%s"
	strategy = "%s"
}
`

const appResourceDockerInvocationTimeout = `

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

resource "cloudfoundry_route" "app_route_test_timeout" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "app_route_test_timeout"
}
resource "cloudfoundry_app" "test-docker-app" {
  name = "test-docker-app"
  space = "${data.cloudfoundry_space.space.id}"
  docker_image = "cloudfoundry/diego-docker-app:latest"
  timeout = 900
  enable_ssh = true
  health_check_invocation_timeout = %d

  routes {
    route = "${cloudfoundry_route.app_route_test_timeout.id}"
  }
}
`

const multipleBuildpacks = `
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
resource "cloudfoundry_route" "app_1" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "app-1-tf"
	target {
		app = "${cloudfoundry_app.app_1.id}"
	}
}
resource "cloudfoundry_app" "app_1" {
	name = "app-1"
	space = "${data.cloudfoundry_space.space.id}"
	buildpacks = ["binary_buildpack", "tomee_buildpack"]

	path = "%s"
	strategy = "%s"
}
`

var appPath = asset("dummy-app.zip")

var appPaths = []struct {
	typeOfPath string
	path       string
}{
	{typeOfPath: "local", path: asset("dummy-app.zip")},
	{typeOfPath: "remote", path: "https://raw.githubusercontent.com/cloudfoundry-community/terraform-provider-cloudfoundry/main/tests/cf-acceptance-tests/assets/dummy-app.zip"},
}

func TestAccResAppVersions_app1(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)
	refRoute := "cloudfoundry_route.test-app"

	for _, app := range appPaths {
		appPath = app.path

		t.Run(fmt.Sprintf("AppSource=%s", app.typeOfPath), func(t *testing.T) {
			resource.Test(t,
				resource.TestCase{
					PreCheck:          func() { testAccPreCheck(t) },
					ProviderFactories: testAccProvidersFactories,
					CheckDestroy:      testAccCheckAppDestroyed([]string{"test-app"}),
					Steps: []resource.TestStep{

						resource.TestStep{
							Config: fmt.Sprintf(multipleVersion, defaultAppDomain(), orgName, spaceName, appPath),
							Check: resource.ComposeTestCheckFunc(
								testAccCheckRouteExists(refRoute, func() (err error) {

									if err = assertHTTPResponse("https://test-app-tf."+defaultAppDomain(), 200, nil); err != nil {
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

									if err = assertHTTPResponse("https://test-app-tf."+defaultAppDomain(), 200, nil); err != nil {
										return err
									}
									return
								}),
							),
						},
					},
				})
		})
	}
}

func TestAccDefaultValues_app1(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	refApp := "cloudfoundry_app.app_1"

	// Default memory and disk quota is managed globally
	// defaultDiskQuota := 1024
	// defaultMemory := 1024
	defaultInstances := 1
	defaultPort := 8080
	// whether ssh is disabled or not depends on the global setting
	// globalSSHEnabled := "true"

	for _, app := range appPaths {
		appPath = app.path

		t.Run(fmt.Sprintf("AppSource=%s", app.typeOfPath), func(t *testing.T) {
			resource.Test(t,
				resource.TestCase{
					PreCheck:          func() { testAccPreCheck(t) },
					ProviderFactories: testAccProvidersFactories,
					CheckDestroy:      testAccCheckAppDestroyed([]string{"app-1"}),
					Steps: []resource.TestStep{

						resource.TestStep{
							Config: fmt.Sprintf(defaultValues, defaultAppDomain(), orgName, spaceName, appPath, "standard"),
							Check: resource.ComposeTestCheckFunc(
								testAccCheckAppExists(refApp, func() (err error) {

									if err = assertHTTPResponse("https://app-1-tf."+defaultAppDomain(), 200, nil); err != nil {
										return err
									}
									return
								}),
								resource.TestCheckResourceAttr(refApp, "name", "app-1"),
								resource.TestCheckResourceAttr(refApp, "space", spaceID),
								resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
								resource.TestCheckResourceAttr(refApp, "ports.0", fmt.Sprint(defaultPort)),
								resource.TestCheckResourceAttr(refApp, "instances", fmt.Sprint(defaultInstances)),
								// resource.TestCheckResourceAttr(refApp, "memory", fmt.Sprint(defaultMemory)),
								// resource.TestCheckResourceAttr(refApp, "disk_quota", fmt.Sprint(defaultDiskQuota)),
								resource.TestCheckResourceAttrSet(refApp, "stack"),
								// resource.TestCheckResourceAttr(refApp, "enable_ssh", globalSSHEnabled),
							),
						},

						resource.TestStep{
							Config: fmt.Sprintf(overrideDefaultValues, defaultAppDomain(), orgName, spaceName, appPath, "standard"),
							Check: resource.ComposeTestCheckFunc(
								testAccCheckAppExists(refApp, func() (err error) {
									if err = assertHTTPResponse("https://app-1-tf."+defaultAppDomain(), 200, nil); err != nil {
										return err
									}
									return
								}),
								resource.TestCheckResourceAttr(refApp, "name", "app-1-update"),
								resource.TestCheckResourceAttr(refApp, "space", spaceID),
								resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
								resource.TestCheckResourceAttr(refApp, "ports.0", "8080"),
								resource.TestCheckResourceAttr(refApp, "instances", "2"),
								resource.TestCheckResourceAttr(refApp, "memory", "128"),
								resource.TestCheckResourceAttr(refApp, "disk_quota", "64"),
								resource.TestCheckResourceAttrSet(refApp, "stack"),
								resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
							),
						},
					},
				})

		})
	}
}

func TestAccDefaultValuesRolling_app1(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	refApp := "cloudfoundry_app.app_1"

	// Default memory and disk quota is managed globally
	// defaultDiskQuota := 1024
	// defaultMemory := 1024
	defaultInstances := 1
	defaultPort := 8080
	// whether ssh is disabled or not depends on the global setting
	// globalSSHEnabled := "true"

	for _, app := range appPaths {
		appPath = app.path

		t.Run(fmt.Sprintf("AppSource=%s", app.typeOfPath), func(t *testing.T) {
			resource.Test(t,
				resource.TestCase{
					PreCheck:          func() { testAccPreCheck(t) },
					ProviderFactories: testAccProvidersFactories,
					CheckDestroy:      testAccCheckAppDestroyed([]string{"app-1"}),
					Steps: []resource.TestStep{

						resource.TestStep{
							Config: fmt.Sprintf(defaultValues, defaultAppDomain(), orgName, spaceName, appPath, "rolling"),
							Check: resource.ComposeTestCheckFunc(
								testAccCheckAppExists(refApp, func() (err error) {

									if err = assertHTTPResponse("https://app-1-tf."+defaultAppDomain(), 200, nil); err != nil {
										return err
									}
									return
								}),
								resource.TestCheckResourceAttr(refApp, "name", "app-1"),
								resource.TestCheckResourceAttr(refApp, "space", spaceID),
								resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
								resource.TestCheckResourceAttr(refApp, "ports.0", fmt.Sprint(defaultPort)),
								resource.TestCheckResourceAttr(refApp, "instances", fmt.Sprint(defaultInstances)),
								// resource.TestCheckResourceAttr(refApp, "memory", fmt.Sprint(defaultMemory)),
								// resource.TestCheckResourceAttr(refApp, "disk_quota", fmt.Sprint(defaultDiskQuota)),
								resource.TestCheckResourceAttrSet(refApp, "stack"),
								// resource.TestCheckResourceAttr(refApp, "enable_ssh", globalSSHEnabled),
							),
						},

						resource.TestStep{
							Config: fmt.Sprintf(overrideDefaultValues, defaultAppDomain(), orgName, spaceName, appPath, "rolling"),
							Check: resource.ComposeTestCheckFunc(
								testAccCheckAppExists(refApp, func() (err error) {
									if err = assertHTTPResponse("https://app-1-tf."+defaultAppDomain(), 200, nil); err != nil {
										return err
									}
									return
								}),
								resource.TestCheckResourceAttr(refApp, "name", "app-1-update"),
								resource.TestCheckResourceAttr(refApp, "space", spaceID),
								resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
								resource.TestCheckResourceAttr(refApp, "ports.0", "8080"),
								resource.TestCheckResourceAttr(refApp, "instances", "2"),
								resource.TestCheckResourceAttr(refApp, "memory", "128"),
								resource.TestCheckResourceAttr(refApp, "disk_quota", "64"),
								resource.TestCheckResourceAttrSet(refApp, "stack"),
								resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
							),
						},
					},
				})
		})
	}
}

func TestAccResApp_app1(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.dummy-app"

	for _, app := range appPaths {
		appPath = app.path

		t.Run(fmt.Sprintf("AppSource=%s", app.typeOfPath), func(t *testing.T) {

			resource.Test(t,
				resource.TestCase{
					PreCheck:          func() { testAccPreCheck(t) },
					ProviderFactories: testAccProvidersFactories,
					CheckDestroy:      testAccCheckAppDestroyed([]string{"dummy-app"}),
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

									if err = assertHTTPResponse("https://dummy-app-tf."+defaultAppDomain(), 200, nil); err != nil {
										return err
									}
									return
								}),
								resource.TestCheckResourceAttr(refApp, "name", "dummy-app"),
								resource.TestCheckResourceAttr(refApp, "space", spaceID),
								resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
								resource.TestCheckResourceAttr(refApp, "ports.0", "8080"),
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
							Config: fmt.Sprintf(appResource,
								defaultAppDomain(),
								orgName, spaceName,
								serviceName1, serviceName2, servicePlan, servicePlan,
								appPath,
							),
							Check: resource.ComposeTestCheckFunc(
								setEnvironmentVariables(refApp, map[string]interface{}{
									"TEST_VAR_3": "testval3",
								}),
							),
							ExpectNonEmptyPlan: true,
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

									if err = assertHTTPResponse("https://dummy-app-tf."+defaultAppDomain(), 200, nil); err != nil {
										return err
									}
									return
								}),
								testAccCheckEnvironmentVariableSet(refApp, "TEST_VAR_1"),
								testAccCheckEnvironmentVariableNotSet(refApp, "TEST_VAR_2"),
								testAccCheckEnvironmentVariableNotSet(refApp, "TEST_VAR_3"),
								resource.TestCheckResourceAttr(refApp, "name", "dummy-app-updated"),
								resource.TestCheckResourceAttr(refApp, "space", spaceID),
								resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
								resource.TestCheckResourceAttr(refApp, "ports.0", "8080"),
								resource.TestCheckResourceAttr(refApp, "instances", "2"),
								resource.TestCheckResourceAttr(refApp, "memory", "128"),
								resource.TestCheckResourceAttr(refApp, "disk_quota", "1024"),
								resource.TestCheckResourceAttrSet(refApp, "stack"),
								resource.TestCheckResourceAttr(refApp, "environment.%", "1"),
								resource.TestCheckResourceAttr(refApp, "environment.TEST_VAR_1", "testval1"),
								resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
								resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
								resource.TestCheckResourceAttr(refApp, "service_binding.#", "3"),
							),
						},
					},
				})
		})
	}
}

func testAccCheckEnvironmentVariableSet(resApp string, envVar string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)
		rs, ok := s.RootModule().Resources[resApp]
		if !ok {
			return fmt.Errorf("app '%s' not found in terraform state", resApp)
		}
		id := rs.Primary.ID

		vars, err := session.BitsManager.GetAppEnvironmentVariables(id)
		if err != nil {
			return err
		}
		for v := range vars {
			if v == envVar {
				return nil
			}
		}
		return fmt.Errorf("environment variable '%s' not set", envVar)
	}
}

func testAccCheckEnvironmentVariableNotSet(resApp string, envVar string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)
		rs, ok := s.RootModule().Resources[resApp]
		if !ok {
			return fmt.Errorf("app '%s' not found in terraform state", resApp)
		}
		id := rs.Primary.ID

		vars, err := session.BitsManager.GetAppEnvironmentVariables(id)
		if err != nil {
			return err
		}
		for v := range vars {
			if v == envVar {
				return fmt.Errorf("environemnt variable '%s' is set", envVar)
			}
		}
		return nil
	}
}

func setEnvironmentVariables(resApp string, m map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resApp]
		if !ok {
			return fmt.Errorf("app '%s' not found in terraform state", resApp)
		}
		id := rs.Primary.ID

		_, _, err := session.BitsManager.UpdateAppEnvironment(id, m)
		return err
	}
}

func TestAccResApp_Routes_updateToAndmore(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.dummy-app"

	for _, app := range appPaths {
		appPath = app.path

		t.Run(fmt.Sprintf("AppSource=%s", app.typeOfPath), func(t *testing.T) {
			resource.Test(t,
				resource.TestCase{
					PreCheck:          func() { testAccPreCheck(t) },
					ProviderFactories: testAccProvidersFactories,
					CheckDestroy:      testAccCheckAppDestroyed([]string{"dummy-app"}),
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

									if err = assertHTTPResponse("https://dummy-app-tf."+defaultAppDomain(), 200, nil); err != nil {
										return err
									}
									return
								}),
								resource.TestCheckResourceAttr(refApp, "name", "dummy-app"),
								resource.TestCheckResourceAttr(refApp, "space", spaceID),
								resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
								resource.TestCheckResourceAttr(refApp, "ports.0", "8080"),
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

									if err = assertHTTPResponse("https://dummy-app-tf."+defaultAppDomain(), 200, nil); err != nil {
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
								resource.TestCheckResourceAttr(refApp, "ports.0", "8080"),
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

									if err = assertHTTPResponse("https://dummy-app-tf."+defaultAppDomain(), 200, nil); err != nil {
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
								resource.TestCheckResourceAttr(refApp, "ports.0", "8080"),
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

									if err = assertHTTPResponse("https://dummy-app-tf."+defaultAppDomain(), 404, nil); err != nil {
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
								resource.TestCheckResourceAttr(refApp, "ports.0", "8080"),
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
		})
	}
}

func TestAccResApp_dockerApp(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)

	refApp := "cloudfoundry_app.test-docker-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			CheckDestroy:      testAccCheckAppDestroyed([]string{"test-docker-app"}),
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
						// For docker apps, ports are not set (not supported in v3)
						resource.TestCheckResourceAttr(
							refApp, "ports.#", "0"),
						resource.TestCheckResourceAttr(
							refApp, "instances", "1"),
						// For docker apps, stack is ""
						resource.TestCheckResourceAttr(
							refApp, "stack", ""),
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

func TestAccResApp_dockerAppInvocationTimeout(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)

	refApp := "cloudfoundry_app.test-docker-app"
	invocationTimeout := 10
	resource.Test(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			CheckDestroy:      testAccCheckAppDestroyed([]string{"test-docker-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceDockerInvocationTimeout, defaultAppDomain(), orgName, spaceName, invocationTimeout),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://app_route_test_timeout."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(
							refApp, "name", "test-docker-app"),
						resource.TestCheckResourceAttr(
							refApp, "space", spaceID),
						// For docker apps, ports are not set (not supported in v3)
						resource.TestCheckResourceAttr(
							refApp, "ports.#", "0"),
						resource.TestCheckResourceAttr(
							refApp, "instances", "1"),
						// For docker apps, stack is ""
						resource.TestCheckResourceAttr(
							refApp, "stack", ""),
						resource.TestCheckResourceAttr(
							refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(
							refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(
							refApp, "docker_image", "cloudfoundry/diego-docker-app:latest"),
						resource.TestCheckResourceAttr(
							refApp, "health_check_invocation_timeout", fmt.Sprintf("%d", invocationTimeout)),
					),
				},
			},
		})
}

func TestAccResApp_app_bluegreen(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)

	refApp := "cloudfoundry_app.test-docker-app"
	invocationTimeout := 10
	resource.Test(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			CheckDestroy:      testAccCheckAppDestroyed([]string{"test-docker-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceDockerInvocationTimeout, defaultAppDomain(), orgName, spaceName, invocationTimeout),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://app_route_test_timeout."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(
							refApp, "name", "test-docker-app"),
						resource.TestCheckResourceAttr(
							refApp, "space", spaceID),
						// For docker apps, ports are not set (not supported in v3)
						resource.TestCheckResourceAttr(
							refApp, "ports.#", "0"),
						resource.TestCheckResourceAttr(
							refApp, "instances", "1"),
						// For docker apps, stack is ""
						resource.TestCheckResourceAttr(
							refApp, "stack", ""),
						resource.TestCheckResourceAttr(
							refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(
							refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(
							refApp, "docker_image", "cloudfoundry/diego-docker-app:latest"),
						resource.TestCheckResourceAttr(
							refApp, "health_check_invocation_timeout", fmt.Sprintf("%d", invocationTimeout)),
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

		// app, _, err := session.ClientV2.GetApplication(id)
		// if err != nil {
		// 	return err
		// }

		query := ccv3.Query{
			Key:    ccv3.GUIDFilter,
			Values: []string{id},
		}
		apps, _, err := session.ClientV3.GetApplications(query)
		if err != nil || len(apps) == 0 {
			return err
		}

		app := apps[0]

		// Get enabled_ssh
		enableSSH, _, err := session.ClientV3.GetAppFeature(app.GUID, "ssh")
		if err != nil {
			return err
		}

		// Get environment variables
		env, err := session.BitsManager.GetAppEnvironmentVariables(app.GUID)
		if err != nil {
			return err
		}

		// Get route mapping
		mappings, _, err := session.ClientV3.GetApplicationRoutes(app.GUID)
		if err != nil {
			return err
		}

		// Get service bindings
		bindings, _, err := session.ClientV3.GetServiceCredentialBindings(ccv3.Query{
			Key:    ccv3.AppGUIDFilter,
			Values: []string{app.GUID},
		})
		if err != nil {
			return err
		}

		// Fetch process information
		proc, _, err := session.ClientV3.GetApplicationProcessByType(app.GUID, constantV3.ProcessTypeWeb)
		// ProcessToResourceData(d, proc)

		if err = assertEquals(attributes, "name", app.Name); err != nil {
			return err
		}
		if err = assertEquals(attributes, "space", app.SpaceGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "instances", proc.Instances.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "memory", proc.MemoryInMB.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "disk_quota", proc.DiskInMB.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "stack", app.StackName); err != nil {
			return err
		}
		if app.LifecycleType == constantV3.AppLifecycleTypeBuildpack {
			if err = assertEquals(attributes, "buildpack", app.LifecycleBuildpacks[0]); err != nil {
				return err
			}
		}
		if err = assertEquals(attributes, "command", proc.Command.Value); err != nil {
			return err
		}
		if err = assertEquals(attributes, "enable_ssh", enableSSH.Enabled); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_http_endpoint", proc.HealthCheckEndpoint); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_type", proc.HealthCheckType); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_timeout", proc.HealthCheckTimeout); err != nil {
			return err
		}

		if err = assertMapEquals("environment", attributes, env); err != nil {
			return err
		}

		if err = assertListEquals(attributes, "service_binding", len(bindings),
			func(values map[string]string, i int) (match bool) {
				found := false
				for _, b := range bindings {
					if values["service_instance"] == b.ServiceInstanceGUID {
						found = true
						break
					}
				}
				return found

			}); err != nil {
			return err
		}

		routeMappings := make([]map[string]interface{}, 0)
		for _, mapping := range mappings {
			curMapping := make(map[string]interface{})
			curMapping["route"] = mapping.GUID
			curMapping["port"] = mapping.Port
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
	routesTf := GetListOfStructs(result.Value)
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

func TestAccMultipleBuildpacks(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)
	refApp := "cloudfoundry_app.app_1"

	resource.Test(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			CheckDestroy:      testAccCheckAppDestroyed([]string{"app-1"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(multipleBuildpacks, defaultAppDomain(), orgName, spaceName, appPath, "standard"),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {
							if err = assertHTTPResponse("https://app-1-tf."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "app-1"),
						resource.TestCheckResourceAttr(refApp, "buildpacks.#", "2"),
						resource.TestCheckResourceAttr(refApp, "buildpacks.0", "binary_buildpack"),
						resource.TestCheckResourceAttr(refApp, "buildpacks.1", "tomee_buildpack"),
					),
				},
			},
		})
}
