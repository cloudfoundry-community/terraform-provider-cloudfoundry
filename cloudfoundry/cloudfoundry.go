package cloudfoundry

import "fmt"

// AppStatusStaging - CF status for running apps
const AppStatusStaging = "staging"

// AppStatusRunning - CF status for staging apps
const AppStatusRunning = "running"

// NewResourceMeta -
type NewResourceMeta struct {
	meta interface{}
}

// validateDefaultRunningStagingName -
func validateDefaultRunningStagingName(v interface{}, k string) (ws []string, errs []error) {
	value := v.(string)
	if value != AppStatusRunning && value != AppStatusStaging {
		errs = append(errs, fmt.Errorf("%q must be one of staging or running", k))
	}
	return ws, errs
}
