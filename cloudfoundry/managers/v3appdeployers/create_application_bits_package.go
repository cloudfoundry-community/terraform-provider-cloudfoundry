package v3appdeployers

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
)

// CreateApplicationBitsPackage : Creates package for application and upload bits
func (a Actor) CreateApplicationBitsPackage(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			var pkg resources.Package
			var err error
			if appResp.App.LifecycleType == constant.AppLifecycleTypeDocker {
				pkg, _, err = a.bitsManager.CreateDockerPackage(appResp.App.GUID, appDeploy.AppPackage.DockerImage, appDeploy.AppPackage.DockerUsername, appDeploy.AppPackage.DockerPassword)
			} else {
				// Bits will be loaded entirely into memory for each app
				// If Path = "", bits will be copied
				if appDeploy.Path == "" {
					return ctx, nil
				}
				pkg, _, err = a.bitsManager.CreateAndUploadBitsPackage(appResp.App.GUID, appDeploy.Path, appDeploy.StageTimeout)
			}
			if err != nil {
				return ctx, err
			}
			ctx["app_response"] = AppDeployResponse{
				App:             appResp.App,
				EnableSSH:       appResp.EnableSSH,
				EnvVars:         appResp.EnvVars,
				Mappings:        appResp.Mappings,
				ServiceBindings: appResp.ServiceBindings,
				AppPackage:      pkg,
			}
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}
