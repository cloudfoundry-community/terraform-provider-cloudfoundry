package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const evgRunningResource = `

resource "cloudfoundry_evg" "running" {

	name = "running"

    variables = {
        name1 = "value1"
        name2 = "value2"
        name3 = "value3"
        name4 = "value4"
    }
}
`

const evgRunningResourceUpdated = `

resource "cloudfoundry_evg" "running" {

	name = "running"

    variables = {
        name1 = "value1"
        name2 = "value2"
        name3 = "valueC"
        name4 = "valueD"
        name5 = "valueE"
    }
}
`

const evgStagingResource = `
resource "cloudfoundry_evg" "staging" {

	name = "staging"

    variables = {
        name3 = "value3"
        name4 = "value4"
        name5 = "value5"
    }
}
`

const evgStagingResourceUpdated = `
resource "cloudfoundry_evg" "staging" {

	name = "staging"

    variables = {
        name4 = "value4"
        name5 = "valueE"
    }
}
`

var defaultLenStagingEvg int
var defaultLenRunningEvg int

func TestAccResunningEvg_normal(t *testing.T) {
	evg, _, err := testSession().ClientV2.GetEnvVarGroupRunning()
	if err != nil {
		panic(err)
	}
	defaultLenRunningEvg = len(evg)
	evg, _, err = testSession().ClientV2.GetEnvVarGroupStaging()
	if err != nil {
		panic(err)
	}
	defaultLenStagingEvg = len(evg)

	ref := "cloudfoundry_evg.running"
	name := "running"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckEvgDestroy(name),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: evgRunningResource,
					Check: resource.ComposeTestCheckFunc(
						checkEvgExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "running"),
						resource.TestCheckResourceAttr(
							ref, "variables.%", "4"),
						resource.TestCheckResourceAttr(
							ref, "variables.name1", "value1"),
						resource.TestCheckResourceAttr(
							ref, "variables.name2", "value2"),
						resource.TestCheckResourceAttr(
							ref, "variables.name3", "value3"),
						resource.TestCheckResourceAttr(
							ref, "variables.name4", "value4"),
					),
				},
				resource.TestStep{
					ResourceName:      ref,
					ImportState:       true,
					ImportStateVerify: true,
				},
				resource.TestStep{
					Config: evgRunningResourceUpdated,
					Check: resource.ComposeTestCheckFunc(
						checkEvgExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "running"),
						resource.TestCheckResourceAttr(
							ref, "variables.%", "5"),
						resource.TestCheckResourceAttr(
							ref, "variables.name1", "value1"),
						resource.TestCheckResourceAttr(
							ref, "variables.name2", "value2"),
						resource.TestCheckResourceAttr(
							ref, "variables.name3", "valueC"),
						resource.TestCheckResourceAttr(
							ref, "variables.name4", "valueD"),
						resource.TestCheckResourceAttr(
							ref, "variables.name5", "valueE"),
					),
				},
			},
		})
}

func TestAccResStagingEvg_normal(t *testing.T) {

	ref := "cloudfoundry_evg.staging"
	name := "staging"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckEvgDestroy(name),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: evgStagingResource,
					Check: resource.ComposeTestCheckFunc(
						checkEvgExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "staging"),
						resource.TestCheckResourceAttr(
							ref, "variables.%", "3"),
						resource.TestCheckResourceAttr(
							ref, "variables.name3", "value3"),
						resource.TestCheckResourceAttr(
							ref, "variables.name4", "value4"),
						resource.TestCheckResourceAttr(
							ref, "variables.name5", "value5"),
					),
				},
				resource.TestStep{
					Config: evgStagingResourceUpdated,
					Check: resource.ComposeTestCheckFunc(
						checkEvgExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "staging"),
						resource.TestCheckResourceAttr(
							ref, "variables.%", "2"),
						resource.TestCheckResourceAttr(
							ref, "variables.name4", "value4"),
						resource.TestCheckResourceAttr(
							ref, "variables.name5", "valueE"),
					),
				},
			},
		})
}

func checkEvgExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("asg '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var variables map[string]string
		var err error
		switch id {
		case AppStatusRunning:
			variables, _, err = session.ClientV2.GetEnvVarGroupRunning()
		case AppStatusStaging:
			variables, _, err = session.ClientV2.GetEnvVarGroupStaging()
		}
		if err != nil {
			return err
		}
		variablesInterface := make(map[string]interface{})
		for k, v := range variables {
			variablesInterface[k] = v
		}
		if err := assertMapEquals("variables", attributes, variablesInterface); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckEvgDestroy(name string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)

		var variables map[string]string
		var defaultLen int
		var err error
		switch name {
		case AppStatusRunning:
			variables, _, err = session.ClientV2.GetEnvVarGroupRunning()
			defaultLen = defaultLenRunningEvg
		case AppStatusStaging:
			variables, _, err = session.ClientV2.GetEnvVarGroupStaging()
			defaultLen = defaultLenStagingEvg
		}
		if err != nil {
			return err
		}
		if len(variables) != defaultLen {
			return fmt.Errorf("%s variables are not empty", name)
		}
		return nil
	}
}
