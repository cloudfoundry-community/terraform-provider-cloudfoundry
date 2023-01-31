package cloudcontroller

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"github.com/blang/semver"
)

// MinimumAPIVersionCheck compares `current` to `minimum`.  If `current` is
// older than `minimum` then an error is returned; otherwise, nil is returned.
func MinimumAPIVersionCheck(current string, minimum string) error {
	if minimum == "" {
		return nil
	}

	currentSemver, err := semver.Make(current)
	if err != nil {
		return err
	}

	minimumSemver, err := semver.Make(minimum)
	if err != nil {
		return err
	}

	if currentSemver.Compare(minimumSemver) == -1 {
		return ccerror.MinimumAPIVersionNotMetError{
			CurrentVersion: current,
			MinimumVersion: minimum,
		}
	}

	return nil
}
