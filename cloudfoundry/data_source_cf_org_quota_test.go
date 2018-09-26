package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const orgQuotaDataResource = `

resource "cloudfoundry_org_quota" "q" {
  name = "100g-org"
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
	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: orgQuotaDataResource,
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceOrgQuotaExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "100g-org"),
					),
				},
			},
		})
}

func checkDataSourceOrgQuotaExists(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*cfapi.Session)
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("quota '%s' not found in terraform state", resource)
		}
		session.Log.DebugMessage("terraform state for resource '%s': %# v", resource, rs)
		id := rs.Primary.ID
		var (
			err error
		)
		_, err = session.QuotaManager().ReadQuota(cfapi.OrgQuota, id)
		return err
	}
}
