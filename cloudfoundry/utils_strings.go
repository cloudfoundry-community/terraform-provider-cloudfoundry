package cloudfoundry

import (
	"encoding/json"
	"github.com/hashicorp/terraform/helper/hashcode"
)

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

// json.Marshal any non-string values found in data
func encodeMapJsonValues(data map[string]interface{}) (res map[string]interface{}) {
	res = make(map[string]interface{})
	for k, v := range data {
		switch v.(type) {
		case string:
			res[k] = v
		default:
			b, _ := json.Marshal(v)
			res[k] = string(b)
		}
	}
	return
}

// json.Unmarshal all values of data
func decodeMapJsonValues(data map[string]interface{}) (res map[string]interface{}) {
	res = make(map[string]interface{})
	var ival interface{}

	for k, v := range data {
		value := v.(string)
		if err := json.Unmarshal([]byte(value), &ival); err == nil {
			res[k] = ival
		} else {
			res[k] = value
		}
	}
	return
}
