package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strings"
	"testing"
)

const testSetupTfResourceForSpaceAsgs = `
resource "cloudfoundry_asg" "asg1" {
	name = "asg-required-at-runtime"
    rule {
        protocol = "all"
        destination = "192.168.100.0/24"
    }
}
resource "cloudfoundry_asg" "asg2" {
	name = "asg-required-at-staging"
    rule {
        protocol = "all"
        destination = "192.168.101.0/24"
    }
}
resource "cloudfoundry_org" "org1" {
	name = "organization-one"
}
resource "cloudfoundry_space" "space1" {
    name = "space-two"
	org = cloudfoundry_org.org1.id
	quota = ""
}
resource "cloudfoundry_space_asgs" "spaceasgs1" {
    space = cloudfoundry_space.space1.id
    running_asgs = [ cloudfoundry_asg.asg1.id ]
    staging_asgs = [ cloudfoundry_asg.asg2.id ]
}
`

func TestAccResSpaceAsgs(t *testing.T) {

	resource.Test(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			CheckDestroy:      testAccCheckSpaceAsgsDestroyed(),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: testSetupTfResourceForSpaceAsgs,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSpaceAsgsExists("cloudfoundry_space.space1"),
					),
				},
			},
		})
}

func testAccCheckSpaceAsgsExists(resourceTfName string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resourceTfName]
		if !ok {
			return fmt.Errorf("Resource '%s' not found in terraform state", resourceTfName)
		}
		id := rs.Primary.ID

		runningAsgs, _, err := session.ClientV2.GetSpaceSecurityGroups(id)
		if err != nil {
			return fmt.Errorf("Error while trying to load asgs for space %s: %s.", id, err.Error())
		}

		finalRunningAsgs, _ := getInSlice(runningAsgs, func(object interface{}) bool {
			return !object.(ccv2.SecurityGroup).RunningDefault && object.(ccv2.SecurityGroup).Name == "asg-required-at-runtime"
		})

		if len(finalRunningAsgs) != 1 {
			return fmt.Errorf("Expected one running asg named 'asg-required-at-runtime' but found '%s'.", getNamesOfSecorityGroups(finalRunningAsgs))
		}

		stagingAsgs, _, err := session.ClientV2.GetSpaceStagingSecurityGroups(id)

		finalStagingAsgs, _ := getInSlice(stagingAsgs, func(object interface{}) bool {
			return !object.(ccv2.SecurityGroup).StagingDefault && object.(ccv2.SecurityGroup).Name == "asg-required-at-staging"
		})

		if len(finalStagingAsgs) != 1 {
			return fmt.Errorf("Expected one staging asg named 'asg-required-at-staging' but found '%s'.", getNamesOfSecorityGroups(finalStagingAsgs))
		}

		return nil
	}
}

func getNamesOfSecorityGroups(objs []interface{}) string {
	var names []string
	for _, o := range objs {
		names = append(names, o.(ccv2.SecurityGroup).Name)
	}
	return strings.Join(names, ",")
}

func testAccCheckSpaceAsgsDestroyed() resource.TestCheckFunc {

	return func(s *terraform.State) error {
		// it does not make much sense to test this because the relation space - asg is not an object itself
		return nil
	}
}
