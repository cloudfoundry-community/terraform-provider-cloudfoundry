package cloudfoundry

import (
	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"
)

const netPolicyResource = `
data "cloudfoundry_domain" "internal" {
    name = "apps.internal"
}
data "cloudfoundry_domain" "local" {
    name = "%s"
}
resource "cloudfoundry_route" "net-policy-res-front" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "%s"
  hostname = "net-policy-res-front"
}
resource "cloudfoundry_route" "net-policy-res-back" {
  domain = "${data.cloudfoundry_domain.internal.id}"
  space = "%s"
  hostname = "net-policy-res-back"
}
resource "cloudfoundry_app" "net-policy-res-front" {
  name = "net-policy-res-front"
  buildpack = "binary_buildpack"
  space = "%s"
  memory = "64"
  disk_quota = "512"
  timeout = 1800
  path = "%s"
  environment = {
    "FORWARD" = "http://net-policy-res-back.apps.internal:8080/"
  }
  routes {
    route = "${cloudfoundry_route.net-policy-res-front.id}"
  }
}
resource "cloudfoundry_app" "net-policy-res-back" {
  name = "net-policy-res-back"
  buildpack = "binary_buildpack"
  space = "%s"
  memory = "64"
  disk_quota = "512"
  timeout = 1800
  path = "%s"

  routes {
    route = "${cloudfoundry_route.net-policy-res-back.id}"
  }
}
resource "cloudfoundry_network_policy" "policy-res" {
  policy {
    source_app = "${cloudfoundry_app.net-policy-res-front.id}"
    destination_app = "${cloudfoundry_app.net-policy-res-back.id}"
    port = "8080"
  }
  policy {
    source_app = "${cloudfoundry_app.net-policy-res-front.id}"
    destination_app = "${cloudfoundry_app.net-policy-res-back.id}"
    port = "8081"
  }
}
`

const netPolicyResourceUpdate = `
data "cloudfoundry_domain" "internal" {
    name = "apps.internal"
}
data "cloudfoundry_domain" "local" {
    name = "%s"
}
resource "cloudfoundry_route" "net-policy-res-front" {
  domain = "${data.cloudfoundry_domain.local.id}"
  space = "%s"
  hostname = "net-policy-res-front"
}
resource "cloudfoundry_route" "net-policy-res-back" {
  domain = "${data.cloudfoundry_domain.internal.id}"
  space = "%s"
  hostname = "net-policy-res-back"
}
resource "cloudfoundry_app" "net-policy-res-front" {
  name = "net-policy-res-front"
  buildpack = "binary_buildpack"
  space = "%s"
  memory = "64"
  disk_quota = "512"
  timeout = 1800
  path = "%s"
  environment = {
    "FORWARD" = "http://net-policy-res-back.apps.internal:8080/"
  }
  routes {
    route = "${cloudfoundry_route.net-policy-res-front.id}"
  }
}
resource "cloudfoundry_app" "net-policy-res-back" {
  name = "net-policy-res-back"
  buildpack = "binary_buildpack"
  space = "%s"
  memory = "64"
  disk_quota = "512"
  timeout = 1800
  path = "%s"

  routes {
    route = "${cloudfoundry_route.net-policy-res-back.id}"
  }
}
resource "cloudfoundry_network_policy" "policy-res" {
  policy {
    source_app = "${cloudfoundry_app.net-policy-res-front.id}"
    destination_app = "${cloudfoundry_app.net-policy-res-back.id}"
    port = "8081"
  }
}
`

func TestAccResNetworkPolicy_normal(t *testing.T) {

	spaceId, _ := defaultTestSpace(t)
	ref := "cloudfoundry_network_policy.policy-res"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(netPolicyResource,
						defaultAppDomain(),
						spaceId,
						spaceId,
						spaceId, asset("dummy-app-rp.zip"),
						spaceId, asset("dummy-app.zip"),
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckNetworkPoliciesExists(ref, func() (err error) {

							if err = assertHTTPResponse("https://net-policy-res-front."+defaultAppDomain(), 200, nil); err != nil {
								return err
							}
							return
						}),
					),
				},

				resource.TestStep{
					Config: fmt.Sprintf(netPolicyResourceUpdate,
						defaultAppDomain(),
						spaceId,
						spaceId,
						spaceId, asset("dummy-app-rp.zip"),
						spaceId, asset("dummy-app.zip"),
					),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckNetworkPoliciesExists(ref, func() (err error) {

							if err = assertHTTPResponse("https://net-policy-res-front."+defaultAppDomain(), 500, nil); err != nil {
								return err
							}
							return
						}),
					),
				},
			},
		})
}

func testAccCheckNetworkPoliciesExists(ref string, validate func() error) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[ref]
		if !ok {
			return fmt.Errorf("network policy '%s' not found in terraform state", ref)
		}
		attributes := rs.Primary.Attributes

		reader := &schema.MapFieldReader{
			Schema: resourceNetworkPolicy().Schema,
			Map:    schema.BasicMapReader(attributes),
		}

		result, err := reader.ReadField([]string{"policy"})
		if err != nil {
			return err
		}
		policiesTf := getListOfStructs(result.Value)
		idsMap := make(map[string]bool)
		for _, p := range policiesTf {
			idsMap[p["source_app"].(string)] = true
			idsMap[p["destination_app"].(string)] = true
		}
		ids := make([]string, 0)
		for k := range idsMap {
			ids = append(ids, k)
		}
		policies, err := session.NetClient.ListPolicies(ids...)
		if err != nil {
			return err
		}

		for _, policyTf := range policiesTf {
			inSlice := isInSlice(policies, func(object interface{}) bool {
				policy := object.(cfnetv1.Policy)
				start, end, err := portRangeParse(policyTf["port"].(string))
				if err != nil {
					// this is already validated in validate func
					// so if we have something wrong we are deeply unlucky
					panic(err)
				}
				if start != policy.Destination.Ports.Start || end != policy.Destination.Ports.End {
					return false
				}
				return policyTf["source_app"].(string) == policy.Source.ID &&
					policyTf["destination_app"].(string) == policy.Destination.ID &&
					policyTf["protocol"].(string) == string(policy.Destination.Protocol)
			})
			if !inSlice {
				return fmt.Errorf("Missing network policy %#v", policyTf)
			}
		}

		return validate()
	}
}
