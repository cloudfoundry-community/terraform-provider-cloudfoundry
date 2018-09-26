package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const orgResource = `

resource "cloudfoundry_org_quota" "runaway" {
	name = "runaway_test"
    allow_paid_service_plans = true
    instance_memory = -1
    total_app_instances = -1
    total_memory = 204800
    total_routes = 2000
    total_services = -1
    total_route_ports = 0
}
resource "cloudfoundry_user" "u1" {
    name = "test-user1@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u2" {
    name = "test-user2@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u3" {
    name = "test-user3@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u4" {
    name = "test-user4@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u5" {
    name = "test-user5@acme.com"
	password = "password"
}

resource "cloudfoundry_org" "org1" {

    name = "organization-one"
    quota = "${cloudfoundry_org_quota.runaway.id}"
    managers = [ "${cloudfoundry_user.u1.id}", "${cloudfoundry_user.u2.id}" ]
    billing_managers = [ "${cloudfoundry_user.u3.id}", "${cloudfoundry_user.u4.id}" ]
    auditors = [ "${cloudfoundry_user.u5.id}" ]
}
`

const orgResourceUpdate = `

data "cloudfoundry_org_quota" "default" {
  name = "default"
}

resource "cloudfoundry_org_quota" "runaway" {
	name = "runaway_test"
    allow_paid_service_plans = true
    instance_memory = -1
    total_app_instances = -1
    total_memory = 204800
    total_routes = 2000
    total_services = -1
    total_route_ports = 0
}
resource "cloudfoundry_user" "u1" {
    name = "test-user1@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u2" {
    name = "test-user2@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u3" {
    name = "test-user3@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u4" {
    name = "test-user4@acme.com"
	password = "password"
}
resource "cloudfoundry_user" "u5" {
    name = "test-user5@acme.com"
	password = "password"
}

resource "cloudfoundry_org" "org1" {
	name = "organization-one-updated"
  quota = "${data.cloudfoundry_org_quota.default.id}"
	managers = [ "${cloudfoundry_user.u1.id}" ]
	billing_managers = [ "${cloudfoundry_user.u2.id}", "${cloudfoundry_user.u3.id}" ]
	auditors = [ "${cloudfoundry_user.u5.id}" ]
}
`

func TestAccOrg_normal(t *testing.T) {

	refOrg := "cloudfoundry_org.org1"
	refQuotaRunway := "cloudfoundry_org_quota.runaway"
	refQuotaDefault := "data.cloudfoundry_org_quota.default"
	refUserRemoved := "cloudfoundry_user.u4"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckOrgDestroyed("organization-one-updated"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: orgResource,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckOrgExists(refOrg, refQuotaRunway, nil),
						resource.TestCheckResourceAttr(
							refOrg, "name", "organization-one"),
						resource.TestCheckResourceAttr(
							refOrg, "managers.#", "2"),
						resource.TestCheckResourceAttr(
							refOrg, "billing_managers.#", "2"),
						resource.TestCheckResourceAttr(
							refOrg, "auditors.#", "1"),
					),
				},

				resource.TestStep{
					Config: orgResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckOrgExists(refOrg, refQuotaDefault, &refUserRemoved),
						resource.TestCheckResourceAttr(
							refOrg, "name", "organization-one-updated"),
						resource.TestCheckResourceAttr(
							refOrg, "managers.#", "1"),
						resource.TestCheckResourceAttr(
							refOrg, "billing_managers.#", "2"),
						resource.TestCheckResourceAttr(
							refOrg, "auditors.#", "1"),
					),
				},
			},
		})
}

func testAccCheckOrgExists(resOrg, resQuota string, refUserRemoved *string) resource.TestCheckFunc {

	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*cfapi.Session)

		rs, ok := s.RootModule().Resources[resOrg]
		if !ok {
			return fmt.Errorf("org '%s' not found in terraform state", resOrg)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resOrg, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var org cfapi.CCOrg
		om := session.OrgManager()
		if org, err = om.ReadOrg(id); err != nil {
			return err
		}
		session.Log.DebugMessage(
			"retrieved org for resource '%s' with id '%s': %# v",
			resOrg, id, org)

		if err = assertEquals(attributes, "name", org.Name); err != nil {
			return err
		}
		if err = assertEquals(attributes, "quota", org.QuotaGUID); err != nil {
			return err
		}

		rs = s.RootModule().Resources[resQuota]
		if org.QuotaGUID != rs.Primary.ID {
			return fmt.Errorf("expected org '%s' to be associated with quota '%s' but it was not", resOrg, resQuota)
		}

		for t, r := range orgRoleMap {
			var users []interface{}
			if users, err = om.ListUsers(id, r); err != nil {
				return err
			}
			if err = assertSetEquals(attributes, t, users); err != nil {
				return err
			}
		}

		return testUserRemovedFromOrg(refUserRemoved, id, om, s)
	}
}

func testAccCheckOrgDestroyed(orgname string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*cfapi.Session)
		if _, err := session.OrgManager().FindOrg(orgname); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}
		return fmt.Errorf("org with name '%s' still exists in cloud foundry", orgname)
	}
}

func testUserRemovedFromOrg(
	refUserRemoved *string,
	orgID string,
	om *cfapi.OrgManager,
	s *terraform.State) (err error) {

	if refUserRemoved != nil {

		rs, found := s.RootModule().Resources[*refUserRemoved]
		if !found {
			err = fmt.Errorf("expected user resource '%s' was not found", *refUserRemoved)
			return
		}

		var users []interface{}
		if users, err = om.ListUsers(orgID, cfapi.OrgRoleMember); err != nil {
			return
		}

		found = false
		for _, u := range users {
			if rs.Primary.ID == u {
				found = true
				break
			}
		}
		if found {
			err = fmt.Errorf(
				"expected user resource '%s' with if '%s' to be removed from the organization but it was not",
				*refUserRemoved, rs.Primary.ID)
			return
		}
	}
	return err
}
