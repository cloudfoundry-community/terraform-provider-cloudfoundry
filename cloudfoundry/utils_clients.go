package cloudfoundry

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	constantV3 "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func IntToNullInt(v int) types.NullInt {
	return types.NullInt{
		IsSet: true,
		Value: v,
	}
}

func IntToNullUint64Zero(v int) types.NullUint64 {
	if v < 0 {
		return types.NullUint64{
			IsSet: false,
		}
	}

	return types.NullUint64{
		IsSet: true,
		Value: uint64(v),
	}
}

func BoolToNullBool(b bool) types.NullBool {
	return types.NullBool{
		IsSet: true,
		Value: b,
	}
}

func IntToNullByteSize(v int) types.NullByteSizeInMb {
	if v < 0 {
		return types.NullByteSizeInMb{
			IsSet: false,
		}
	}
	return types.NullByteSizeInMb{
		IsSet: true,
		Value: uint64(v),
	}
}

func StringToFilteredString(val string) types.FilteredString {
	if val == "" {
		return types.FilteredString{
			IsSet: false,
		}
	}
	return types.FilteredString{
		IsSet: true,
		Value: val,
	}
}

func IntToNullByteSizeZero(v int) types.NullByteSizeInMb {
	if v <= 0 {
		return types.NullByteSizeInMb{
			IsSet: false,
		}
	}
	return types.NullByteSizeInMb{
		IsSet: true,
		Value: uint64(v),
	}
}

func NullByteSizeToInt(v types.NullByteSizeInMb) int {
	if !v.IsSet {
		return -1
	}
	return int(v.Value)
}

func UsersToIDs(users []ccv2.User) []interface{} {
	ids := make([]interface{}, len(users))
	for i, u := range users {
		ids[i] = u.GUID
	}
	return ids
}

// MapToEnvironmentVariables :
// Convert map[string]string to resources.EnvironmentVariables
func MapToEnvironmentVariables(env map[string]string) resources.EnvironmentVariables {
	envVars := make(resources.EnvironmentVariables, len(env))
	for k, v := range env {
		envVars[k] = StringToFilteredString(v)
	}
	return envVars
}

// EnvironmentVariablesToMap :
// Convert resources.EnvironmentVariables to map[string]string
func EnvironmentVariablesToMap(vars resources.EnvironmentVariables) map[string]string {
	env := make(map[string]string, len(vars))
	for k, v := range vars {
		env[k] = v.String()
	}
	return env
}

func IsErrNotAuthorized(err error) bool {
	if _, ok := err.(ccerror.ForbiddenError); ok {
		return true
	}
	if httpErr, ok := err.(ccerror.RawHTTPStatusError); ok && httpErr.StatusCode == 403 {
		return true
	}
	if uaaErr, ok := err.(uaa.RawHTTPStatusError); ok && uaaErr.StatusCode == 403 {
		return true
	}
	return false
}

func IsErrNotFound(err error) bool {
	if httpErr, ok := err.(ccerror.RawHTTPStatusError); ok && httpErr.StatusCode == 404 {
		return true
	}
	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return true
	}
	if uaaErr, ok := err.(uaa.RawHTTPStatusError); ok && uaaErr.StatusCode == 404 {
		return true
	}
	return false
}

type PollingConfig struct {
	session  *managers.Session
	jobURL   ccv3.JobURL
	interval time.Duration
	timeout  time.Duration
}

// PollAsyncJob periodically check the state of the async job, return when the job failed / completed or timeout is reached
func PollAsyncJob(config PollingConfig) error {
	s := config.session

	if config.interval == 0 {
		config.interval = 5 * time.Second
	}
	if config.timeout == 0 {
		config.timeout = 60 * time.Second
	}
	return common.PollingWithTimeout(func() (bool, error) {
		job, _, err := s.ClientV3.GetJob(config.jobURL)
		if err != nil {
			return true, err
		}

		// Stop polling and return error if job failed
		if job.State == constantV3.JobFailed {
			return true, fmt.Errorf(
				"Operation failed, reason: %+v",
				job.Errors(),
			)
		}

		// Check binding state if job completed
		if job.State == constantV3.JobComplete {
			return true, nil
		}
		// Last operation initial or inprogress or job not completed, continue polling
		return false, nil
	}, config.interval, config.timeout)
}
