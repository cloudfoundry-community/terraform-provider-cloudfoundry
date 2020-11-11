package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestAccResDefaultRunningAsg_normal(t *testing.T) {

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
	resource.ParallelTest(t,
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
					ResourceName: ref,
					ImportState:  true,
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						if len(states) == 0 {
							return fmt.Errorf("There is no import state")
						}
						entity := resourceDefaultAsg()
						state := states[0]
						reader := &schema.MapFieldReader{
							Schema: entity.Schema,
							Map:    schema.BasicMapReader(state.Attributes),
						}
						result, err := reader.ReadField([]string{"asgs"})
						if err != nil {
							return err
						}
						if len(result.Value.(*schema.Set).List()) != defaultLenRunningSecGroup+2 {
							return fmt.Errorf("missing default running sec group")
						}
						return nil
					},
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
