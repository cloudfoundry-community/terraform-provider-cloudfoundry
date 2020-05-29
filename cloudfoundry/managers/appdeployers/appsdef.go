package appdeployers

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type AppDeploy struct {
	App             ccv2.Application
	Mappings        []ccv2.RouteMapping
	ServiceBindings []ccv2.ServiceBinding
	Path            string
	BindTimeout     time.Duration
	StageTimeout    time.Duration
	StartTimeout    time.Duration
}

func (a AppDeploy) IsDockerImage() bool {
	return a.App.DockerImage != ""
}

type AppDeployResponse struct {
	App             ccv2.Application
	RouteMapping    []ccv2.RouteMapping
	ServiceBindings []ccv2.ServiceBinding
}
