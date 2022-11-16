package v3appdeployers

type Strategy interface {
	Deploy(appDeploy AppDeploy) (AppDeployResponse, error)
	Restage(appDeploy AppDeploy) (AppDeployResponse, error)

	IsCreateNewApp() bool
	Names() []string
}

// CustomRestartStrategy : Stategies that implement their own custom restart function
type CustomRestartStrategy interface {
	Restart(appDeploy AppDeploy) error
}
