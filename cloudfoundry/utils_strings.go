package cloudfoundry

import "github.com/hashicorp/terraform/helper/hashcode"

// isStringInList -
func isStringInList(list []string, str string) bool {
	for _, s := range list {
		if str == s {
			return true
		}
	}
	return false
}

// isStringInInterfaceList -
func isStringInInterfaceList(list []interface{}, str string) bool {
	for _, s := range list {
		if str == s.(string) {
			return true
		}
	}
	return false
}

// resourceStringHash -
func resourceStringHash(v interface{}) int {
	return hashcode.String(v.(string))
}
