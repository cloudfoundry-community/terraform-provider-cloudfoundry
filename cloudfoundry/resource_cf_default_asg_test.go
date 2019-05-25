package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const defaultRunningSecurityGroupResource = `

resource "cloudfoundry_asg" "apps" {

	name = "pcf-apps"

    rule {
        destination = "192.168.100.0/24"
        protocol = "all"
    }
}

resource "cloudfoundry_asg" "services" {

	name = "pcf-services"

    rule {
        destination = "192.168.101.0/24"
        protocol = "all"
    }
}

resource "cloudfoundry_default_asg" "running" {
	name = "running"
    asgs = [ "${cloudfoundry_asg.apps.id}", "${cloudfoundry_asg.services.id}" ]
}
`

const defaultRunningSecurityGroupResourceUpdate = `
resource "cloudfoundry_asg" "apps" {

	name = "pcf-apps"

    rule {
        destination = "192.168.100.0/24"
        protocol = "all"
    }
}

resource "cloudfoundry_asg" "services" {

	name = "pcf-services"

    rule {
        destination = "192.168.101.0/24"
        protocol = "all"
    }
}

resource "cloudfoundry_default_asg" "running" {
	name = "running"
    asgs = [ "${cloudfoundry_asg.apps.id}" ]
}
`

const defaultStagingSecurityGroupResource = `

resource "cloudfoundry_asg" "apps" {

	name = "pcf-apps"

    rule {
        destination = "192.168.100.0/24"
        protocol = "all"
    }
}

resource "cloudfoundry_default_asg" "staging" {
  name = "staging"
  asgs = [ "${cloudfoundry_asg.apps.id}" ]
}
`

var defaultLenRunningSecGroup int
var defaultLenStagingSecGroup int

func TestAccDefaultRunningAsg_normal(t *testing.T) {

	ref := "cloudfoundry_default_asg.running"
	asgs, _, err := testSession().ClientV2.GetRunningSecurityGroups()
	if err != nil {
		panic(err)
	}
	defaultLenRunningSecGroup = len(asgs)
	asgs, _, err = testSession().ClientV2.GetStagingSecurityGroups()
	if err != nil {
		panic(err)
	}
	defaultLenStagingSecGroup = len(asgs)
	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDefaultRunningAsgDestroy,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: defaultRunningSecurityGroupResource,
					Check: resource.ComposeTestCheckFunc(
						checkDefaultAsgsExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "running"),
						resource.TestCheckResourceAttr(
							ref, "asgs.#", "2"),
					),
				},
				resource.TestStep{
					Config: fmt.Sprintf(defaultRunningSecurityGroupResourceUpdate),
					Check: resource.ComposeTestCheckFunc(
						checkDefaultAsgsExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "running"),
						resource.TestCheckResourceAttr(
							ref, "asgs.#", "1"),
					),
				},
			},
		})
}

func TestAccDefaultStagingAsg_normal(t *testing.T) {

	ref := "cloudfoundry_default_asg.staging"

	asgs, _, err := testSession().ClientV2.GetRunningSecurityGroups()
	if err != nil {
		panic(err)
	}
	defaultLenRunningSecGroup = len(asgs)
	asgs, _, err = testSession().ClientV2.GetStagingSecurityGroups()
	if err != nil {
		panic(err)
	}
	defaultLenStagingSecGroup = len(asgs)

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDefaultStagingAsgDestroy,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: defaultStagingSecurityGroupResource,
					Check: resource.ComposeTestCheckFunc(
						checkDefaultAsgsExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "staging"),
						resource.TestCheckResourceAttr(
							ref, "asgs.#", "1"),
					),
				},
			},
		})
}

func checkDefaultAsgsExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("asg '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var asgs []ccv2.SecurityGroup
		var lenAsgs int
		var err error
		switch id {
		case "running":
			asgs, _, err = session.ClientV2.GetRunningSecurityGroups()
			if err != nil {
				return err
			}
			lenAsgs = len(asgs) - defaultLenRunningSecGroup
		case "staging":
			asgs, _, err = session.ClientV2.GetStagingSecurityGroups()
			if err != nil {
				return err
			}
			lenAsgs = len(asgs) - defaultLenStagingSecGroup
		}

		entity := resourceDefaultAsg()
		reader := &schema.MapFieldReader{
			Schema: entity.Schema,
			Map:    schema.BasicMapReader(attributes),
		}
		result, err := reader.ReadField([]string{"asgs"})
		if err != nil {
			return err
		}
		asgsTf := result.Value.(*schema.Set).List()

		if len(asgsTf) != lenAsgs {
			return fmt.Errorf("Expected %d asgs got %d", len(asgsTf), lenAsgs)
		}

		for _, asgTf := range asgsTf {
			inside := isInSlice(asgs, func(object interface{}) bool {
				asg := object.(ccv2.SecurityGroup)
				return asg.GUID == asgTf.(string)
			})
			if !inside {
				return fmt.Errorf("Missing creation of %s asgs", asgTf.(string))
			}
		}

		return nil
	}
}

func testAccCheckDefaultRunningAsgDestroy(s *terraform.State) error {

	session := testAccProvider.Meta().(*managers.Session)
	am := session.ClientV2

	asgs, _, err := am.GetRunningSecurityGroups()
	if err != nil {
		return err
	}
	if len(asgs) != defaultLenRunningSecGroup {
		return fmt.Errorf("running asgs are not empty")
	}

	return nil
}

func testAccCheckDefaultStagingAsgDestroy(s *terraform.State) error {

	session := testAccProvider.Meta().(*managers.Session)
	am := session.ClientV2

	asgs, _, err := am.GetStagingSecurityGroups()
	if err != nil {
		return err
	}
	if len(asgs) != defaultLenStagingSecGroup {
		return fmt.Errorf("staging asgs are not empty")
	}
	return nil
}
