package cloudfoundry

import (
	"fmt"
	"strconv"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
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
	org_id = "${cloudfoundry_org.org1.id}"
    allow_paid_service_plans = true
    instance_memory = 1024
    total_memory = 51200
    total_app_instances = 100
    total_routes = 100
    total_services = 150
}

resource "cloudfoundry_space" "space1" {
	name = "space-one"
	org_id = "${cloudfoundry_org.org1.id}"
	quota_id = "${cloudfoundry_space_quota.dev.id}"
	asg_ids = [ "${cloudfoundry_asg.svc.id}" ]
	staging_asg_ids = [ "${cloudfoundry_asg.stg1.id}", "${cloudfoundry_asg.stg2.id}" ]
    manager_ids = [
        "${cloudfoundry_user.tl.id}"
    ]
    developer_ids = [
        "${cloudfoundry_user.tl.id}",
        "${cloudfoundry_user.dev1.id}",
		"${cloudfoundry_user.dev2.id}"
    ]
    auditor_ids = [
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
	org_id = "${cloudfoundry_org.org1.id}"
    allow_paid_service_plans = true
    instance_memory = 1024
    total_memory = 51200
    total_app_instances = 100
    total_routes = 100
    total_services = 150
}

resource "cloudfoundry_space" "space1" {
	name = "space-one-updated"
	org_id = "${cloudfoundry_org.org1.id}"
	quota_id = "${cloudfoundry_space_quota.dev.id}"
	asg_ids = [ "${cloudfoundry_asg.svc.id}" ]
	staging_asg_ids = [ "${cloudfoundry_asg.stg2.id}", "${cloudfoundry_asg.stg3.id}" ]
    manager_ids = [
        "${cloudfoundry_user.tl.id}"
    ]
    developer_ids = [
        "${cloudfoundry_user.tl.id}",
        "${cloudfoundry_user.dev1.id}",
    ]
    auditor_ids = [
        "${cloudfoundry_user.adr.id}",
		"${cloudfoundry_user.dev2.id}"
    ]
	allow_ssh = true
}
`

func TestAccSpace_normal(t *testing.T) {

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
							ref, "asg_ids.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "manager_ids.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "developer_ids.#", "3"),
						resource.TestCheckResourceAttr(
							ref, "auditor_ids.#", "2"),
					),
				},

				resource.TestStep{
					Config: spaceResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSpaceExists(ref, &refUserRemoved),
						resource.TestCheckResourceAttr(
							ref, "name", "space-one-updated"),
						resource.TestCheckResourceAttr(
							ref, "asg_ids.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "manager_ids.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "developer_ids.#", "2"),
						resource.TestCheckResourceAttr(
							ref, "auditor_ids.#", "2"),
					),
				},
			},
		})
}

func testAccCheckSpaceExists(resource string, refUserRemoved *string) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("quota '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var (
			space cfapi.CCSpace

			runningAsgIDs, stagingAsgIDs       []string
			spaceAsgIDs                      []interface{}
			managerIDs, developerIDs, auditorIDs []interface{}
		)

		sm := session.SpaceManager()
		if space, err = sm.ReadSpace(id); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved space for resource '%s' with id '%s': %# v",
			resource, id, space)

		if err = assertEquals(attributes, "name", space.Name); err != nil {
			return err
		}
		if err = assertEquals(attributes, "org", space.OrgGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "quota", space.QuotaGUID); err != nil {
			return err
		}
		if err = assertEquals(attributes, "allow_ssh", strconv.FormatBool(space.AllowSSH)); err != nil {
			return err
		}

		if runningAsgIDs, err = session.ASGManager().Running(); err != nil {
			return err
		}
		if spaceAsgIDs, err = sm.ListASGs(id); err != nil {
			return
		}
		asgIDs := []interface{}{}
		for _, a := range spaceAsgIDs {
			if !isStringInList(runningAsgIDs, a.(string)) {
				asgIDs = append(asgIDs, a)
			}
		}
		session.Log.DebugMessage(
			"retrieved asgs of space identified resource '%s': %# v",
			resource, asgIDs)

		if err = assertSetEquals(attributes, "asg_ids", asgIDs); err != nil {
			return err
		}

		if stagingAsgIDs, err = session.ASGManager().Staging(); err != nil {
			return err
		}
		if spaceAsgIDs, err = sm.ListStagingASGs(id); err != nil {
			return
		}
		asgIDs = []interface{}{}
		for _, a := range spaceAsgIDs {
			if !isStringInList(stagingAsgIDs, a.(string)) {
				asgIDs = append(asgIDs, a)
			}
		}
		session.Log.DebugMessage(
			"retrieved staging asgs of space identified resource '%s': %# v",
			resource, asgIDs)

		if err = assertSetEquals(attributes, "staging_asg_ids", asgIDs); err != nil {
			return err
		}

		if managerIDs, err = sm.ListUsers(id, cfapi.SpaceRoleManager); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved managers of space identified resource '%s': %# v",
			resource, managerIDs)

		if err = assertSetEquals(attributes, "manager_ids", managerIDs); err != nil {
			return err
		}

		if developerIDs, err = sm.ListUsers(id, cfapi.SpaceRoleDeveloper); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved developers of space identified resource '%s': %# v",
			resource, developerIDs)

		if err = assertSetEquals(attributes, "developer_ids", developerIDs); err != nil {
			return err
		}

		if auditorIDs, err = sm.ListUsers(id, cfapi.SpaceRoleAuditor); err != nil {
			return
		}
		session.Log.DebugMessage(
			"retrieved managers of space identified resource '%s': %# v",
			resource, auditorIDs)

		if err = assertSetEquals(attributes, "auditor_ids", auditorIDs); err != nil {
			return err
		}

		err = testUserRemovedFromOrg(refUserRemoved, space.OrgGUID, session.OrgManager(), s)

		return
	}
}

func testAccCheckSpaceDestroyed(spacename string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)
		if _, err := session.SpaceManager().FindSpace(spacename); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}
		return fmt.Errorf("space with name '%s' still exists in cloud foundry", spacename)
	}
}
