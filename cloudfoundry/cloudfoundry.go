package cloudfoundry

import "fmt"

// NewResourceMeta -
type NewResourceMeta struct {
	meta interface{}
}

// validateDefaultRunningStagingName -
func validateDefaultRunningStagingName(v interface{}, k string) (ws []string, errs []error) {
	value := v.(string)
	if value != "running" && value != "staging" {
		errs = append(errs, fmt.Errorf("%q must be one of staging or running", k))
	}
	return
}
