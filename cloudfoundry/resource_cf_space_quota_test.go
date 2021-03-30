package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const spaceQuotaResource = `
resource "cloudfoundry_org_quota" "quota100g-org" {
  name                     = "100g-org"
  allow_paid_service_plans = false
  instance_memory          = 1024
  total_memory             = 102400
  total_app_instances      = 20
  total_routes             = 10
  total_services           = 20
  total_route_ports        = 10
}

resource "cloudfoundry_org" "quota-org" {
  name  = "quota-org"
  quota = "${cloudfoundry_org_quota.quota100g-org.id}"
}

resource "cloudfoundry_space_quota" "quota10g-space" {
  name                     = "10g-space"
  allow_paid_service_plans = false
  instance_memory          = 512
  total_memory             = 10240
  total_app_instances      = 10
  total_routes             = 5
  total_services           = 10
  total_route_ports        = 5
	org                      = "${cloudfoundry_org.quota-org.id}"
}
`

func TestAccResSpaceQuota_normal(t *testing.T) {

	orgID, _ := defaultTestOrg(t)

	ref := "cloudfoundry_space_quota.quota10g-space"
	quotaname := "10g-space"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSpaceQuotaResourceDestroy(quotaname, orgID),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: spaceQuotaResource,
					Check: resource.ComposeTestCheckFunc(
						checkSpaceQuotaExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "10g-space"),
						resource.TestCheckResourceAttr(
							ref, "allow_paid_service_plans", "false"),
						resource.TestCheckResourceAttr(
							ref, "instance_memory", "512"),
						resource.TestCheckResourceAttr(
							ref, "total_memory", "10240"),
						resource.TestCheckResourceAttr(
							ref, "total_app_instances", "10"),
						resource.TestCheckResourceAttr(
							ref, "total_routes", "5"),
						resource.TestCheckResourceAttr(
							ref, "total_services", "10"),
						resource.TestCheckResourceAttr(
							ref, "total_route_ports", "5"),
					),
				},
			},
		})
}

func checkSpaceQuotaExists(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		session := testAccProvider.Meta().(*managers.Session)
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("quota '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		quota, _, err := session.ClientV2.GetQuota(constant.SpaceQuota, id)
		if err != nil {
			return
		}

		if err := assertEquals(attributes, "name", quota.Name); err != nil {
			return err
		}
		if err := assertEquals(attributes, "org", quota.OrganizationGUID); err != nil {
			return err
		}
		if err := assertEquals(attributes, "allow_paid_service_plans", strconv.FormatBool(quota.NonBasicServicesAllowed)); err != nil {
			return err
		}
		if err := assertEquals(attributes, "instance_memory", strconv.Itoa(int(quota.InstanceMemoryLimit.Value))); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_memory", strconv.Itoa(int(quota.MemoryLimit.Value))); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_app_instances", strconv.Itoa(quota.AppInstanceLimit.Value)); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_services", strconv.Itoa(quota.TotalServices)); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_routes", strconv.Itoa(quota.TotalRoutes)); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_route_ports", strconv.Itoa(quota.TotalReservedRoutePorts.Value)); err != nil {
			return err
		}
		return
	}
}

func testAccCheckSpaceQuotaResourceDestroy(quotaname string, orgID string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {

		session := testAccProvider.Meta().(*managers.Session)
		quotas, _, err := session.ClientV2.GetQuotas(constant.SpaceQuota)
		if err != nil {
			return err
		}
		for _, quota := range quotas {
			if quota.Name != quotaname {
				continue
			}
			if quota.OrganizationGUID != orgID {
				continue
			}
			return fmt.Errorf("space quota with name '%s' still exists in cloud foundry", quotaname)
		}
		return nil
	}
}
