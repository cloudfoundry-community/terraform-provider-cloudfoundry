package v3appdeployers

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
)

// CreateApplicationBitsPackage : Creates package for application and upload bits
func (a Actor) CreateApplicationBitsPackage(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			var pkg *resource.Package
			var err error
			if constant.AppLifecycleType(appResp.App.Lifecycle.Type) == constant.AppLifecycleTypeDocker {
				// Adjusting assignment to match the number of returned values and type
				pkg, err = a.bitsManager.CreateDockerPackage(appResp.App.GUID, appDeploy.AppPackage.Data.Docker.Image, appDeploy.AppPackage.Data.Docker.Username, appDeploy.AppPackage.Data.Docker.Password)
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
				AppPackage:      *pkg, // Dereferencing the pointer to match the expected type
			}
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}
