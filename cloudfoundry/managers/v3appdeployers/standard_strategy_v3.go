package v3appdeployers

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	constantV3 "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	"code.cloudfoundry.org/cli/resources"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/bits"
)

// Standard is the standard deployment strategy using v3 API
type Standard struct {
	bitsManager *bits.BitsManager
	client      *ccv3.Client
	runBinder   *RunBinder
	actor       *Actor
}

// NewStandard initializes a v3 standard deployment strategy
func NewStandard(bitsManager *bits.BitsManager, client *ccv3.Client, runBinder *RunBinder, actor *Actor) *Standard {
	return &Standard{
		bitsManager: bitsManager,
		client:      client,
		runBinder:   runBinder,
		actor:       actor,
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
		// Stop app then delete
		app, _, err := s.client.UpdateApplicationStop(appResp.App.GUID)
		if err != nil {
			return err
		}

		jobURL, _, err := s.client.DeleteApplication(app.GUID)
		if err != nil {
			return err
		}

		err = common.PollingWithTimeout(func() (bool, error) {
			job, _, err := s.client.GetJob(jobURL)
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
		return err
	}
	actions := Actions{
		{
			Forward: func(ctx Context) (Context, error) {
				app := appDeploy.App
				app.State = constant.ApplicationStopped
				// If update app, remove spaceGUID request body
				// if space changes, force new ( delete + recreate )
				if app.GUID != "" {
					app.SpaceGUID = ""
				}

				if appDeploy.IsDockerImage() {
					app.LifecycleType = constant.AppLifecycleTypeDocker
				} else {
					app.LifecycleType = constant.AppLifecycleTypeBuildpack
				}

				app, _, err := deployFunc(app)
				if err != nil {
					return ctx, err
				}

				// When created, apps will always be in stopped state
				// During update, apps will be stopped and restarted after staging
				app, _, err = s.client.UpdateApplicationStop(app.GUID)
				if err != nil {
					return ctx, err
				}

				createdEnv, _, err := s.bitsManager.UpdateAppEnvironment(app.GUID, appDeploy.EnvVars)
				if err != nil {
					return ctx, err
				}

				// Set ssh_enabled
				if appDeploy.EnableSSH.IsSet {
					_, err = s.client.UpdateAppFeature(app.GUID, appDeploy.EnableSSH.Value, "ssh")
					if err != nil {
						return ctx, err
					}
				}
				enabledSSH, _, err := s.client.GetAppFeature(app.GUID, "ssh")
				if err != nil {
					return ctx, err
				}

				ctx["app_response"] = AppDeployResponse{
					App:        app,
					EnableSSH:  AppFeatureToNullBool(enabledSSH),
					EnvVars:    createdEnv,
					Process:    appDeploy.Process,
					AppPackage: appDeploy.AppPackage,
				}
				return ctx, nil
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				appResp := ctx["app_response"].(AppDeployResponse)
				mappings, err := s.runBinder.MapRoutes(AppDeploy{
					App:          appResp.App,
					Mappings:     appDeploy.Mappings,
					StageTimeout: appDeploy.StageTimeout,
					BindTimeout:  appDeploy.BindTimeout,
					StartTimeout: appDeploy.StartTimeout,
				})
				if err != nil {
					return ctx, err
				}
				ctx["app_response"] = AppDeployResponse{
					App:       appResp.App,
					EnableSSH: appResp.EnableSSH,
					EnvVars:   appResp.EnvVars,
					Mappings:  mappings,
				}
				return ctx, nil
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
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
					EnableSSH:       appResp.EnableSSH,
					EnvVars:         appResp.EnvVars,
					Mappings:        appResp.Mappings,
					ServiceBindings: bindings,
				}
				return ctx, nil
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				appResp := ctx["app_response"].(AppDeployResponse)

				var pkg resources.Package
				var err error
				if appResp.App.LifecycleType == constant.AppLifecycleTypeDocker {
					pkg, _, err = s.bitsManager.CreateDockerPackage(appResp.App.GUID, appDeploy.AppPackage.DockerImage, appDeploy.AppPackage.DockerUsername, appDeploy.AppPackage.DockerPassword)
				} else {
					// Bits will be loaded entirely into memory for each app
					// If Path = "", bits will be copied
					if appDeploy.Path == "" {
						return ctx, nil
					}
					pkg, _, err = s.bitsManager.CreateAndUploadBitsPackage(appResp.App.GUID, appDeploy.Path, appDeploy.StageTimeout)
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
			ReversePrevious: defaultReverse,
		},
		s.actor.ScaleApplicationProcess(appDeploy, defaultReverse),
		s.actor.UpdateApplicationProcess(appDeploy, defaultReverse),
		{
			Forward: func(ctx Context) (Context, error) {
				if stateAsk == constant.ApplicationStopped {
					return ctx, nil
				}

				appResp := ctx["app_response"].(AppDeployResponse)
				app, proc, err := s.runBinder.Restart(AppDeploy{
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
				// appProcess, _, err := s.client.GetApplicationProcessByType(app.GUID, constant.ProcessTypeWeb)

				ctx["app_response"] = AppDeployResponse{
					App:             app,
					Process:         proc,
					ServiceBindings: appResp.ServiceBindings,
					Mappings:        appResp.Mappings,
					AppPackage:      appResp.AppPackage,
					EnableSSH:       appResp.EnableSSH,
					EnvVars:         appResp.EnvVars,
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
	if err != nil {
		return AppDeployResponse{}, err
	}

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

	if err != nil {
		return AppDeployResponse{}, err
	}

	createdBuild, _, err := s.client.GetBuild(build.GUID)
	if err != nil {
		return AppDeployResponse{}, err
	}

	// Stop the app
	app, _, err := s.client.UpdateApplicationStop(appDeploy.App.GUID)
	if err != nil {
		return AppDeployResponse{}, err
	}
	appDeploy.App = app

	// Set droplet
	_, _, err = s.client.SetApplicationDroplet(appDeploy.App.GUID, createdBuild.DropletGUID)
	if err != nil {
		return AppDeployResponse{}, err
	}

	// Start application
	app, proc, err := s.runBinder.Start(appDeploy)
	if err != nil {
		return AppDeployResponse{}, err
	}
	appDeploy.App = app

	appResp := AppDeployResponse{
		App:             app,
		Process:         proc,
		Mappings:        appDeploy.Mappings,
		ServiceBindings: appDeploy.ServiceBindings,
		AppPackage:      appDeploy.AppPackage,
		EnableSSH:       appDeploy.EnableSSH,
		EnvVars:         appDeploy.EnvVars,
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
