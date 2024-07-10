package v3appdeployers

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
)

// StageApplicationPackage : stage application package
func (a Actor) StageApplicationPackage(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {

			appResp := ctx["app_response"].(AppDeployResponse)

			// Get newest READY package for an app
			pkg, err := a.GetMostRecentPackage(appResp.App)
			if err != nil {
				return ctx, err
			}
			// Get first package in the list
			packageGUID := pkg.GUID

			// Stage the package
			build, _, err := a.client.CreateBuild(ccv3.Build{
				PackageGUID: packageGUID,
			})
			if err != nil {
				return ctx, err
			}

			// Poll once every 5 sec, timeout ${stageTimeout} fixed by appDeploy
			err = a.WaitStaging(build.GUID, appDeploy.StageTimeout)

			if err != nil {
				return ctx, err
			}

			// Get staged package
			stagedBuild, _, err := a.client.GetBuild(build.GUID)
			if err != nil {
				return ctx, err
			}
			droplet, _, err := a.client.GetDroplet(stagedBuild.DropletGUID)
			if err != nil {
				return ctx, err
			}

			ctx["app_response"] = AppDeployResponse{
				App:             appResp.App,
				Process:         appResp.Process,
				ServiceBindings: appResp.ServiceBindings,
				Mappings:        appResp.Mappings,
				AppPackage:      pkg,
				EnableSSH:       appResp.EnableSSH,
				EnvVars:         appResp.EnvVars,
			}
			ctx["droplet"] = droplet

			return ctx, err
		},
		ReversePrevious: reverse,
	}
}

// GetMostRecentPackage : get ready packages ordered by creation date (reverse)
func (a Actor) GetMostRecentPackage(app resource.App) (resource.Package, error) {
	packages, _, err := a.client.GetPackages(ccv3.Query{
		Key:    ccv3.AppGUIDFilter,
		Values: []string{app.GUID},
	}, ccv3.Query{
		Key:    ccv3.StackFilter,
		Values: []string{"READY"},
	}, ccv3.Query{
		Key:    ccv3.OrderBy,
		Values: []string{"-created_at"},
	})
	if err != nil {
		return resource.Package{}, err
	}
	if len(packages) < 1 {
		return resource.Package{}, fmt.Errorf("No READY package found")
	}

	return resource.Package{}, nil
}

// WaitStaging : poll status of the created build with timeout
func (a Actor) WaitStaging(buildGUID string, stageTimeout time.Duration) error {
	return common.PollingWithTimeout(func() (bool, error) {

		ccBuild, _, err := a.client.GetBuild(buildGUID)
		if err != nil {
			return true, err
		}

		if ccBuild.State == constant.BuildStaged {
			return true, nil
		}

		if ccBuild.State == constant.BuildFailed {
			return true, fmt.Errorf("Package staging failed: %s", ccBuild.Error)
		}

		return false, nil
	}, 5*time.Second, stageTimeout)
}
