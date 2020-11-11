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

const userProvidedServiceResourceCreate = `
resource "cloudfoundry_org" "org1" {
	name = "organization-one"
}

resource "cloudfoundry_space_quota" "dev" {
	name = "50g"
	org = "${cloudfoundry_org.org1.id}"
  allow_paid_service_plans = true
  instance_memory = 1024
  total_memory = 51200
  total_app_instances = 100
  total_routes = 100
  total_services = 150
}

resource "cloudfoundry_space" "space1" {
	name = "space-one"
	org = "${cloudfoundry_org.org1.id}"
	quota = "${cloudfoundry_space_quota.dev.id}"
	allow_ssh = true
}

resource "cloudfoundry_user_provided_service" "mq" {
	name = "mq"
  space = "${cloudfoundry_space.space1.id}"
  credentials = {
		"url" = "mq://localhost:9000"
		"username" = "user"
		"password" = "pwd"
	}
}
`

const userProvidedServiceComplexResourceCreate = `
resource "cloudfoundry_org" "org1" {
	name = "organization-one"
}

resource "cloudfoundry_space" "space1" {
	name = "space-one"
	org = "${cloudfoundry_org.org1.id}"
	allow_ssh = true
}

resource "cloudfoundry_user_provided_service" "complex" {
	name = "complex"
  space = "${cloudfoundry_space.space1.id}"
  credentials_json = "{ \"cnx\": { \"host\": \"localhost\", \"ports\": [ 8080, 8081, 8082 ] } }"
}
`

const userProvidedServiceResourceUpdate = `
resource "cloudfoundry_org" "org1" {
  name = "organization-one"
}

resource "cloudfoundry_space_quota" "dev" {
  name = "50g"
  org = "${cloudfoundry_org.org1.id}"
  allow_paid_service_plans = true
  instance_memory = 1024
  total_memory = 51200
  total_app_instances = 100
  total_routes = 100
  total_services = 150
}

resource "cloudfoundry_space" "space1" {
  name = "space-one"
  org = "${cloudfoundry_org.org1.id}"
  quota = "${cloudfoundry_space_quota.dev.id}"
  allow_ssh = true
}

resource "cloudfoundry_user_provided_service" "mq" {
  name = "mq"
  space = "${cloudfoundry_space.space1.id}"
  credentials = {
    "url" = "mq://localhost:9000"
    "username" = "new-user"
    "password" = "new-pwd"
  }
  syslog_drain_url = "http://localhost/syslog"
  route_service_url = "https://localhost/route"
}
`

const userProvidedServiceComplexResourceUpdate = `
resource "cloudfoundry_org" "org1" {
  name = "organization-one"
}

resource "cloudfoundry_space" "space1" {
  name = "space-one"
  org = "${cloudfoundry_org.org1.id}"
  allow_ssh = true
}

resource "cloudfoundry_user_provided_service" "complex" {
	name = "complex"
  space = "${cloudfoundry_space.space1.id}"
  credentials_json = "{ \"cnx\": { \"host\": \"127.0.0.1\", \"ports\": [ 8088 ] } }"
  syslog_drain_url = "http://localhost/syslog"
  route_service_url = "https://localhost/route"
  tags = ["mytag"]
}
`

func TestAccResUserProvidedService_normal(t *testing.T) {

	ref := "cloudfoundry_user_provided_service.mq"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserProvidedServiceDestroyed("mq", "cloudfoundry_space.space1"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: userProvidedServiceResourceCreate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserProvidedServiceExists(ref),
						resource.TestCheckResourceAttr(ref, "name", "mq"),
						resource.TestCheckResourceAttr(ref, "credentials.url", "mq://localhost:9000"),
						resource.TestCheckResourceAttr(ref, "credentials.username", "user"),
						resource.TestCheckResourceAttr(ref, "credentials.password", "pwd"),
						resource.TestCheckNoResourceAttr(ref, "syslog_drain_url"),
						resource.TestCheckNoResourceAttr(ref, "route_service_url"),
						resource.TestCheckNoResourceAttr(ref, "credentials_json"),
					),
				},

				resource.TestStep{
					Config: userProvidedServiceResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserProvidedServiceExists(ref),
						resource.TestCheckResourceAttr(ref, "name", "mq"),
						resource.TestCheckResourceAttr(ref, "credentials.url", "mq://localhost:9000"),
						resource.TestCheckResourceAttr(ref, "credentials.username", "new-user"),
						resource.TestCheckResourceAttr(ref, "credentials.password", "new-pwd"),
						resource.TestCheckResourceAttr(ref, "syslog_drain_url", "http://localhost/syslog"),
						resource.TestCheckResourceAttr(ref, "route_service_url", "https://localhost/route"),
						resource.TestCheckNoResourceAttr(ref, "credentials_json"),
					),
				},
			},
		})
}

func TestAccResUserProvidedService_complex(t *testing.T) {
	ref := "cloudfoundry_user_provided_service.complex"
	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserProvidedServiceDestroyed("complex", "cloudfoundry_space.space1"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: userProvidedServiceComplexResourceCreate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserProvidedServiceExists(ref),
						resource.TestCheckResourceAttr(ref, "name", "complex"),
						resource.TestCheckResourceAttr(ref, "credentials_json", `{ "cnx": { "host": "localhost", "ports": [ 8080, 8081, 8082 ] } }`),
						resource.TestCheckNoResourceAttr(ref, "syslog_drain_url"),
						resource.TestCheckNoResourceAttr(ref, "route_service_url"),
						resource.TestCheckNoResourceAttr(ref, "credentials"),
					),
				},

				resource.TestStep{
					Config: userProvidedServiceComplexResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserProvidedServiceExists(ref),
						resource.TestCheckResourceAttr(ref, "name", "complex"),
						resource.TestCheckResourceAttr(ref, "credentials_json", `{ "cnx": { "host": "127.0.0.1", "ports": [ 8088 ] } }`),
						resource.TestCheckResourceAttr(ref, "syslog_drain_url", "http://localhost/syslog"),
						resource.TestCheckResourceAttr(ref, "route_service_url", "https://localhost/route"),
						resource.TestCheckNoResourceAttr(ref, "credentials"),
						resource.TestCheckResourceAttr(ref, "tags.#", "1"),
					),
				},
			},
		})
}

func testAccCheckUserProvidedServiceExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("user provided service '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID

		_, _, err := session.ClientV2.GetUserProvidedServiceInstance(id)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckUserProvidedServiceDestroyed(name string, spaceResource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[spaceResource]
		if !ok {
			return fmt.Errorf("space '%s' not found in terraform state", spaceResource)
		}
		svcs, _, err := session.ClientV2.GetServiceInstances(ccv2.FilterByName(name), ccv2.FilterEqual(constant.SpaceGUIDFilter, rs.Primary.ID))
		if err != nil {
			return err
		}
		if len(svcs) > 0 {
			return fmt.Errorf("user provided service with name '%s' still exists in cloud foundry", name)
		}
		return nil
	}
}
