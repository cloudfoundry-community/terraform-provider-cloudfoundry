package v3appdeployers

// CreateRouteMappings : create route mappings
func (a Actor) CreateRouteMappings(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			mappings, err := a.runBinder.MapRoutes(AppDeploy{
				App:          appResp.App,
				Mappings:     appDeploy.Mappings,
				StageTimeout: appDeploy.StageTimeout,
				BindTimeout:  appDeploy.BindTimeout,
				StartTimeout: appDeploy.StartTimeout,
			})
			if err != nil {
				return ctx, err
			}

			appResp.Mappings = mappings
			ctx["app_response"] = appResp
			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}
