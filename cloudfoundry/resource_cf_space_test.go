package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const spaceResource = `

resource "cloudfoundry_asg" "svc" {
	name = "app-services"
    rule {
        protocol = "all"
        destination = "192.168.100.0/24"
    }
}
resource "cloudfoundry_asg" "stg1" {
	name = "app-services1"
    rule {
        protocol = "all"
        destination = "192.168.101.0/24"
    }
}
resource "cloudfoundry_asg" "stg2" {
	name = "app-services2"
    rule {
        protocol = "all"
        destination = "192.168.102.0/24"
    }
}
resource "cloudfoundry_asg" "stg3" {
	name = "app-services3"
    rule {
        protocol = "all"
        destination = "192.168.103.0/24"
    }
}
resource "cloudfoundry_user" "tl" {
	name = "teamlead@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "dev1" {
    name = "developer1@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "dev2" {
    name = "developer2@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "dev3" {
    name = "developer3@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "adr" {
    name = "auditor@acme.com"
	password = "password"
}
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
	asgs = [ "${cloudfoundry_asg.svc.id}" ]
	staging_asgs = [ "${cloudfoundry_asg.stg1.id}", "${cloudfoundry_asg.stg2.id}" ]
    managers = [
        "${cloudfoundry_user.tl.id}"
    ]
    developers = [
        "${cloudfoundry_user.tl.id}",
        "${cloudfoundry_user.dev1.id}",
		"${cloudfoundry_user.dev2.id}"
    ]
    auditors = [
        "${cloudfoundry_user.adr.id}",
		"${cloudfoundry_user.dev3.id}"
    ]
	allow_ssh = true
}
`

const spaceResourceUpdate = `

resource "cloudfoundry_asg" "svc" {
	name = "app-services"
    rule {
        protocol = "all"
        destination = "192.168.100.0/24"
    }
}
resource "cloudfoundry_asg" "stg1" {
	name = "app-services1"
    rule {
        protocol = "all"
        destination = "192.168.101.0/24"
    }
}
resource "cloudfoundry_asg" "stg2" {
	name = "app-services2"
    rule {
        protocol = "all"
        destination = "192.168.102.0/24"
    }
}
resource "cloudfoundry_asg" "stg3" {
	name = "app-services3"
    rule {
        protocol = "all"
        destination = "192.168.103.0/24"
    }
}
resource "cloudfoundry_user" "tl" {
    name = "teamlead@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "dev1" {
    name = "developer1@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "dev2" {
    name = "developer2@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "dev3" {
    name = "developer3@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "adr" {
    name = "auditor@acme.com"
	password = "password"
}
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

resource "cloudfoundry_space_quota" "dev2" {
	name = "50g-dev2"
	org = "${cloudfoundry_org.org1.id}"
    allow_paid_service_plans = true
    instance_memory = 1024
    total_memory = 51200
    total_app_instances = 100
    total_routes = 100
    total_services = 150
}

resource "cloudfoundry_space" "space1" {
	name = "space-one-updated"
	org = "${cloudfoundry_org.org1.id}"
	quota = "${cloudfoundry_space_quota.dev2.id}"
	asgs = [ "${cloudfoundry_asg.svc.id}" ]
	staging_asgs = [ "${cloudfoundry_asg.stg2.id}", "${cloudfoundry_asg.stg3.id}" ]
    managers = [
        "${cloudfoundry_user.tl.id}"
    ]
    developers = [
        "${cloudfoundry_user.tl.id}",
        "${cloudfoundry_user.dev1.id}",
    ]
    auditors = [
        "${cloudfoundry_user.adr.id}",
		"${cloudfoundry_user.dev2.id}"
    ]
	allow_ssh = true
}
`

func TestAccResSpace_normal(t *testing.T) {

	ref := "cloudfoundry_space.space1"
	refUserRemoved := "cloudfoundry_user.dev3"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSpaceDestroyed("space-one"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: spaceResource,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSpaceExists(ref, nil),
						resource.TestCheckResourceAttr(
							ref, "name", "space-one"),
						resource.TestCheckResourceAttr(
							ref, "asgs.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "staging_asgs.#", "2"),
						resource.TestCheckResourceAttr(
							ref, "managers.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "developers.#", "3"),
						resource.TestCheckResourceAttr(
							ref, "auditors.#", "2"),
					),
				},

				resource.TestStep{
					Config: spaceResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSpaceExists(ref, &refUserRemoved),
						resource.TestCheckResourceAttr(
							ref, "name", "space-one-updated"),
						resource.TestCheckResourceAttr(
							ref, "asgs.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "staging_asgs.#", "2"),
						resource.TestCheckResourceAttr(
							ref, "managers.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "developers.#", "2"),
						resource.TestCheckResourceAttr(
							ref, "auditors.#", "2"),
					),
				},
			},
		})
}

func testAccCheckSpaceExists(resource string, refUserRemoved *string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("quota '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		sm := session.ClientV2
		space, _, err := sm.GetSpace(id)
		if err != nil {
			return err
		}

		if err = assertEquals(attributes, "name", space.Name); err != nil {
			return err
		}
		if err = assertEquals(attributes, "org", space.OrganizationGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "quota", space.SpaceQuotaDefinitionGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "allow_ssh", strconv.FormatBool(space.AllowSSH)); err != nil {
			return err
		}

		for t, r := range typeToSpaceRoleMap {
			users, _, err := session.ClientV2.GetSpaceUsersByRole(r, id)
			if err != nil {
				return err
			}
			if err = assertSetEquals(attributes, t, objectsToIds(users, func(object interface{}) string {
				return object.(ccv2.User).GUID
			})); err != nil {
				return err
			}
		}

		err = testUserRemovedFromOrg(refUserRemoved, space.GUID, session, s)

		return nil
	}
}

func testAccCheckSpaceDestroyed(spacename string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)
		spaces, _, err := session.ClientV2.GetSpaces(ccv2.FilterByName(spacename))
		if err != nil {
			return err
		}
		if len(spaces) > 0 {
			return fmt.Errorf("space with name '%s' still exists in cloud foundry", spacename)
		}
		return nil
	}
}
