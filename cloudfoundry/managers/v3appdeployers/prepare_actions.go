package v3appdeployers

// PrepareActions compile a list of ChangeApplicationFunctions to return a list of actions
func (a Actor) PrepareActions(plan []ChangeApplicationFunction, appDeploy AppDeploy, reverse FallbackFunction) Actions {
	actions := Actions{}
	for _, f := range plan {
		actions = append(actions, f(appDeploy, reverse))
	}

	return actions
}
