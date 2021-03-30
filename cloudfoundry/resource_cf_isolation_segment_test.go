package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const segmentResource = `
resource "cloudfoundry_org" "iso-org1" {
	name = "iso-organization-one"
}
resource "cloudfoundry_org" "iso-org2" {
	name = "iso-organization-two"
}
resource "cloudfoundry_isolation_segment" "segment1" {
	name = "segment-one"
}
resource "cloudfoundry_isolation_segment_entitlement" "segment1_orgs" {
	segment = "${cloudfoundry_isolation_segment.segment1.id}"
	orgs = [
    "${cloudfoundry_org.iso-org1.id}",
    "${cloudfoundry_org.iso-org2.id}"
  ]
}
`

const segmentResourceUpdateName = `
resource "cloudfoundry_org" "iso-org1" {
	name = "iso-organization-one"
}
resource "cloudfoundry_org" "iso-org2" {
	name = "iso-organization-two"
}
resource "cloudfoundry_isolation_segment" "segment1" {
	name = "segment-one-name"
}
resource "cloudfoundry_isolation_segment_entitlement" "segment1_orgs" {
	segment = "${cloudfoundry_isolation_segment.segment1.id}"
	orgs = [
    "${cloudfoundry_org.iso-org1.id}",
    "${cloudfoundry_org.iso-org2.id}"
  ]
}
`

const segmentResourceUpdateOrgs = `
resource "cloudfoundry_org" "iso-org1" {
	name = "iso-organization-one"
}
resource "cloudfoundry_isolation_segment" "segment1" {
	name = "segment-one-name"
}
resource "cloudfoundry_isolation_segment_entitlement" "segment1_orgs" {
	segment = "${cloudfoundry_isolation_segment.segment1.id}"
	orgs = [
    	"${cloudfoundry_org.iso-org1.id}"
  	]
	default = true
}
`

var defaultLenIsolationSegments int

func TestAccResSegment_normal(t *testing.T) {
	segRef := "cloudfoundry_isolation_segment.segment1"
	entitleRef := "cloudfoundry_isolation_segment_entitlement.segment1_orgs"

	segments, _, err := testSession().ClientV3.GetIsolationSegments()
	if err != nil {
		panic(err)
	}
	defaultLenIsolationSegments = len(segments)
	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckSegmentDestroyed("segment-one-name"),
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: segmentResource,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSegmentExists(segRef, entitleRef),
						resource.TestCheckResourceAttr(segRef, "name", "segment-one"),
						resource.TestCheckResourceAttr(entitleRef, "orgs.#", "2"),
					),
				},
				resource.TestStep{
					Config: segmentResourceUpdateName,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSegmentExists(segRef, entitleRef),
						resource.TestCheckResourceAttr(segRef, "name", "segment-one-name"),
						resource.TestCheckResourceAttr(entitleRef, "orgs.#", "2"),
					),
				},
				resource.TestStep{
					Config: segmentResourceUpdateOrgs,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckSegmentExists(segRef, entitleRef),
						resource.TestCheckResourceAttr(segRef, "name", "segment-one-name"),
						resource.TestCheckResourceAttr(entitleRef, "orgs.#", "1"),
					),
				},
			},
		})
}

func testAccCheckSegmentExists(segName string, entitleName string) resource.TestCheckFunc {
	return func(s *terraform.State) (err error) {
		session := testAccProvider.Meta().(*managers.Session)
		sm := session.ClientV3

		seg, ok := s.RootModule().Resources[segName]
		if !ok {
			return fmt.Errorf("segment '%s' not found in terraform state", segName)
		}
		entitle, ok := s.RootModule().Resources[entitleName]
		if !ok {
			return fmt.Errorf("segment '%s' not found in terraform state", entitleName)
		}

		segID := seg.Primary.ID
		segAttributes := seg.Primary.Attributes
		segment, _, err := sm.GetIsolationSegment(segID)
		if err != nil {
			return fmt.Errorf("segment '%s' not found", segName)
		}

		if err = assertEquals(segAttributes, "name", segment.Name); err != nil {
			return err
		}

		entitleID := entitle.Primary.ID
		entitleAttributes := entitle.Primary.Attributes
		if entitleAttributes["segment"] != segID {
			return fmt.Errorf("entitlement resource id '%s does not match segment id '%s'", entitleID, segID)
		}

		if err = assertEquals(entitleAttributes, "segment", segID); err != nil {
			return err
		}
		orgs, _, err := sm.GetIsolationSegmentOrganizations(entitleAttributes["segment"])
		if err != nil {
			return fmt.Errorf("could fetch orgs from segment '%s'", segID)
		}
		tfOrgs := make([]interface{}, len(orgs))
		for i, org := range orgs {
			tfOrgs[i] = org.GUID
		}
		if err = assertSetEquals(entitleAttributes, "orgs", tfOrgs); err != nil {
			return err
		}
		for _, org := range orgs {
			rl, _, err := sm.GetOrganizationDefaultIsolationSegment(org.GUID)
			if err != nil {
				return err
			}
			if entitleAttributes["default"] == "1" && rl.GUID != segID {
				return fmt.Errorf("entitlement resource default isolation segment for org %s mismatch, this should be using current segment", org.Name)
			}
			if entitleAttributes["default"] == "0" && rl.GUID == segID {
				return fmt.Errorf("entitlement resource default isolation segment for org %s mismatch, this should be not using current segment", org.Name)
			}
		}
		return
	}
}

func testAccCheckSegmentDestroyed(segName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*managers.Session)
		orgs, _, err := session.ClientV3.GetIsolationSegments(ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{segName},
		})
		if err != nil {
			return err
		}
		if len(orgs) > 0 {
			return fmt.Errorf("segment with name '%s' still exists in cloud foundry", segName)
		}
		return nil
	}
}
