package cloudfoundry

import (
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
)

// resourceStringHash -
func resourceStringHash(v interface{}) int {
	return hashcode.String(v.(string))
}

func CaseDifference(_, old, new string, _ *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}
