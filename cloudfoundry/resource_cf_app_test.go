package cloudfoundry

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

const appResourceUrlDockerTemplate = `

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
resource "cloudfoundry_route" "java-spring" {
	domain = "${data.cloudfoundry_domain.local.id}"
	space = "${data.cloudfoundry_space.space.id}"
	hostname = "java-spring"
}
resource "cloudfoundry_app" "java-spring" {
	name = "java-spring"
	space = "${data.cloudfoundry_space.space.id}"
	memory = "768"
	disk_quota = "512"
	timeout = 1800
	instances = 1
	routes {
		route = "${cloudfoundry_route.java-spring.id}"
	}

%%s
}
`

const appResourceJavaSpringTemplate = `

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

resource "cloudfoundry_route" "java-spring" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "java-spring"
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
resource "cloudfoundry_app" "java-spring" {
  name = "java-spring"
  space = "${data.cloudfoundry_space.space.id}"
  memory = "768"
  disk_quota = "512"
  timeout = 1800

	url = "file://../tests/cf-acceptance-tests/assets/java-spring/java-spring.jar"
	
%%s
}
`

const appResourceJavaSpring = `

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

resource "cloudfoundry_route" "java-spring" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "java-spring"
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
resource "cloudfoundry_app" "java-spring" {
  name = "java-spring"
  space = "${data.cloudfoundry_space.space.id}"
  memory = "768"
  disk_quota = "512"
  timeout = 1800

  url = "file://../tests/cf-acceptance-tests/assets/java-spring/java-spring.jar"

  service_binding {
    service_instance = "${cloudfoundry_service_instance.db.id}"
  }
  service_binding {
    service_instance = "${cloudfoundry_service_instance.fs1.id}"
  }

  route {
    default_route = "${cloudfoundry_route.java-spring.id}"
  }

  environment {
    TEST_VAR_1 = "testval1"
    TEST_VAR_2 = "testval2"
  }
}
`

const appResourceJavaSpringUpdate = `

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

resource "cloudfoundry_route" "java-spring" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "java-spring"
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
resource "cloudfoundry_app" "java-spring" {
  name = "java-spring-updated"
  space = "${data.cloudfoundry_space.space.id}"
  instances ="2"
  memory = "1024"
  disk_quota = "1024"
  timeout = 1800

  url = "file://../tests/cf-acceptance-tests/assets/java-spring/java-spring.jar"

  service_binding {
    service_instance = "${cloudfoundry_service_instance.db.id}"
  }
  service_binding {
    service_instance = "${cloudfoundry_service_instance.fs2.id}"
  }
  service_binding {
    service_instance = "${cloudfoundry_service_instance.fs1.id}"
  }

  route {
    default_route = "${cloudfoundry_route.java-spring.id}"
  }

  environment {
    TEST_VAR_1 = "testval1"
    TEST_VAR_2 = "testval2"
  }
}
`

const appResourceWithMultiplePorts = `

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

resource "cloudfoundry_app" "test-app" {
  name = "test-app"
  space = "${data.cloudfoundry_space.space.id}"
  timeout = 1800
  ports = [ 8888, 9999 ]
  buildpack = "binary_buildpack"
  command = "chmod 0755 test-app && ./test-app --ports=8888,9999"
  health_check_type = "process"

  github_release {
    owner = "mevansam"
    repo = "test-app"
    filename = "test-app"
    version = "v0.0.1"
    user = "%s"
    password = "%s"
  }
}
resource "cloudfoundry_route" "test-app-8888" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "test-app-8888"

  target {
    app = "${cloudfoundry_app.test-app.id}"
    port = 8888
  }
}
resource "cloudfoundry_route" "test-app-9999" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "test-app-9999"

  target {
    app = "${cloudfoundry_app.test-app.id}"
    port = 9999
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
  target {
    app = "${cloudfoundry_app.test-docker-app.id}"
    port = 8080
  }
}
resource "cloudfoundry_app" "test-docker-app" {
  name = "test-docker-app"
  space = "${data.cloudfoundry_space.space.id}"
  docker_image = "cloudfoundry/diego-docker-app:latest"
  timeout = 900
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
	target = {app = "${cloudfoundry_app.test-app.id}"}
}
resource "cloudfoundry_app" "test-app" {
  name = "test-app"
  space = "${data.cloudfoundry_space.space.id}"
  command = "test-app --ports=8080"
  timeout = 1800
	memory = "512"

  git {
    url = "https://github.com/mevansam/test-app.git"
  }
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
	target = {app = "${cloudfoundry_app.test-app.id}"}
}
resource "cloudfoundry_app" "test-app" {
  name = "test-app"
  space = "${data.cloudfoundry_space.space.id}"
  command = "test-app --ports=8080"
  timeout = 1800
	memory = "1024"

  git {
    url = "https://github.com/mevansam/test-app.git"
  }
}
`

const createManyJavaSpringApps = `

data "cloudfoundry_domain" "java-spring-domain" {
  name = "%s"
}

data "cloudfoundry_org" "org" {
  name = "%s"
}
data "cloudfoundry_space" "space" {
  name = "%s"
  org = "${data.cloudfoundry_org.org.id}"
}

resource "cloudfoundry_route" "java-spring-route-1" {
  domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
  space = "${data.cloudfoundry_space.space.id}"
  hostname = "java-spring-1"
  depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-1" {
  name = "java-spring-app-1"
  url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
  space = "${data.cloudfoundry_space.space.id}"
  timeout = 700
  memory = 512
  buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

  route {
    default_route = "${cloudfoundry_route.java-spring-route-1.id}"
  }

  depends_on = ["cloudfoundry_route.java-spring-route-1"]
}

resource "cloudfoundry_route" "java-spring-route-2" {
  domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
	space = "${data.cloudfoundry_space.space.id}"
  hostname = "java-spring-2"
  depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-2" {
	name = "java-spring-app-2"
  url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
  space = "${data.cloudfoundry_space.space.id}"
  timeout = 700
	memory = 512
	buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

  route {
    default_route = "${cloudfoundry_route.java-spring-route-2.id}"
  }

  depends_on = ["cloudfoundry_route.java-spring-route-2"]
}

resource "cloudfoundry_route" "java-spring-route-3" {
  domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
	space = "${data.cloudfoundry_space.space.id}"
  hostname = "java-spring-3"
  depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-3" {
	name = "java-spring-app-3"
  url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
  space = "${data.cloudfoundry_space.space.id}"
  timeout = 700
	memory = 512
	buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

  route {
    default_route = "${cloudfoundry_route.java-spring-route-3.id}"
  }

  depends_on = ["cloudfoundry_route.java-spring-route-3"]
}

resource "cloudfoundry_route" "java-spring-route-4" {
  domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
	space = "${data.cloudfoundry_space.space.id}"
  hostname = "java-spring-4"
  depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-4" {
	name = "java-spring-app-4"
  url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
  space = "${data.cloudfoundry_space.space.id}"
  timeout = 700
	memory = 512
	buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

  route {
    default_route = "${cloudfoundry_route.java-spring-route-4.id}"
  }

  depends_on = ["cloudfoundry_route.java-spring-route-4"]
}

resource "cloudfoundry_route" "java-spring-route-5" {
  domain = "${data.cloudfoundry_domain.java-spring-domain.id}"
	space = "${data.cloudfoundry_space.space.id}"
  hostname = "java-spring-5"
  depends_on = ["data.cloudfoundry_domain.java-spring-domain"]
}

resource "cloudfoundry_app" "java-spring-app-5" {
	name = "java-spring-app-5"
  url = "file://../tests/cf-acceptance-tests/assets/java-spring/"
  space = "${data.cloudfoundry_space.space.id}"
  timeout = 700
	memory = 512
	buildpack = "https://github.com/cloudfoundry/java-buildpack.git"

  route {
    default_route = "${cloudfoundry_route.java-spring-route-5.id}"
  }

  depends_on = ["cloudfoundry_route.java-spring-route-5"]
}
`

// If the PR is not applied, after running this test many times, it should crash with this error
// === RUN   TestAccApp_reproduceIssue88
// Application downloaded to: ../tests/cf-acceptance-tests/assets/java-spring/
// Application downloaded to: ../tests/cf-acceptance-tests/assets/java-spring/
// fatal error: concurrent map read and map write
//
// goroutine 1542 [running]:
// ...
// created by github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry.resourceAppCreate
// .../golang/src/github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/resource_cf_app.go:421 +0x1ac4

func TestAccApp_reproduceIssue88(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)

	refApp1 := "cloudfoundry_app.java-spring-app-1"
	refApp2 := "cloudfoundry_app.java-spring-app-2"
	refApp3 := "cloudfoundry_app.java-spring-app-3"
	refApp4 := "cloudfoundry_app.java-spring-app-4"
	refApp5 := "cloudfoundry_app.java-spring-app-5"

	failRegExp, _ := regexp.Compile("app java-spring-app-[0-9] failed to start")

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring-app-`", "java-spring-app-2", "java-spring-app-3", "java-spring-app-4", "java-spring-app-5"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(createManyJavaSpringApps,
						defaultAppDomain(),
						orgName, spaceName),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccCheckAppExists(refApp1, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-1."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						testAccCheckAppExists(refApp2, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						testAccCheckAppExists(refApp3, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-3."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						testAccCheckAppExists(refApp4, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-4."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						testAccCheckAppExists(refApp5, func() (err error) {

							if err = assertHTTPResponse("https://java-spring-5."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
					),
					// the jar in the test is enough big, and allows us to test for the failure
					ExpectError: failRegExp,
				},
			},
		})
}

func TestAccAppVersions_app1(t *testing.T) {

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
					Config: fmt.Sprintf(multipleVersion, defaultAppDomain(), orgName, spaceName),
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
					Config: fmt.Sprintf(multipleVersionUpdate, defaultAppDomain(), orgName, spaceName),
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

func TestAccApp_app1(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.java-spring"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceJavaSpring,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
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
					Config: fmt.Sprintf(appResourceJavaSpringUpdate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan, servicePlan),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring-updated"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "2"),
						resource.TestCheckResourceAttr(refApp, "memory", "1024"),
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
func TestAccApp_app2(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)

	refApp := "cloudfoundry_app.test-app"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"test-app"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(appResourceWithMultiplePorts,
						defaultAppDomain(),
						orgName, spaceName,
						os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_TOKEN")),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {
							responses := []string{"8888"}
							if err = assertHTTPResponse("https://test-app-8888."+defaultAppDomain()+"/port", 200, &responses); err != nil {
								return err
							}
							responses = []string{"9999"}
							if err = assertHTTPResponse("https://test-app-9999."+defaultAppDomain()+"/port", 200, &responses); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "test-app"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "2"),
						resource.TestCheckResourceAttr(refApp, "ports.8888", "8888"),
						resource.TestCheckResourceAttr(refApp, "ports.9999", "9999"),
					),
				},
			},
		})
}

func TestApp_OldStyleRoutes_failLiveStage(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	resource.Test(t,
		resource.TestCase{
			IsUnitTest:   true,
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("\\[REMOVED\\] Support for the non-default route has been removed."),
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						``,
						`route {
              live_route = "${cloudfoundry_route.java-spring.id}"
            }`,
					),
				},

				resource.TestStep{
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("\\[REMOVED\\] Support for the non-default route has been removed."),
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						``,
						`route {
              stage_route = "${cloudfoundry_route.java-spring.id}"
            }`,
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_updateTo(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.java-spring"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						``,
						`route {
              default_route = "${cloudfoundry_route.java-spring.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "1"),
						resource.TestCheckNoResourceAttr(refApp, "routes.#"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(), orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						``,
						`routes {
              route = "${cloudfoundry_route.java-spring.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "0"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_updateToAndmore(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.java-spring"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						``,
						`route {
              default_route = "${cloudfoundry_route.java-spring.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "1"),
						resource.TestCheckNoResourceAttr(refApp, "routes.#"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						`resource "cloudfoundry_route" "java-spring-2" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "java-spring-2"
            }`,
						`routes {
              route = "${cloudfoundry_route.java-spring.id}"
            }
            routes {
              route = "${cloudfoundry_route.java-spring-2.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://java-spring-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "0"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "2"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						`resource "cloudfoundry_route" "java-spring-2" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "java-spring-2"
            }`,
						`routes {
              route = "${cloudfoundry_route.java-spring.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://java-spring-2."+defaultAppDomain(), 404, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "0"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						`resource "cloudfoundry_route" "java-spring-2" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "java-spring-2"
            }
            resource "cloudfoundry_route" "java-spring-3" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "java-spring-3"
            }`,
						`routes {
              route = "${cloudfoundry_route.java-spring-2.id}"
            }
            routes {
              route = "${cloudfoundry_route.java-spring-3.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 404, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://java-spring-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://java-spring-3."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckResourceAttr(refApp, "route.#", "0"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "2"),
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_Create(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.java-spring"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						``,
						`routes {
              route = "${cloudfoundry_route.java-spring.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_Change(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.java-spring"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						``,
						`routes {
              route = "${cloudfoundry_route.java-spring.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						`resource "cloudfoundry_route" "java-spring-2" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "java-spring-2"
            }`,
						`routes {
              route = "${cloudfoundry_route.java-spring-2.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 404, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://java-spring-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},
			},
		})
}

func TestAccApp_NewStyleRoutes_Add(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	serviceName1, serviceName2, servicePlan := getTestServiceBrokers(t)

	refApp := "cloudfoundry_app.java-spring"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						``,
						`routes {
              route = "${cloudfoundry_route.java-spring.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceJavaSpringTemplate,
						defaultAppDomain(),
						orgName, spaceName,
						serviceName1, serviceName2, servicePlan, servicePlan),
						`resource "cloudfoundry_route" "java-spring-2" {
              domain = "${data.cloudfoundry_domain.local.id}"
              space = "${data.cloudfoundry_space.space.id}"
              hostname = "java-spring-2"
            }`,
						`routes {
              route = "${cloudfoundry_route.java-spring.id}"
            }
            routes {
              route = "${cloudfoundry_route.java-spring-2.id}"
            }`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							if err = assertHTTPResponse("https://java-spring-2."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "service_binding.#"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "2"),
					),
				},
			},
		})
}

func TestAccApp_dockerApp(t *testing.T) {

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
							refApp, "ports.8080", "8080"),
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

func TestAccApp_url_docker_switch(t *testing.T) {

	_, orgName := defaultTestOrg(t)
	spaceID, spaceName := defaultTestSpace(t)
	refApp := "cloudfoundry_app.java-spring"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAppDestroyed([]string{"java-spring"}),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceUrlDockerTemplate, defaultAppDomain(), orgName, spaceName),
						`docker_image = "cloudfoundry/diego-docker-app:latest"`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(fmt.Sprintf(appResourceUrlDockerTemplate, defaultAppDomain()),
						`url = "file://../tests/cf-acceptance-tests/assets/java-spring/java-spring.jar"`,
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAppExists(refApp, func() (err error) {

							if err = assertHTTPResponse("https://java-spring."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
						resource.TestCheckResourceAttr(refApp, "name", "java-spring"),
						resource.TestCheckResourceAttr(refApp, "space", spaceID),
						resource.TestCheckResourceAttr(refApp, "ports.#", "1"),
						resource.TestCheckResourceAttr(refApp, "ports.8080", "8080"),
						resource.TestCheckResourceAttr(refApp, "instances", "1"),
						resource.TestCheckResourceAttr(refApp, "memory", "768"),
						resource.TestCheckResourceAttr(refApp, "disk_quota", "512"),
						resource.TestCheckResourceAttrSet(refApp, "stack"),
						resource.TestCheckResourceAttr(refApp, "environment.%", "0"),
						resource.TestCheckResourceAttr(refApp, "enable_ssh", "true"),
						resource.TestCheckResourceAttr(refApp, "health_check_type", "port"),
						resource.TestCheckNoResourceAttr(refApp, "route.#"),
						resource.TestCheckResourceAttr(refApp, "routes.#", "1"),
					),
				},
			},
		})
}

func testAccCheckAppExists(resApp string, validate func() error) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resApp]
		if !ok {
			return fmt.Errorf("app '%s' not found in terraform state", resApp)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resApp, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var (
			app             cfapi.CCApp
			routeMappings   []map[string]interface{}
			serviceBindings []map[string]interface{}
		)

		am := session.AppManager()
		rm := session.RouteManager()

		if app, err = am.ReadApp(id); err != nil {
			return err
		}
		session.Log.DebugMessage(
			"retrieved app for resource '%s' with id '%s': %# v",
			resApp, id, app)

		if err = assertEquals(attributes, "name", app.Name); err != nil {
			return err
		}
		if err = assertEquals(attributes, "space", app.SpaceGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "instances", app.Instances); err != nil {
			return err
		}
		if err = assertEquals(attributes, "memory", app.Memory); err != nil {
			return err
		}
		if err = assertEquals(attributes, "disk_quota", app.DiskQuota); err != nil {
			return err
		}
		if err = assertEquals(attributes, "stack", app.StackGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "buildpack", app.Buildpack); err != nil {
			return err
		}
		if err = assertEquals(attributes, "command", app.Command); err != nil {
			return err
		}
		if err = assertEquals(attributes, "enable_ssh", app.EnableSSH); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_http_endpoint", app.HealthCheckHTTPEndpoint); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_type", app.HealthCheckType); err != nil {
			return err
		}
		if err = assertEquals(attributes, "health_check_timeout", app.HealthCheckTimeout); err != nil {
			return err
		}
		if err = assertMapEquals("environment", attributes, *app.Environment); err != nil {
			return err
		}

		if serviceBindings, err = am.ReadServiceBindingsByApp(id); err != nil {
			return err
		}
		session.Log.DebugMessage(
			"retrieved service bindings for app with id '%s': %# v",
			id, serviceBindings)

		if err = assertListEquals(attributes, "service_binding", len(serviceBindings),
			func(values map[string]string, i int) (match bool) {
				var binding map[string]interface{}

				serviceInstanceID := values["service_instance"]
				binding = nil

				for _, b := range serviceBindings {
					if serviceInstanceID == b["service_instance"] {
						binding = b
						break
					}
				}

				if binding != nil && values["binding_id"] == binding["binding_id"] {
					return true
				}
				return false

			}); err != nil {
			return err
		}

		if routeMappings, err = rm.ReadRouteMappingsByApp(id); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved routes for app with id '%s': %# v",
			id, routeMappings)

		if err = validateRouteMappings(attributes, routeMappings); err != nil {
			return
		}

		err = validate()
		return
	}
}

func validateRouteMappings(attributes map[string]string, routeMappings []map[string]interface{}) (err error) {

	var (
		routeID, mappingID string
		mapping            map[string]interface{}

		ok bool
	)

	if _, isOldStyle := attributes["route.0.default_route"]; isOldStyle {
		routeKey := "route.0.default_route"
		routeMappingKey := "route.0.default_route_mapping_id"

		if routeID, ok = attributes[routeKey]; ok && len(routeID) > 0 {
			if mappingID, ok = attributes[routeMappingKey]; !ok || len(mappingID) == 0 {
				return fmt.Errorf("default route '%s' does not have a corresponding mapping id in the state", routeID)
			}

			mapping = nil
			for _, r := range routeMappings {
				if mappingID == r["mapping_id"] {
					mapping = r
					break
				}
			}
			if mapping == nil {
				return fmt.Errorf("unable to find route mapping with id '%s' for route '%s'", mappingID, routeID)
			}
			if routeID != mapping["route"] {
				return fmt.Errorf("route mapping with id '%s' does not map to route '%s'", mappingID, routeID)
			}
		}
		return err
	} else if _, isNewStyle := attributes["routes.0.route"]; isNewStyle {

		for i := 0; true; i++ {
			if routeID, ok := attributes[fmt.Sprintf("routes.%d.route", i)]; !ok {
				break
			} else {
				if mappingID, ok := attributes[fmt.Sprintf("routes.%d.mapping_id", i)]; !ok {
					return fmt.Errorf("Route with no mapping ID recored (routes.%d.route=%s)", i, routeID)
				} else {
					for _, r := range routeMappings {
						if mappingID == r["mapping_id"] {
							mapping = r
							break
						}
					}
					if mapping == nil {
						return fmt.Errorf("unable to find route mapping with id '%s' for route '%s'", mappingID, routeID)
					}
					if routeID != mapping["route"] {
						return fmt.Errorf("route mapping with id '%s' does not map to route '%s'", mappingID, routeID)
					}
				}
			}
		}
	}
	return nil
}

func testAccCheckAppDestroyed(apps []string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)
		for _, a := range apps {
			if _, err := session.AppManager().FindApp(a); err != nil {
				switch err.(type) {
				case *errors.ModelNotFoundError:
					continue
				default:
					return err
				}
			}
			return fmt.Errorf("app with name '%s' still exists in cloud foundry", a)
		}
		return nil
	}
}
