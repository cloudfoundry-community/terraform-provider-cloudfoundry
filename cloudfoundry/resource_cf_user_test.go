package cloudfoundry

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
)

const ldapUserResource = `

resource "cloudfoundry_user" "manager1" {
    name = "manager1@acme.com"
    origin = "ldap"
}
`

const userResourceWithGroups = `

resource "cloudfoundry_user" "admin-service-user" {
    name = "cf-admin"
	password = "qwerty"
	given_name = "Build"
	family_name = "User"
    groups = [ "cloud_controller.admin", "scim.read", "scim.write" ]
}
`

const userResourceWithGroupsUpdate = `

resource "cloudfoundry_user" "admin-service-user" {
    name = "cf-admin"
	password = "asdfg"
	email = "cf-admin@acme.com"
    groups = [ "cloud_controller.admin", "clients.admin", "uaa.admin", "doppler.firehose" ]
}
`

const userResourceWithEmptyGroup = `

resource "cloudfoundry_user" "empty-group" {
	name = "jdoe"
	password = "password"
	origin = "uaa"
	given_name = "John"
	family_name = "Doe"
	email = "john.doe@acme.com"
	groups = []
}
`
const userResourceWithEmptyGroupUpdate = `

resource "cloudfoundry_user" "empty-group" {
	name = "jdoe"
	password = "password"
	origin = "uaa"
	given_name = "John2"
	family_name = "Doe2"
	email = "john.doe@acme.com"
	groups = []
}
`

func TestAccResUser_LdapOrigin_normal(t *testing.T) {

	ref := "cloudfoundry_user.manager1"
	username := "manager1@acme.com"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserDestroy(username),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: ldapUserResource,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", username),
						resource.TestCheckResourceAttr(
							ref, "origin", "ldap"),
						resource.TestCheckResourceAttr(
							ref, "email", username),
						resource.TestCheckResourceAttr(
							ref, "groups.#", "0"),
					),
				},
			},
		})
}

func TestAccResUser_WithGroups_normal(t *testing.T) {

	ref := "cloudfoundry_user.admin-service-user"
	username := "cf-admin"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserDestroy(username),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: userResourceWithGroups,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", username),
						resource.TestCheckResourceAttr(
							ref, "password", "qwerty"),
						resource.TestCheckResourceAttr(
							ref, "email", username),
						resource.TestCheckResourceAttr(
							ref, "groups.#", "3"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("cloud_controller.admin")),
							"cloud_controller.admin"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("scim.read")),
							"scim.read"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("scim.write")),
							"scim.write"),
					),
				},

				resource.TestStep{
					Config: userResourceWithGroupsUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "cf-admin"),
						resource.TestCheckResourceAttr(
							ref, "password", "asdfg"),
						resource.TestCheckResourceAttr(
							ref, "email", "cf-admin@acme.com"),
						resource.TestCheckResourceAttr(
							ref, "groups.#", "4"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("cloud_controller.admin")),
							"cloud_controller.admin"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("clients.admin")),
							"clients.admin"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("uaa.admin")),
							"uaa.admin"),
						resource.TestCheckResourceAttr(
							ref, "groups."+strconv.Itoa(hashcode.String("doppler.firehose")),
							"doppler.firehose"),
					),
				},
			},
		})
}

func TestAccResUser_EmptyGroups_normal(t *testing.T) {

	ref := "cloudfoundry_user.empty-group"
	username := "jdoe"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckUserDestroy(username),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: userResourceWithEmptyGroup,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", username),
						resource.TestCheckResourceAttr(
							ref, "origin", "uaa"),
						resource.TestCheckResourceAttr(
							ref, "email", "john.doe@acme.com"),
						resource.TestCheckResourceAttr(
							ref, "groups.#", "0"),
					),
				},

				resource.TestStep{
					Config: userResourceWithEmptyGroupUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckUserExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", username),
						resource.TestCheckResourceAttr(
							ref, "origin", "uaa"),
						resource.TestCheckResourceAttr(
							ref, "email", "john.doe@acme.com"),
						resource.TestCheckResourceAttr(
							ref, "groups.#", "0"),
					),
				},
			},
		})
}

func testAccCheckUserExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("user '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		um := session.ClientUAA
		user, err := um.GetUser(id)
		if err != nil {
			return err
		}

		if err = assertEquals(attributes, "name", user.Username); err != nil {
			return err
		}
		if err = assertEquals(attributes, "origin", user.Origin); err != nil {
			return err
		}
		if err = assertEquals(attributes, "given_name", user.Name.GivenName); err != nil {
			return err
		}
		if err = assertEquals(attributes, "family_name", user.Name.FamilyName); err != nil {
			return err
		}
		if err = assertEquals(attributes, "email", user.Emails[0].Value); err != nil {
			return err
		}

		return err
	}
}

func testAccCheckUserDestroy(username string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
		um := session.ClientUAA
		users, err := um.GetUsersByUsername(username)
		if err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}
		if len(users) > 0 {
			return fmt.Errorf("user with username '%s' still exists in cloud foundry", username)
		}
		return nil
	}
}
