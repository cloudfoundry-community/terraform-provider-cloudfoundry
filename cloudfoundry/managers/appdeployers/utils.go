package appdeployers

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

func venerableAppName(appName string) string {
	return fmt.Sprintf("%s-venerable", appName)
}

func clearMappingId(mappings []ccv2.RouteMapping) []ccv2.RouteMapping {
	for i, mapping := range mappings {
		mapping.GUID = ""
		mappings[i] = mapping
	}
	return mappings
}

func rejoinMappingPort(defaultPort int, mappings []ccv2.RouteMapping) []ccv2.RouteMapping {
	for i, mapping := range mappings {
		if mapping.AppPort <= 0 {
			mapping.AppPort = defaultPort
		}
		mappings[i] = mapping
	}
	return mappings
}

func clearBindingId(bindings []ccv2.ServiceBinding) []ccv2.ServiceBinding {
	for i, binding := range bindings {
		binding.GUID = ""
		bindings[i] = binding
	}
	return bindings
}
