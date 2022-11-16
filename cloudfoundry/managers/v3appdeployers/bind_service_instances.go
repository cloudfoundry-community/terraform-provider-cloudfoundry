package v3appdeployers

// BindServiceInstances : bind service instances to application
func (a Actor) BindServiceInstances(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			bindings, err := a.runBinder.BindServiceInstances(AppDeploy{
				App:             appResp.App,
				ServiceBindings: appDeploy.ServiceBindings,
				StageTimeout:    appDeploy.StageTimeout,
				BindTimeout:     appDeploy.BindTimeout,
				StartTimeout:    appDeploy.StartTimeout,
			})
			if err != nil {
				return ctx, err
			}

			appResp.ServiceBindings = bindings
			ctx["app_response"] = appResp
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}
