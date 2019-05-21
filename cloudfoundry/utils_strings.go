package cloudfoundry

import "github.com/hashicorp/terraform/helper/hashcode"

// resourceStringHash -
func resourceStringHash(v interface{}) int {
	return hashcode.String(v.(string))
}
