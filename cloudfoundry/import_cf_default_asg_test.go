package cloudfoundry

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDefaultRunningAsg_importBasic(t *testing.T) {
	resourceName := "cloudfoundry_default_asg.running"
	asgs, _, err := testSession().ClientV2.GetRunningSecurityGroups()
	if err != nil {
		panic(err)
	}
	defaultLenRunningSecGroup = len(asgs)
	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckDefaultRunningAsgDestroy,
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: defaultRunningSecurityGroupResource,
				},

				resource.TestStep{
					ResourceName: resourceName,
					ImportState:  true,
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						if len(states) == 0 {
							return fmt.Errorf("There is no import state")
						}
						entity := resourceDefaultAsg()
						state := states[0]
						reader := &schema.MapFieldReader{
							Schema: entity.Schema,
							Map:    schema.BasicMapReader(state.Attributes),
						}
						result, err := reader.ReadField([]string{"asgs"})
						if err != nil {
							return err
						}
						if len(result.Value.(*schema.Set).List()) != defaultLenRunningSecGroup+2 {
							return fmt.Errorf("missing default running sec group")
						}
						return nil
					},
				},
			},
		})
}
