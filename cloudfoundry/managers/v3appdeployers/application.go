package v3appdeployers

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	constantV3 "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"github.com/cloudfoundry/go-cfclient/v3/resource" // Correct import path
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
)


// CreateApplication : return a create application action
func (a Actor) CreateApplication(appDeploy AppDeploy, reverse FallbackFunction) Action {

	return Action{
		Forward: func(ctx Context) (Context, error) {
			var deployFunc func(app resource.App) (resource.App, ccv3.Warnings, error)
			appResp := ctx["app_response"].(AppDeployResponse)

			app := appDeploy.App
			app.State = constant.ApplicationStopped

			if app.GUID != "" {
				// Assuming a.client.UpdateApplication returns ccv3.Application
				deployFunc = func(app resource.App) (resource.App, ccv3.Warnings, error) {
					appUpdate, warnings, err := a.client.UpdateApplication(app)
					if err != nil {
						return resource.App{}, warnings, err
					}
					return resource.App{}, warnings, nil // Adjust return value as needed
				}
				app.SpaceGUID = ""
			} else {
				// Assuming a.client.CreateApplication returns ccv3.Application
				deployFunc = func(app resource.App) (resource.App, ccv3.Warnings, error) {
					appCreate, warnings, err := a.client.CreateApplication(app)
					if err != nil {
						return resource.App{}, warnings, err
					}
					return resource.App{}, warnings, nil // Adjust return value as needed
				}
			}

			if appDeploy.IsDockerImage() {
				app.LifecycleType = constant.AppLifecycleTypeDocker
			}

			// Only set lifecycle type for explicit buildpack declaration
			// This is to avoid cloudcontroller error on empty buildpack name
			if bpkg := appDeploy.App.LifecycleBuildpacks; len(bpkg) > 0 && bpkg[0] != "" {
				app.LifecycleType = constant.AppLifecycleTypeBuildpack
			}

			application, _, err := deployFunc(app)
			if err != nil {
				return ctx, err
			}
			appResp.App = application

			ctx["app_response"] = appResp
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}

// StopApplication : Stop application
func (a Actor) StopApplication(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			// Action code
			_, _, err := a.client.UpdateApplicationStop(appResp.App.GUID)
			if err != nil {
				return ctx, err
			}

			ctx["app_response"] = appResp
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}

// StartApplication : Start application
func (a Actor) StartApplication(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			// Action code
			_, _, err := a.client.UpdateApplicationStart(appResp.App.GUID)
			if err != nil {
				return ctx, err
			}

			ctx["app_response"] = appResp
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}

// SetApplicationEnvironment : set application environment variable
func (a Actor) SetApplicationEnvironment(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			// Action code
			createdEnv, _, err := a.bitsManager.UpdateAppEnvironment(appResp.App.GUID, appDeploy.EnvVars)
			if err != nil {
				return ctx, err
			}
			appResp.EnvVars = createdEnv

			ctx["app_response"] = appResp
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}

// SetApplicationSSHEnabled : set enable ssh feature for application
func (a Actor) SetApplicationSSHEnabled(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			// Action code
			if appDeploy.EnableSSH.IsSet {
				_, err := a.client.UpdateAppFeature(appResp.App.GUID, appDeploy.EnableSSH.Value, "ssh")
				if err != nil {
					return ctx, err
				}
			}
			enabledSSH, _, err := a.client.GetAppFeature(appResp.App.GUID, "ssh")
			if err != nil {
				return ctx, err
			}

			appResp.EnableSSH = AppFeatureToNullBool(enabledSSH)

			ctx["app_response"] = appResp
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}

// DeleteApplicationWithPolling : delete
func (a Actor) DeleteApplicationWithPolling(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			// Action code
			jobURL, _, err := a.client.DeleteApplication(appResp.App.GUID)
			if err != nil {
				return ctx, err
			}

			err = common.PollingWithTimeout(func() (bool, error) {
				job, _, err := a.client.GetJob(jobURL)
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

				if job.State == constantV3.JobComplete {
					return true, nil
				}

				return false, nil
			}, 5*time.Second, 1*time.Minute)

			if err != nil {
				return ctx, err
			}

			ctx["app_response"] = appResp
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}
