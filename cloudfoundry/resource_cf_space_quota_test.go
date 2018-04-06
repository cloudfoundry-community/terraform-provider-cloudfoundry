package cloudfoundry

import (
	"fmt"
	"strconv"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const spaceQuotaResource = `

resource "cloudfoundry_space_quota" "10g-space" {
  name = "10g-space"
  allow_paid_service_plans = false
  instance_memory = 512
  total_memory = 10240
  total_app_instances = 10
  total_routes = 5
  total_services = 20
  total_route_ports = 1
	org = "%s"
}
`

func TestAccSpaceQuota_normal(t *testing.T) {

	ref := "cloudfoundry_space_quota.10g-space"
	quotaname := "10g-space"
	orgID := defaultPcfDevOrgID()

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSpaceQuotaResourceDestroy(quotaname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(spaceQuotaResource, orgID),
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
							ref, "total_services", "20"),
						resource.TestCheckResourceAttr(
							ref, "total_route_ports", "1"),
						resource.TestCheckResourceAttr(
							ref, "org", orgID),
					),
				},
			},
		})
}

func checkSpaceQuotaExists(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		session := testAccProvider.Meta().(*cfapi.Session)
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("quota '%s' not found in terraform state", resource)
		}

		session.Log.DebugMessage(
			"terraform state for resource '%s': %# v",
			resource, rs)

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		var quota cfapi.CCQuota
		if quota, err = session.QuotaManager().ReadQuota(cfapi.SpaceQuota, id); err != nil {
			return
		}

		session.Log.DebugMessage(
			"quota detail read from cloud foundry '%s': %# v",
			resource, quota)

		if err := assertEquals(attributes, "name", quota.Name); err != nil {
			return err
		}
		if err := assertEquals(attributes, "org", quota.OrgGUID); err != nil {
			return err
		}
		if err := assertEquals(attributes, "allow_paid_service_plans", strconv.FormatBool(quota.NonBasicServicesAllowed)); err != nil {
			return err
		}
		if err := assertEquals(attributes, "instance_memory", strconv.Itoa(int(quota.InstanceMemoryLimit))); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_memory", strconv.Itoa(int(quota.MemoryLimit))); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_app_instances", strconv.Itoa(quota.AppInstanceLimit)); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_services", strconv.Itoa(quota.TotalServices)); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_routes", strconv.Itoa(quota.TotalRoutes)); err != nil {
			return err
		}
		if err := assertEquals(attributes, "total_route_ports", strconv.Itoa(quota.TotalReserveredPorts)); err != nil {
			return err
		}
		return
	}
}

func testAccCheckSpaceQuotaResourceDestroy(quotaname string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		session := testAccProvider.Meta().(*cfapi.Session)
		org := defaultPcfDevOrgID()
		if _, err := session.QuotaManager().FindQuotaByName(cfapi.SpaceQuota, quotaname, &org); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}
		return fmt.Errorf("space quota with name '%s' still exists in cloud foundry", quotaname)
	}
}
