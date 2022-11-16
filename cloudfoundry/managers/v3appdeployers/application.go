package v3appdeployers

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
)

// CreateApplication : return a create application action
func (a Actor) CreateApplication(appDeploy AppDeploy, reverse FallbackFunction) Action {

	return Action{
		Forward: func(ctx Context) (Context, error) {

			var appResp AppDeployResponse
			appRespInterface := ctx["app_response"]
			if appRespInterface != nil {
				appResp = ctx["app_response"].(AppDeployResponse)
			} else {
				appResp = AppDeployResponse{}
			}

			app := appDeploy.App
			app.State = constant.ApplicationStopped

			if appDeploy.IsDockerImage() {
				app.LifecycleType = constant.AppLifecycleTypeDocker
			}

			// Only set lifecycle type for explicit buildpack declaration
			// This is to avoid cloudcontroller error on empty buildpack name
			if bpkg := appDeploy.App.LifecycleBuildpacks; len(bpkg) > 0 && bpkg[0] != "" {
				app.LifecycleType = constant.AppLifecycleTypeBuildpack
			}

			createdApp, _, err := a.client.CreateApplication(
				resources.Application{
					LifecycleType:       app.LifecycleType,
					LifecycleBuildpacks: app.LifecycleBuildpacks,
					StackName:           app.StackName,
					Name:                app.Name,
					SpaceGUID:           app.SpaceGUID,
				},
			)
			if err != nil {
				return ctx, err
			}
			appResp.App = createdApp

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
