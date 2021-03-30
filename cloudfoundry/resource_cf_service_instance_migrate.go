package cloudfoundry

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func resourceServiceInstanceMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found service instance Record State v0; migrating from v0 to v1")
		return migrateServiceInstanceStateV0toV1(is, meta)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateServiceInstanceStateV0toV1(is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	writer := schema.MapFieldWriter{
		Schema: resourceServiceInstance().Schema,
	}
	err := writer.WriteField([]string{"replace_on_service_plan_change"}, false)
	if err != nil {
		return is, err
	}
	err = writer.WriteField([]string{"replace_on_params_change"}, false)
	if err != nil {
		return is, err
	}
	attr := is.Attributes
	for k, v := range writer.Map() {
		attr[k] = v
	}
	is.Attributes = attr
	return is, nil
}
