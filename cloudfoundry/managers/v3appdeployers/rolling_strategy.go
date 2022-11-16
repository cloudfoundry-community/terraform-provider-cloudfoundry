package v3appdeployers

// Rolling : Rolling strategy deployer
type Rolling struct {
	actor *Actor
}

// NewRolling initializes a rolling deployer
func NewRolling(actor *Actor) *Rolling {
	return &Rolling{
		actor: actor,
	}
}

// Deploy : deploy an app using the rolling strategy
func (s Rolling) Deploy(appDeploy AppDeploy) (AppDeployResponse, error) {
	actions := s.actor.PrepareActions([]ChangeApplicationFunction{
		s.actor.Initalize,
		s.actor.CreateApplication,
		s.actor.SetApplicationEnvironment,
		s.actor.SetApplicationSSHEnabled,
		s.actor.CreateRouteMappings,
		s.actor.BindServiceInstances,
		s.actor.CreateApplicationBitsPackage,
		s.actor.StageApplicationPackage,
		s.actor.ScaleApplicationProcess,
		s.actor.UpdateApplicationProcess,
		s.actor.CreateApplicationDeployment,
	}, appDeploy, s.actor.ReverseActionDeleteApp)

	var appResp AppDeployResponse
	ctx, err := actions.Execute()
	if appRespCtx, ok := ctx["app_response"]; ok {
		appResp = appRespCtx.(AppDeployResponse)
	}

	return appResp, err
}

// Restage : deploy an app using the rolling strategy
func (s Rolling) Restage(appDeploy AppDeploy) (AppDeployResponse, error) {
	actions := s.actor.PrepareActions([]ChangeApplicationFunction{
		s.actor.Initalize,
		s.actor.SetCurrentRevision,
		s.actor.StageApplicationPackage,
		s.actor.CreateApplicationDeployment,
	}, appDeploy, s.actor.ReverseActionDeployRevision)

	var appResp AppDeployResponse
	ctx, err := actions.Execute()
	if appRespCtx, ok := ctx["app_response"]; ok {
		appResp = appRespCtx.(AppDeployResponse)
	}

	return appResp, err
}

// Restart : restart an app without downtime
func (s Rolling) Restart(appDeploy AppDeploy) error {
	actions := s.actor.PrepareActions([]ChangeApplicationFunction{
		s.actor.Initalize,
		s.actor.SetCurrentRevision,
		s.actor.CreateApplicationDeployment,
	}, appDeploy, s.actor.ReverseActionDeployRevision)

	_, err := actions.Execute()

	return err
}

// Names : accepted aliases for this deployment strategy
func (s Rolling) Names() []string {
	return []string{"rolling"}
}

// IsCreateNewApp : true if new app is created when updating
func (s Rolling) IsCreateNewApp() bool {
	return false
}
