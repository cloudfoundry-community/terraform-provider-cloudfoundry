package cloudfoundry

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
)

func resourceBuildpackMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found buildpack Record State v0; migrating from v0 to v3")
		return migrateBuildpackStateV2toV3(is, meta)
	case 2:
		log.Println("[INFO] Found buildpack Record State v2; migrating from v2 to v3")
		return migrateBuildpackStateV2toV3(is, meta)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateBuildpackStateV2toV3(is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	return migrateBitsStateV2toV3(is, meta)
}
