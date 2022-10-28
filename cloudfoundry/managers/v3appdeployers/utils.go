package v3appdeployers

import (
	"fmt"

	"code.cloudfoundry.org/cli/resources"
)

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

func clearBindingId(bindings []resources.ServiceCredentialBinding) []resources.ServiceCredentialBinding {
	for i, binding := range bindings {
		binding.GUID = ""
		bindings[i] = binding
	}
	return bindings
}
