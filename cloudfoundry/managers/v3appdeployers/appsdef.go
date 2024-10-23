package v3appdeployers

import (
	"time"

	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
)

type AppDeploy struct {
	App                          resources.Application
	EnableSSH                    types.NullBool
	AppPackage                   resources.Package
	Process                      resources.Process
	Mappings                     []resources.Route
	ServiceBindings              []resources.ServiceCredentialBinding
	EnvVars                      map[string]interface{}
	Path                         string
	BindTimeout                  time.Duration
	StageTimeout                 time.Duration
	StartTimeout                 time.Duration
	BlueGreenPostStartupWaitTime time.Duration
	Ports                        []int
}

func (a AppDeploy) IsDockerImage() bool {
	return a.AppPackage.DockerImage != ""
}

type AppDeployResponse struct {
	App             resources.Application
	EnableSSH       types.NullBool
	AppPackage      resources.Package
	Process         resources.Process
	EnvVars         map[string]interface{}
	Mappings        []resources.Route
	ServiceBindings []resources.ServiceCredentialBinding
	Ports           []int
}
