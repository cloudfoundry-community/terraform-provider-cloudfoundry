package v3appdeployers

import (
	"time"

	"code.cloudfoundry.org/cli/resources"
)

type AppDeploy struct {
	App             resources.Application
	feature         resources.ApplicationFeature
	appPackage      resources.Package
	process         resources.Process
	Mappings        []resources.Route
	ServiceBindings []resources.ServiceCredentialBinding
	Path            string
	BindTimeout     time.Duration
	StageTimeout    time.Duration
	StartTimeout    time.Duration
}

func (a AppDeploy) IsDockerImage() bool {
	return a.appPackage.DockerImage != ""
}

type AppDeployResponse struct {
	App             resources.Application
	feature         resources.ApplicationFeature
	appPackage      resources.Package
	process         resources.Process
	Mappings        []resources.Route
	ServiceBindings []resources.ServiceCredentialBinding
}
