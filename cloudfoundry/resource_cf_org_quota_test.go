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

const orgQuotaResource = `

resource "cloudfoundry_org_quota" "50g-org" {
  name = "50g-org"
  allow_paid_service_plans = false
  instance_memory = 2048
  total_memory = 51200
  total_app_instances = 100
  total_routes = 50
  total_services = 200
  total_route_ports = 5
}
`

const orgQuotaResourceUpdate = `

resource "cloudfoundry_org_quota" "50g-org" {
  name = "50g-org"
  allow_paid_service_plans = true
  instance_memory = 1024
  total_memory = 51200
  total_app_instances = 100
  total_routes = 100
  total_services = 150
  total_route_ports = 10
}
`

func TestAccOrgQuota_normal(t *testing.T) {

	ref := "cloudfoundry_org_quota.50g-org"
	quotaname := "50g-org"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckOrgQuotaResourceDestroy(quotaname),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: orgQuotaResource,
					Check: resource.ComposeTestCheckFunc(
						checkOrgQuotaExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "50g-org"),
						resource.TestCheckResourceAttr(
							ref, "allow_paid_service_plans", "false"),
						resource.TestCheckResourceAttr(
							ref, "instance_memory", "2048"),
						resource.TestCheckResourceAttr(
							ref, "total_memory", "51200"),
						resource.TestCheckResourceAttr(
							ref, "total_app_instances", "100"),
						resource.TestCheckResourceAttr(
							ref, "total_routes", "50"),
						resource.TestCheckResourceAttr(
							ref, "total_services", "200"),
						resource.TestCheckResourceAttr(
							ref, "total_route_ports", "5"),
					),
				},

				resource.TestStep{
					Config: orgQuotaResourceUpdate,
					Check: resource.ComposeTestCheckFunc(
						checkOrgQuotaExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", "50g-org"),
						resource.TestCheckResourceAttr(
							ref, "allow_paid_service_plans", "true"),
						resource.TestCheckResourceAttr(
							ref, "instance_memory", "1024"),
						resource.TestCheckResourceAttr(
							ref, "total_memory", "51200"),
						resource.TestCheckResourceAttr(
							ref, "total_app_instances", "100"),
						resource.TestCheckResourceAttr(
							ref, "total_routes", "100"),
						resource.TestCheckResourceAttr(
							ref, "total_services", "150"),
						resource.TestCheckResourceAttr(
							ref, "total_route_ports", "10"),
					),
				},
			},
		})
}

func checkOrgQuotaExists(resource string) resource.TestCheckFunc {

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
		if quota, err = session.QuotaManager().ReadQuota(cfapi.OrgQuota, id); err != nil {
			return
		}

		session.Log.DebugMessage(
			"quota detail read from cloud foundry '%s': %# v",
			resource, quota)

		if err := assertEquals(attributes, "name", quota.Name); err != nil {
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

func testAccCheckOrgQuotaResourceDestroy(quotaname string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		session := testAccProvider.Meta().(*cfapi.Session)
		if _, err := session.QuotaManager().FindQuotaByName(cfapi.OrgQuota, quotaname, nil); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}
		return fmt.Errorf("quota with name '%s' still exists in cloud foundry", quotaname)
	}
}
