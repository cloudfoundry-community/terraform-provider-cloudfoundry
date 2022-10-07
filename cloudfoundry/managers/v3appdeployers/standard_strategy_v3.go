package v3appdeployers

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/bits"
)

// Standard is the standard deployment strategy using v3 API
type Standard struct {
	bitsManager *bits.BitsManager
	client      *ccv3.Client
	runBinder   *RunBinder
}

// NewStandard initializes a v3 standard deployment strategy
func NewStandard(bitsManager *bits.BitsManager, client *ccv3.Client, runBinder *RunBinder) *Standard {
	return &Standard{
		bitsManager: bitsManager,
		client:      client,
		runBinder:   runBinder,
	}
}

// Deploy an application using standard deployment strategy
func (s Standard) Deploy(appDeploy AppDeploy) (AppDeployResponse, error) {
	stateAsk := appDeploy.App.State
	var deployFunc func(app resources.Application) (resources.Application, ccv3.Warnings, error)
	if appDeploy.App.GUID != "" {
		deployFunc = s.client.UpdateApplication
	} else {
		deployFunc = s.client.CreateApplication
	}
	defaultReverse := func(ctx Context) error {
		appResp := ctx["app_response"].(AppDeployResponse)
		if appResp.App.GUID == "" {
			return nil
		}
		_, _, err := s.client.DeleteApplication(appResp.App.GUID)
		return err
	}
	actions := Actions{
		{
			Forward: func(ctx Context) (Context, error) {
				logDebug(fmt.Sprintf("Create/Update app: %+v", appDeploy.App))

				app := appDeploy.App
				app.State = constant.ApplicationStopped
				app, _, err := deployFunc(app)
				if err != nil {
					return ctx, err
				}
				ctx["app_response"] = AppDeployResponse{
					App:        app,
					Process:    appDeploy.Process,
					EnableSSH:  appDeploy.EnableSSH,
					AppPackage: appDeploy.AppPackage,
					EnvVars:    appDeploy.EnvVars,
				}
				return ctx, nil
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				logDebug(fmt.Sprintf("Map routes: %+v", appDeploy.Mappings))

				// Check for processes to see if we can scale to correct memory and nb instances
				appProcesses, _, err := s.client.GetApplicationProcesses(appDeploy.App.GUID)
				logDebug(fmt.Sprintf("app processes : %+v", appProcesses))

				appResp := ctx["app_response"].(AppDeployResponse)
				mappings, err := s.runBinder.MapRoutes(AppDeploy{
					App:          appResp.App,
					Mappings:     appDeploy.Mappings,
					StageTimeout: appDeploy.StageTimeout,
					BindTimeout:  appDeploy.BindTimeout,
					StartTimeout: appDeploy.StartTimeout,
					Process:      appDeploy.Process,
					EnableSSH:    appDeploy.EnableSSH,
					AppPackage:   appDeploy.AppPackage,
					EnvVars:      appDeploy.EnvVars,
				})
				if err != nil {
					return ctx, err
				}
				ctx["app_response"] = AppDeployResponse{
					App:      appResp.App,
					Mappings: mappings,
				}
				return ctx, nil
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				logDebug(fmt.Sprintf("Bind service instance %+v", appDeploy.ServiceBindings))

				appResp := ctx["app_response"].(AppDeployResponse)
				bindings, err := s.runBinder.BindServiceInstances(AppDeploy{
					App:             appResp.App,
					ServiceBindings: appDeploy.ServiceBindings,
					StageTimeout:    appDeploy.StageTimeout,
					BindTimeout:     appDeploy.BindTimeout,
					StartTimeout:    appDeploy.StartTimeout,
				})
				if err != nil {
					return ctx, err
				}
				ctx["app_response"] = AppDeployResponse{
					App:             appResp.App,
					Mappings:        appResp.Mappings,
					ServiceBindings: bindings,
				}
				return ctx, nil
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				logDebug("Uploading package")
				if appDeploy.Path == "" {
					return ctx, nil
				}
				appResp := ctx["app_response"].(AppDeployResponse)
				pkg, _, err := s.bitsManager.CreateAndUploadBitsPackage(appResp.App.GUID, appDeploy.Path, appDeploy.StageTimeout)
				if err != nil {
					return ctx, err
				}
				ctx["app_response"] = AppDeployResponse{
					App:             appResp.App,
					Mappings:        appResp.Mappings,
					ServiceBindings: appResp.ServiceBindings,
					AppPackage:      pkg,
				}
				return ctx, nil
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				logDebug(fmt.Sprintf("Start application %+v", appDeploy))
				if stateAsk == constant.ApplicationStopped {
					return ctx, nil
				}

				appResp := ctx["app_response"].(AppDeployResponse)
				app, err := s.runBinder.Start(AppDeploy{
					App:          appResp.App,
					Process:      appDeploy.Process,
					EnableSSH:    appDeploy.EnableSSH,
					AppPackage:   appDeploy.AppPackage,
					EnvVars:      appDeploy.EnvVars,
					StageTimeout: appDeploy.StageTimeout,
					BindTimeout:  appDeploy.BindTimeout,
					StartTimeout: appDeploy.StartTimeout,
				})
				if err != nil {
					return ctx, err
				}

				// Get process information
				appProcess, _, err := s.client.GetApplicationProcessByType(app.GUID, constant.ProcessTypeWeb)
				logDebug(fmt.Sprintf("app and app web process : %+v, %+v", app, appProcess))

				ctx["app_response"] = AppDeployResponse{
					App:             app,
					Mappings:        appResp.Mappings,
					ServiceBindings: appResp.ServiceBindings,
					Process:         appProcess,
					EnableSSH:       appDeploy.EnableSSH,
					AppPackage:      appResp.AppPackage,
					EnvVars:         appDeploy.EnvVars,
				}
				return ctx, err
			},
			ReversePrevious: defaultReverse,
		},
	}
	var appResp AppDeployResponse
	ctx, err := actions.Execute()
	if appRespCtx, ok := ctx["app_response"]; ok {
		appResp = appRespCtx.(AppDeployResponse)
	}
	if stateAsk == constant.ApplicationStopped || err != nil {
		appResp.App.State = constant.ApplicationStopped
	} else {
		appResp.App.State = constant.ApplicationStarted
	}

	return appResp, err
}

// Restage is deprecated in v3 and replaced by the build resource
func (s Standard) Restage(appDeploy AppDeploy) (AppDeployResponse, error) {

	// Get newest READY package for an app
	packages, _, err := s.client.GetPackages(ccv3.Query{
		Key:    ccv3.AppGUIDFilter,
		Values: []string{appDeploy.App.GUID},
	}, ccv3.Query{
		Key:    ccv3.StatesFilter,
		Values: []string{"READY"},
	}, ccv3.Query{
		Key:    ccv3.OrderBy,
		Values: []string{"-created_at"},
	})

	// Stage the package
	build, _, err := s.client.CreateBuild(resources.Build{
		PackageGUID: packages[0].GUID,
	})
	if err != nil {
		return AppDeployResponse{}, err
	}

	err = common.PollingWithTimeout(func() (bool, error) {

		ccBuild, _, err := s.client.GetBuild(build.GUID)
		if err != nil {
			return true, err
		}

		if ccBuild.State == constant.BuildStaged {
			return true, nil
		}

		if ccBuild.State == constant.BuildFailed {
			return true, fmt.Errorf("Package staging failed")
		}

		return false, nil
	}, 5*time.Second, appDeploy.StageTimeout)

	// Stop the app
	app, _, err := s.client.UpdateApplicationStop(appDeploy.App.GUID)
	if err != nil {
		return AppDeployResponse{}, err
	}

	// Set droplet
	_, _, err = s.client.SetApplicationDroplet(appDeploy.App.GUID, build.DropletGUID)
	if err != nil {
		return AppDeployResponse{}, err
	}

	// Start application
	app, _, err = s.client.UpdateApplicationStart(appDeploy.App.GUID)
	if err != nil {
		return AppDeployResponse{}, err
	}

	appDeploy.App = app
	appResp := AppDeployResponse{
		App:             app,
		Mappings:        appDeploy.Mappings,
		ServiceBindings: appDeploy.ServiceBindings,
	}

	err = s.runBinder.WaitStart(appDeploy)
	if err != nil {
		return appResp, err
	}
	if appDeploy.App.State == constant.ApplicationStopped {
		err := s.runBinder.Stop(appDeploy)
		return appResp, err
	}

	return appResp, nil
}

// IsCreateNewApp returns false since we are not creating a new app during rolling deployment
func (s Standard) IsCreateNewApp() bool {
	return false
}

// Names defines acceptable names that can be passed to terraform
func (s Standard) Names() []string {
	return []string{"standard", "v2", DefaultStrategy}
}
