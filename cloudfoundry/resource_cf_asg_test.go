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

const securityGroup = `
resource "cloudfoundry_asg" "rmq" {

	name = "rmq-dev-res"

    rule {
        protocol = "tcp"
        destination = "192.168.1.100"
        ports = "5672,5671,1883,8883,61613,61614"
		log = true
    }

    rule {
        protocol = "icmp"
        destination = "192.168.1.101"
		type = 8
		code = 0
    }

}
`

const securityGroupUpdate = `
resource "cloudfoundry_asg" "rmq" {

	name = "rmq-dev-res"

    rule {
        protocol = "tcp"
        destination = "192.168.1.100"
        ports = "61613,61614"
    }

    rule {
        protocol = "tcp"
        destination = "192.168.1.0/24"
        ports = "61613,61614"
    }

    rule {
        protocol = "all"
        destination = "0.0.0.0/0"
		log = true
    }

}
`

func TestAccResAsg_normal(t *testing.T) {

	ref := "cloudfoundry_asg.rmq"
	asgname := "rmq-dev-res"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckASGDestroy(asgname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: securityGroup,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckASGExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", asgname),
						resource.TestCheckResourceAttr(
							ref, "rule.#", "2"),
						resource.TestCheckResourceAttr(
							ref, "rule.0.protocol", "tcp"),
						resource.TestCheckResourceAttr(
							ref, "rule.0.destination", "192.168.1.100"),
						resource.TestCheckResourceAttr(
							ref, "rule.0.ports", "5672,5671,1883,8883,61613,61614"),
						resource.TestCheckResourceAttr(
							ref, "rule.0.log", "true"),
						resource.TestCheckResourceAttr(
							ref, "rule.1.protocol", "icmp"),
						resource.TestCheckResourceAttr(
							ref, "rule.1.destination", "192.168.1.101"),
						resource.TestCheckResourceAttr(
							ref, "rule.1.ports", ""),
						resource.TestCheckResourceAttr(
							ref, "rule.1.type", "8"),
						resource.TestCheckResourceAttr(
							ref, "rule.1.code", "0"),
					),
				},
				resource.TestStep{
					ResourceName:      ref,
					ImportState:       true,
					ImportStateVerify: true,
				},
				resource.TestStep{
					Config: securityGroupUpdate,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckASGExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", asgname),
						resource.TestCheckResourceAttr(
							ref, "rule.#", "3"),
						resource.TestCheckResourceAttr(
							ref, "rule.0.protocol", "tcp"),
						resource.TestCheckResourceAttr(
							ref, "rule.0.destination", "192.168.1.100"),
						resource.TestCheckResourceAttr(
							ref, "rule.0.ports", "61613,61614"),
						resource.TestCheckResourceAttr(
							ref, "rule.1.protocol", "tcp"),
						resource.TestCheckResourceAttr(
							ref, "rule.1.destination", "192.168.1.0/24"),
						resource.TestCheckResourceAttr(
							ref, "rule.1.ports", "61613,61614"),
						resource.TestCheckResourceAttr(
							ref, "rule.2.protocol", "all"),
						resource.TestCheckResourceAttr(
							ref, "rule.2.destination", "0.0.0.0/0"),
						resource.TestCheckResourceAttr(
							ref, "rule.2.ports", ""),
						resource.TestCheckResourceAttr(
							ref, "rule.2.log", "true"),
					),
				},
			},
		})
}

func testAccCheckASGExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("asg '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		asg, _, err := session.ClientV2.GetSecurityGroup(id)
		if err != nil {
			return err
		}

		if err := assertEquals(attributes, "name", asg.Name); err != nil {
			return err
		}

		if err := assertListEquals(attributes, "rule", len(asg.Rules),
			func(values map[string]string, i int) (match bool) {

				return values["protocol"] == asg.Rules[i].Protocol &&
					values["destination"] == asg.Rules[i].Destination &&
					values["ports"] == asg.Rules[i].Ports &&
					values["log"] == strconv.FormatBool(asg.Rules[i].Log.Value)

			}); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckASGDestroy(asgname string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)
		asgs, _, err := session.ClientV2.GetSecurityGroups(ccv2.FilterByName(asgname))
		if err != nil {
			return err
		}
		if len(asgs) > 0 {
			return fmt.Errorf("asg with name '%s' still exists in cloud foundry", asgname)
		}
		return nil
	}
}
