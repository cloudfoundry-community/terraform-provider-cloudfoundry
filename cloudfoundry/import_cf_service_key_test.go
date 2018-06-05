package cloudfoundry

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccServiceKey_importBasic(t *testing.T) {
	resourceName := "cf_service_key.rabbitmq-key"

	resource.Test(t,
		resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckServiceKeyDestroyed("rabbitmq-key", "cf_service_instance.rabbitmq"),
			Steps: []resource.TestStep{

				resource.TestStep{
					Config: serviceKeyResource,
				},

				resource.TestStep{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"params"},
				},
			},
		})
}
