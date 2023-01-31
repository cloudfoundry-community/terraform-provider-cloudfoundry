package ccerror

import "fmt"

// V2JobFailedError represents a failed Cloud Controller Job. It wraps the error
// returned back from the Cloud Controller.
type V2JobFailedError struct {
	JobGUID string
	Message string
}

func (e V2JobFailedError) Error() string {
	return fmt.Sprintf("Job (%s) failed: %s", e.JobGUID, e.Message)
}
