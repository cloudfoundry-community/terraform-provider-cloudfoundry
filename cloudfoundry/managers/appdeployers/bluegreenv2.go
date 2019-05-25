package appdeployers

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/bits"
)

type BlueGreenV2 struct {
	bitsManager *bits.BitsManager
	client      *ccv2.Client
	runBinder   *RunBinder
	standard    *Standard
}

func NewBlueGreenV2(bitsManager *bits.BitsManager, client *ccv2.Client, runBinder *RunBinder, standard *Standard) *BlueGreenV2 {
	return &BlueGreenV2{
		bitsManager: bitsManager,
		client:      client,
		runBinder:   runBinder,
		standard:    standard,
	}
}

func (s BlueGreenV2) Deploy(appDeploy AppDeploy) (AppDeployResponse, error) {
	if appDeploy.App.State == constant.ApplicationStopped || appDeploy.App.GUID == "" {
		return s.standard.Deploy(appDeploy)
	}
	appDeploy.Mappings = clearMappingId(appDeploy.Mappings)
	appDeploy.ServiceBindings = clearBindingId(appDeploy.ServiceBindings)
	actions := Actions{
		{
			Forward: func(ctx Context) (Context, error) {
				_, _, err := s.client.UpdateApplication(ccv2.Application{
					GUID: appDeploy.App.GUID,
					Name: venerableAppName(appDeploy.App.Name),
				})
				return ctx, err
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				app := appDeploy.App
				app.GUID = ""
				appResp, err := s.standard.Deploy(AppDeploy{
					App:             app,
					ServiceBindings: appDeploy.ServiceBindings,
					Mappings:        appDeploy.Mappings,
					Path:            appDeploy.Path,
					StageTimeout:    appDeploy.StageTimeout,
					BindTimeout:     appDeploy.BindTimeout,
					StartTimeout:    appDeploy.StartTimeout,
				})
				ctx["app_response"] = appResp
				return ctx, err
			},
			ReversePrevious: func(ctx Context) error {
				appResp := ctx["app_response"].(AppDeployResponse)
				if appResp.App.GUID != "" {
					_, err := s.client.DeleteApplication(appResp.App.GUID)
					if err != nil {
						if httpErr, ok := err.(ccerror.RawHTTPStatusError); !ok || httpErr.StatusCode != 404 {
							return err
						}
					}
				}
				_, _, err := s.client.UpdateApplication(ccv2.Application{
					GUID: appDeploy.App.GUID,
					Name: appDeploy.App.Name,
				})
				return err
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				_, err := s.client.DeleteApplication(appDeploy.App.GUID)
				return ctx, err
			},
		},
	}
	ctx, err := actions.Execute()
	if err != nil {
		return AppDeployResponse{}, err
	}
	return ctx["app_response"].(AppDeployResponse), nil
}

func (s BlueGreenV2) Restage(appDeploy AppDeploy) (AppDeployResponse, error) {
	if appDeploy.App.State == constant.ApplicationStopped {
		return s.standard.Restage(appDeploy)
	}
	appDeploy.Mappings = clearMappingId(appDeploy.Mappings)
	appDeploy.ServiceBindings = clearBindingId(appDeploy.ServiceBindings)
	defaultReverse := func(ctx Context) error {
		appResp := ctx["app_response"].(AppDeployResponse)
		if appResp.App.GUID != "" {
			_, err := s.client.DeleteApplication(appResp.App.GUID)
			if err != nil {
				return err
			}
		}
		_, _, err := s.client.UpdateApplication(ccv2.Application{
			GUID: appDeploy.App.GUID,
			Name: appDeploy.App.Name,
		})
		return err
	}
	actions := Actions{
		{
			Forward: func(ctx Context) (Context, error) {
				_, _, err := s.client.UpdateApplication(ccv2.Application{
					GUID: appDeploy.App.GUID,
					Name: venerableAppName(appDeploy.App.Name),
				})
				return ctx, err
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				app := appDeploy.App
				app.GUID = ""
				app.State = constant.ApplicationStopped
				appResp, err := s.standard.Deploy(AppDeploy{
					App:             app,
					ServiceBindings: appDeploy.ServiceBindings,
					Mappings:        appDeploy.Mappings,
					Path:            "",
					StageTimeout:    appDeploy.StageTimeout,
					BindTimeout:     appDeploy.BindTimeout,
					StartTimeout:    appDeploy.StartTimeout,
				})
				ctx["app_response"] = appResp
				return ctx, err
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				appResp := ctx["app_response"].(AppDeployResponse)
				err := s.bitsManager.CopyApp(appDeploy.App.GUID, appResp.App.GUID)
				return ctx, err
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
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
		{
			Forward: func(ctx Context) (Context, error) {
				_, err := s.client.DeleteApplication(appDeploy.App.GUID)
				return ctx, err
			},
		},
	}
	ctx, err := actions.Execute()
	if err != nil {
		return AppDeployResponse{}, err
	}
	return ctx["app_response"].(AppDeployResponse), nil
}

func (BlueGreenV2) IsCreateNewApp() bool {
	return true
}

func (BlueGreenV2) Names() []string {
	return []string{"blue-green", "blue-green-v2"}
}
