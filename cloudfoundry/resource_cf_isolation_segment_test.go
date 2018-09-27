package cloudfoundry

import (
	"fmt"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
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
}
`

func TestAccSegment_normal(t *testing.T) {
	segRef := "cloudfoundry_isolation_segment.segment1"
	entitleRef := "cloudfoundry_isolation_segment_entitlement.segment1_orgs"

	resource.Test(t,
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
		session := testAccProvider.Meta().(*cfapi.Session)
		sm := session.SegmentManager()

		seg, ok := s.RootModule().Resources[segName]
		if !ok {
			return fmt.Errorf("segment '%s' not found in terraform state", segName)
		}
		entitle, ok := s.RootModule().Resources[entitleName]
		if !ok {
			return fmt.Errorf("segment '%s' not found in terraform state", entitleName)
		}
		session.Log.DebugMessage("terraform state for resource '%s': %# v", segName, seg)
		session.Log.DebugMessage("terraform state for resource '%s': %# v", entitleName, entitle)

		segID := seg.Primary.ID
		segAttributes := seg.Primary.Attributes
		segment, err := sm.ReadSegment(segID)
		if err != nil {
			return fmt.Errorf("segment '%s' not found", segName)
		}
		session.Log.DebugMessage("retrieved segment for resource '%s' with id '%s': %# v", segName, segID, segment)
		if err = assertEquals(segAttributes, "name", segment.Name); err != nil {
			return err
		}

		entitleID := entitle.Primary.ID
		entitleAttributes := entitle.Primary.Attributes
		if entitleID != segID {
			return fmt.Errorf("entitlement resource id '%s does not match segment id '%s'", entitleID, segID)
		}

		if err = assertEquals(entitleAttributes, "segment", segID); err != nil {
			return err
		}
		orgs, err := sm.GetSegmentOrgs(entitleID)
		if err != nil {
			return fmt.Errorf("could fetch orgs from segment '%s'", segID)
		}
		if err = assertSetEquals(entitleAttributes, "orgs", orgs); err != nil {
			return err
		}
		return
	}
}

func testAccCheckSegmentDestroyed(segName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*cfapi.Session)
		if _, err := session.SegmentManager().FindSegment(segName); err != nil {
			switch err.(type) {
			case *errors.ModelNotFoundError:
				return nil
			default:
				return err
			}
		}
		return fmt.Errorf("segment with name '%s' still exists in cloud foundry", segName)
	}
}
