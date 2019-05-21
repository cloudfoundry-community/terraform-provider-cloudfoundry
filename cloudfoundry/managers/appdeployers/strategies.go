package appdeployers

type Strategy interface {
	Deploy(appDeploy AppDeploy) (AppDeployResponse, error)
	Restage(appDeploy AppDeploy) (AppDeployResponse, error)
	AppUpdateNeeded() bool
	Names() []string
}
