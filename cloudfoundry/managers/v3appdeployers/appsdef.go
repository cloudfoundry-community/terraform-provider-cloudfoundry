package v3appdeployers

import (
	"time"

	"code.cloudfoundry.org/cli/resources"
)

type AppDeploy struct {
	App             resources.Application
	EnableSSH       resources.ApplicationFeature
	AppPackage      resources.Package
	Process         resources.Process
	Mappings        []resources.Route
	ServiceBindings []resources.ServiceCredentialBinding
	EnvVars         resources.EnvironmentVariables
	Path            string
	BindTimeout     time.Duration
	StageTimeout    time.Duration
	StartTimeout    time.Duration
}

func (a AppDeploy) IsDockerImage() bool {
	return a.AppPackage.DockerImage != ""
}

type AppDeployResponse struct {
	App             resources.Application
	EnableSSH       resources.ApplicationFeature
	AppPackage      resources.Package
	Process         resources.Process
	EnvVars         resources.EnvironmentVariables
	Mappings        []resources.Route
	ServiceBindings []resources.ServiceCredentialBinding
}
