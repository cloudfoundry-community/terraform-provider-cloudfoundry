package v3appdeployers

// SetCurrentRevision : set revision for fallback
func (a Actor) SetCurrentRevision(appDeploy AppDeploy, _ FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			// Action code
			revisions, _, err := a.client.GetApplicationRevisionsDeployed(appResp.App.GUID)
			if err != nil {
				return ctx, err
			}

			if len(revisions) != 0 {
				ctx["revisions"] = revisions[0].GUID
			}

			ctx["app_response"] = appResp
			return ctx, nil
		},
		ReversePrevious: a.ReverseActionBlank,
	}
}
