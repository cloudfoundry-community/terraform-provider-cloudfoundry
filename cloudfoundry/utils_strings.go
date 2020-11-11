package cloudfoundry

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
)

// resourceStringHash -
func resourceStringHash(v interface{}) int {
	return hashcode.String(v.(string))
}

func CaseDifference(_, old, new string, _ *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}
