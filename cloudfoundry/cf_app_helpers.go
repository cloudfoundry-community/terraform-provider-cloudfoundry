package cloudfoundry

import (
	"code.cloudfoundry.org/cli/resources"
)

// GetServiceBindingFromList returns the service credential binding and whether it's in the list
func GetServiceBindingFromList(guid string, bindings []resources.ServiceCredentialBinding) (resources.ServiceCredentialBinding, bool) {
	for _, binding := range bindings {
		if binding.ServiceInstanceGUID == guid {
			return binding, true
		}
	}
	return resources.ServiceCredentialBinding{}, false
}

// ReorderBindings -
func ReorderBindings(bindings []resources.ServiceCredentialBinding, currentBindings []interface{}) []resources.ServiceCredentialBinding {
	finalBindings := make([]resources.ServiceCredentialBinding, 0)
	for _, currentBindings := range currentBindings {
		if currentBindings == nil {
			continue
		}
		item := currentBindings.(map[string]interface{})
		if binding, ok := GetServiceBindingFromList(item["service_instance"].(string), bindings); ok {
			finalBindings = append(finalBindings, binding)
		}
	}
	for _, binding := range bindings {
		if _, ok := GetServiceBindingFromList(binding.ServiceInstanceGUID, finalBindings); ok {
			continue
		}
		finalBindings = append(finalBindings, binding)
	}
	return finalBindings
}

// func IsAppCodeChange(d ResourceChanger) bool {
// 	return d.HasChange("path") || d.HasChange("source_code_hash")
// }

// func IsAppUpdateOnly(d ResourceChanger) bool {
// 	if IsAppCodeChange(d) || IsAppRestageNeeded(d) || IsAppRestartNeeded(d) {
// 		return false
// 	}
// 	return d.HasChange("name") || d.HasChange("instances") ||
// 		d.HasChange("enable_ssh") || d.HasChange("stopped")
// }

// func IsAppRestageNeeded(d ResourceChanger) bool {
// 	return d.HasChange("buildpack") || d.HasChange("stack") ||
// 		d.HasChange("service_binding") || d.HasChange("environment")
// }

// func IsAppRestartNeeded(d ResourceChanger) bool {
// 	return d.HasChange("memory") || d.HasChange("disk_quota") ||
// 		d.HasChange("command") || d.HasChange("health_check_http_endpoint") ||
// 		d.HasChange("docker_image") || d.HasChange("health_check_type") ||
// 		d.HasChange("environment")
// }

// func isDiffAppParamsBinding(oldBinding, currentBinding map[string]interface{}) (bool, error) {
// 	if len(oldBinding["params"].(map[string]interface{})) != len(currentBinding["params"].(map[string]interface{})) {
// 		return true, nil
// 	}
// 	if len(oldBinding["params"].(map[string]interface{})) > 0 {
// 		oldParams := oldBinding["params"].(map[string]interface{})
// 		currentParams := currentBinding["params"].(map[string]interface{})
// 		return reflect.DeepEqual(oldParams, currentParams), nil
// 	}
// 	oldJson := oldBinding["params_json"].(string)
// 	currentJson := oldBinding["params_json"].(string)
// 	if oldJson == "" && currentJson == "" {
// 		return false, nil
// 	}
// 	if oldJson == "" && currentJson != "" || oldJson != "" && currentJson == "" {
// 		return true, nil
// 	}
// 	var oldParams map[string]interface{}
// 	var currentParams map[string]interface{}
// 	err := json.Unmarshal([]byte(oldJson), &oldParams)
// 	if err != nil {
// 		return false, err
// 	}
// 	err = json.Unmarshal([]byte(currentJson), &currentParams)
// 	if err != nil {
// 		return false, err
// 	}
// 	return reflect.DeepEqual(oldParams, currentParams), nil
// }
