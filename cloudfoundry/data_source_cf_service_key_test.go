package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const serviceKeyDataResource = `

data "cf_org" "org" {
	name = "%s"
}
data "cf_space" "space" {
	name = "%s"
	org = "${data.cf_org.org.id}"
}

data "cf_service" "s1" {
	name = "%s"
}
resource "cf_service_instance" "db" {
	name = "db-update"
	space = "${data.cf_space.space.id}"
	service_plan = "${data.cf_service.s1.service_plans.%s}"
}
resource "cf_service_key" "test" {
	name = "test-key-name"
	service_instance = "${cf_service_instance.db.id}"
}
data "cf_service_key" "test" {
	name = "${cf_service_key.test.name}"
	service_instance = "${cf_service_instance.db.id}"
	depends_on = [ cf_service_key.test ]
}
`

func TestAccDataSourceServiceKey_normal(t *testing.T) {

	serviceName1, _, servicePlan := getTestServiceBrokers(t)
	_, orgName := defaultTestOrg(t)
	_, spaceName := defaultTestSpace(t)

	ref := "data.cf_service_key.test"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProvidersFactories,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(serviceKeyDataResource, orgName, spaceName,
						serviceName1, servicePlan),
					Check: func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[ref]
						if !ok {
							return fmt.Errorf("service key '%s' not found in terraform state", ref)
						}
						if rs.Primary.Attributes["name"] != "test-key-name" {
							return fmt.Errorf("service key name is incorrect!")
						}
						return nil
					},
				},
			},
		})
}
