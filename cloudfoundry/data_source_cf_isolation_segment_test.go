package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const isolationDataResource = `

data "cloudfoundry_isolation_segment" "segment-one" {
    name = "%s"
}
`

func TestAccDataSourceIsolationSegment_normal(t *testing.T) {

	_, defaultSegmentName := getTestDefaultIsolationSegment(t)
	ref := "data.cloudfoundry_isolation_segment.segment-one"

	resource.ParallelTest(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: fmt.Sprintf(isolationDataResource, defaultSegmentName),
					Check: resource.ComposeTestCheckFunc(
						checkDataSourceIsolationExists(ref),
						resource.TestCheckResourceAttr(
							ref, "name", defaultSegmentName),
					),
				},
			},
		})
}

func checkDataSourceIsolationExists(resource string) resource.TestCheckFunc {

	return func(s *terraform.State) error {

		session := testAccProvider.Meta().(*managers.Session)

		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("isolation segment '%s' not found in terraform state", resource)
		}

		id := rs.Primary.ID
		attributes := rs.Primary.Attributes

		segment, _, err := session.ClientV3.GetIsolationSegment(id)
		if err != nil {
			return err
		}
		err = assertEquals(attributes, "name", segment.Name)
		return err
	}
}
