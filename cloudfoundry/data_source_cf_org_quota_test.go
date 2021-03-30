package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const orgQuotaDataResource = `

resource "cloudfoundry_org_quota" "q" {
  name = "100g-org-ds"
  allow_paid_service_plans = false
  instance_memory = 2048
  total_memory = 51200
  total_app_instances = 100
  total_routes = 50
  total_services = 200
  total_route_ports = 5
}

data "cloudfoundry_org_quota" "qq" {
  name = "${cloudfoundry_org_quota.q.name}"
}
`

func TestAccDataSourceOrgQuota_normal(t *testing.T) {
	ref := "data.cloudfoundry_org_quota.qq"
	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: orgQuotaDataResource,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceOrgQuotaExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "100g-org-ds"),
					),
				},
			},
		})
}

func checkDataSourceOrgQuotaExists(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("quota '%s' not found in terraform state", resource)
		}
		id := rs.Primary.ID
		var (
			err error
		)
		_, _, err = session.ClientV2.GetQuota(constant.OrgQuota, id)
		return err
	}
}
