package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const spaceQuotaDataResource = `

resource "cloudfoundry_space_quota" "q" {
  name = "20g-space-ds"
  allow_paid_service_plans = false
  instance_memory = 512
  total_memory = 10240
  total_app_instances = 10
  total_routes = 5
  total_services = 20
  org = "%s"
}

data "cloudfoundry_space_quota" "qq" {
  name = "${cloudfoundry_space_quota.q.name}"
  org = "%s"
}
`

func TestAccDataSourceSpaceQuota_normal(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	ut := os.Getenv("UNIT_TEST")
	if !testAccEnvironmentSet() || (len(ut) > 0 && ut != filepath.Base(filename)) {
		fmt.Printf("Skipping tests in '%s'.\n", filepath.Base(filename))
		return
	}

	ref := "data.cloudfoundry_space_quota.qq"
	orgID, _ := defaultTestOrg(t)

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(spaceQuotaDataResource, orgID, orgID),
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceSpaceQuotaExists(ref),
						resource.TestCheckResourceAttr(ref, "name", "20g-space-ds"),
						resource.TestCheckResourceAttr(ref, "org", orgID),
					),
				},
			},
		})
}

func checkDataSourceSpaceQuotaExists(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("quota '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes
		var (
			err   error
			quota ccv2.Quota
		)
		quota, _, err = session.ClientV2.GetQuota(constant.SpaceQuota, id)
		if err != nil {
			return err
		}
		if err := assertEquals(attributes, "org", quota.OrganizationGUID); err != nil {
			return err
		}
		return nil
	}
}
