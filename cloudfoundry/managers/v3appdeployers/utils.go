package v3appdeployers

import (
	"fmt"
	"log"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/resources"
)

func logDebug(message string) {
	log.Printf(" **************** [INFO] %s", message)
}

func venerableAppName(appName string) string {
	return fmt.Sprintf("%s-venerable", appName)
}

func clearMappingId(mappings []resources.Route) []resources.Route {
	for i, mapping := range mappings {
		for j, destination := range mapping.Destinations {
			destination.GUID = ""
			mapping.Destinations[j] = destination
		}
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

func clearBindingId(bindings []resources.ServiceCredentialBinding) []resources.ServiceCredentialBinding {
	for i, binding := range bindings {
		binding.GUID = ""
		bindings[i] = binding
	}
	return bindings
}
