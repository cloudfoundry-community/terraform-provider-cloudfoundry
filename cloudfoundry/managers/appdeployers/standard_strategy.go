package appdeployers

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/bits"
)

type Standard struct {
	bitsManager *bits.BitsManager
	client      *ccv2.Client
	runBinder   *RunBinder
}

func NewStandard(bitsManager *bits.BitsManager, client *ccv2.Client, runBinder *RunBinder) *Standard {
	return &Standard{
		bitsManager: bitsManager,
		client:      client,
		runBinder:   runBinder,
	}
}

func (s Standard) Deploy(appDeploy AppDeploy) (AppDeployResponse, error) {
	stateAsk := appDeploy.App.State
	var deployFunc func(app ccv2.Application) (ccv2.Application, ccv2.Warnings, error)
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
		_, err := s.client.DeleteApplication(appResp.App.GUID)
		return err
	}
	actions := Actions{
		{
			Forward: func(ctx Context) (Context, error) {
				app := appDeploy.App
				app.State = constant.ApplicationStopped
				app, _, err := deployFunc(app)
				if err != nil {
					return ctx, err
				}
				ctx["app_response"] = AppDeployResponse{
					App: app,
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
					App:          appResp.App,
					RouteMapping: mappings,
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
					RouteMapping:    appResp.RouteMapping,
					ServiceBindings: bindings,
				}
				return ctx, nil
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				if appDeploy.Path == "" {
					return ctx, nil
				}
				appResp := ctx["app_response"].(AppDeployResponse)
				err := s.bitsManager.UploadApp(appResp.App.GUID, appDeploy.Path)
				if err != nil {
					return ctx, err
				}
				return ctx, nil
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				if stateAsk == constant.ApplicationStopped {
					return ctx, nil
				}
				appResp := ctx["app_response"].(AppDeployResponse)
				err := s.runBinder.Start(AppDeploy{
					App:          appResp.App,
					StageTimeout: appDeploy.StageTimeout,
					BindTimeout:  appDeploy.BindTimeout,
					StartTimeout: appDeploy.StartTimeout,
				})
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

func (s Standard) Restage(appDeploy AppDeploy) (AppDeployResponse, error) {
	app, _, err := s.client.RestageApplication(appDeploy.App)
	if err != nil {
		return AppDeployResponse{}, err
	}
	appDeploy.App = app

	appResp := AppDeployResponse{
		App:             app,
		RouteMapping:    appDeploy.Mappings,
		ServiceBindings: appDeploy.ServiceBindings,
	}

	err = s.runBinder.WaitStaging(appDeploy)
	if err != nil {
		return appResp, err
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

func (s Standard) IsCreateNewApp() bool {
	return false
}

func (s Standard) Names() []string {
	return []string{"standard", "v2", DefaultStrategie}
}
