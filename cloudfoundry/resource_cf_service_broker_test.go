package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const sbResource = `

resource "cloudfoundry_service_broker" "test" {
	name = "test"
	url = "%s"
	username = "%s"
	password = "%s"
}
`

const sbResourceUpdate = `

resource "cloudfoundry_service_broker" "test" {
	name = "test-renamed"
	url = "%s"
	username = "%s"
	password = "%s"
}
`

const sbResourceUpdateCatalog = `

resource "cloudfoundry_service_broker" "test" {
	name = "test"
	url = "%s"
	username = "%s"
	password = "%s"
	catalog_hash = "1"
}
`

const sbResourceUpdateCatalogFail = `
resource "cloudfoundry_service_broker" "test" {
	name = "test-fail"
	url = "%s/not/accessible"
	username = "%s"
	password = "%s"
}
`

const sbResourceUpdateCatalogFailShow = `
resource "cloudfoundry_service_broker" "test" {
	name = "test-fail-show"
	url = "%s/not/accessible"
	username = "%s"
	password = "%s"
    fail_when_catalog_not_accessible = true
}
`

func TestAccResServiceBroker_normal(t *testing.T) {

	serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword, serviceBrokerPlanPath := getTestBrokerCredentials(t)

	// Ensure any test artifacts from a
	// failed run are deleted if the exist
	deleteServiceBroker("test")
	deleteServiceBroker("test-renamed")

	ref := "cloudfoundry_service_broker.test"
	var catalogHash string
	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceBrokerDestroyed("test"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(sbResource,
						serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceBrokerExists(ref, &catalogHash),
						resource.TestCheckResourceAttr(
							ref, "name", "test"),
						resource.TestCheckResourceAttrSet(
							ref, "catalog_hash"),
						resource.TestCheckResourceAttr(
							ref, "url", serviceBrokerURL),
						resource.TestCheckResourceAttr(
							ref, "username", serviceBrokerUser),
						resource.TestCheckResourceAttrSet(
							ref, "service_plans."+serviceBrokerPlanPath),
					),
				},
				resource.TestStep{
					Config: fmt.Sprintf(sbResourceUpdateCatalog,
						serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckServiceBrokerExists(ref, nil),
						testAccCheckCatalogHash(ref, &catalogHash),
						resource.TestCheckResourceAttr(
							ref, "name", "test"),
						resource.TestCheckResourceAttr(
							ref, "url", serviceBrokerURL),
						resource.TestCheckResourceAttr(
							ref, "username", serviceBrokerUser),
						resource.TestCheckResourceAttrSet(
							ref, "service_plans."+serviceBrokerPlanPath),
					),
					ExpectNonEmptyPlan: true,
				},

				// Bug found when updating service broker, service broker can't be renamed
				// resource.TestStep{
				// 	Config: fmt.Sprintf(sbResourceUpdate,
				// 		serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword),
				// 	Check: resource.ComposeTestCheckFunc(
				// 		testAccCheckServiceBrokerExists(ref),
				// 		resource.TestCheckResourceAttr(
				// 			ref, "name", "test-renamed"),
				// 	),
				// },
			},
		})
}

func TestAccResServiceBroker_fail(t *testing.T) {

	serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword, _ := getTestBrokerCredentials(t)
	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceBrokerDestroyed("test"),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(sbResourceUpdateCatalogFail,
						serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword),
					ExpectError: regexp.MustCompile("^.*service broker rejected the request.*$"),
				},
			},
		})
}

func TestAccResServiceBroker_failShow(t *testing.T) {

	serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword, _ := getTestBrokerCredentials(t)
	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceBrokerDestroyed("test"),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(sbResourceUpdateCatalogFailShow,
						serviceBrokerURL, serviceBrokerUser, serviceBrokerPassword),
					ExpectError: regexp.MustCompile("^.*Error when getting catalog signature.*$"),
				},
			},
		})
}

func testAccCheckCatalogHash(resource string, catalogHash *string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service broker '%s' not found in terraform state", resource)
		}
		attributes := rs.Primary.Attributes
		if attributes["catalog_hash"] != *catalogHash {
			return fmt.Errorf("Catalog hash mismatch, expected '%s' got '%s'", *catalogHash, attributes["catalog_hash"])
		}
		return nil
	}
}

func testAccCheckServiceBrokerExists(resource string, catalogHash *string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("service broker '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes
		if catalogHash != nil && *catalogHash == "" {
			*catalogHash = attributes["catalog_hash"]
		}

		serviceBroker, _, err := session.ClientV2.GetServiceBroker(id)
		if err != nil {
			return err
		}
		if err := assertEquals(attributes, "name", serviceBroker.Name); err != nil {
			return err
		}
		if err := assertEquals(attributes, "url", serviceBroker.BrokerURL); err != nil {
			return err
		}
		if err := assertEquals(attributes, "username", serviceBroker.AuthUsername); err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckServiceBrokerDestroyed(name string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)
		sbs, _, err := session.ClientV2.GetServiceBrokers(ccv2.FilterByName(name))
		if err != nil {
			return err
		}
		if len(sbs) > 0 {
			return fmt.Errorf("service broker with name '%s' still exists in cloud foundry", name)
		}
		return nil
	}
}
