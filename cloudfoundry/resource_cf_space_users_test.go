package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const spaceUsersResource = `
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

resource "cloudfoundry_space_users" "space_users1" {
	space = "%s"
    managers = [
        "${cloudfoundry_user.tl.id}"
    ]
    developers = [
        "${cloudfoundry_user.tl.id}",
        "${cloudfoundry_user.dev1.id}",
		"${cloudfoundry_user.dev2.id}",
        "username@acme.com"
    ]
    auditors = [
        "${cloudfoundry_user.adr.id}",
		"${cloudfoundry_user.dev3.id}"
    ]
}
`

const spaceUsersResourceUpdate = `
resource "cloudfoundry_user" "dev1" {
    name = "developer1@acme.com"
	password = "password"
}

resource "cloudfoundry_space_users" "space_users1" {
	space = "%s"
    developers = [
        "${cloudfoundry_user.dev1.id}",
    ]
	auditors = []
	managers = []
}
`

const spaceUsersResourceForce = `
resource "cloudfoundry_user" "dev1" {
    name = "developer1@acme.com"
	password = "password"
}

resource "cloudfoundry_space_users" "space_users1" {
	space = "%s"
    developers = [
        "${cloudfoundry_user.dev1.id}",
    ]
	force = true
}
`

func TestAccResSpaceUsers_normal(t *testing.T) {
	ref := "cloudfoundry_space_users.space_users1"
	spaceId, _ := defaultTestSpace(t)
	orgId, _ := defaultTestOrg(t)
	usersMap := make(map[string][]ccv2.User)

	sessions := testSession()
	user, err := sessions.ClientUAA.CreateUser("username@acme.com", "paasw0rd", "uaa")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer sessions.ClientUAA.DeleteUser(user.ID)
	err = addOrNothingUserInOrgBySpace(sessions, orgId, user.ID, false)
	if err != nil {
		t.Fatal(err.Error())
	}

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(spaceUsersResource, spaceId),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSpaceUsersExists(ref, &usersMap),
						resource.TestCheckResourceAttr(
							ref, "managers.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "developers.#", "4"),
						resource.TestCheckResourceAttr(
							ref, "auditors.#", "2"),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(spaceUsersResourceUpdate, spaceId),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSpaceUsersExists(ref, &usersMap),
						testAccCheckMapUserInside("developer1@acme.com", "developers", &usersMap),
						resource.TestCheckResourceAttr(
							ref, "managers.#", "0"),
						resource.TestCheckResourceAttr(
							ref, "developers.#", "1"),
						resource.TestCheckResourceAttr(
							ref, "auditors.#", "0"),
					),
				},
			},
		})
}

func TestAccResSpaceUsers_force(t *testing.T) {
	ref := "cloudfoundry_space_users.space_users1"
	spaceId, _ := defaultTestSpace(t)
	orgId, _ := defaultTestOrg(t)
	usersMap := make(map[string][]ccv2.User)

	sessions := testSession()
	user, err := sessions.ClientUAA.CreateUser("test-acc-force@acme.com", "paasw0rd", "uaa")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer sessions.ClientUAA.DeleteUser(user.ID)
	err = addOrNothingUserInOrgBySpace(sessions, orgId, user.ID, false)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = sessions.ClientV2.UpdateSpaceManager(spaceId, user.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(spaceUsersResourceForce, spaceId),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSpaceUsersExists(ref, &usersMap),
						testAccCheckMapUsersNumber(1, &usersMap),
						testAccCheckMapUserInside("developer1@acme.com", "developers", &usersMap),
						resource.TestCheckResourceAttr(
							ref, "developers.#", "1"),
					),
				},
			},
		})
}

func testAccCheckSpaceUsersExists(resource string, users *map[string][]ccv2.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		attributes := rs.Primary.Attributes

		sm := session.ClientV2
		usersMap := make(map[string][]ccv2.User)
		for t, r := range typeToSpaceRoleMap {
			users, _, err := sm.GetSpaceUsersByRole(r, attributes["space"])
			if err != nil {
				return err
			}
			usersMap[t] = users
		}
		if len(usersMap) == 0 {
			return fmt.Errorf("No space users found")
		}
		*users = usersMap
		return nil
	}
}

func testAccCheckMapUserInside(name, role string, users *map[string][]ccv2.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		m := *users
		for _, u := range m[role] {
			if u.Username == name {
				return nil
			}
		}
		return fmt.Errorf("User %s with role %s not found", name, role)
	}
}

func testAccCheckMapUsersNumber(number int, users *map[string][]ccv2.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		l := 0
		m := *users
		for _, v := range m {
			l += len(v)
		}
		if l != number {
			return fmt.Errorf("There is %d users instead of %d", l, number)
		}
		return nil
	}
}
