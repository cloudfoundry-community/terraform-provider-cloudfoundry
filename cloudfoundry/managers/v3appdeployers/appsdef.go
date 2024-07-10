package v3appdeployers

import (
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"code.cloudfoundry.org/cli/types"
)

type AppDeploy struct {
	App             resource.App
	EnableSSH       types.NullBool
	AppPackage      resource.Package
	Process         resource.Process
	Mappings        []resource.Route
	ServiceBindings []resource.ServiceCredentialBinding
	EnvVars         map[string]interface{}
	Path            string
	BindTimeout     time.Duration
	StageTimeout    time.Duration
	StartTimeout    time.Duration
	Ports           []int
}

func (a AppDeploy) IsDockerImage() bool {
	return a.AppPackage.Data.Docker.Image != ""
}

type AppDeployResponse struct {
	App             resource.App
	EnableSSH       types.NullBool
	AppPackage      resource.Package
	Process         resource.Process
	EnvVars         map[string]interface{}
	Mappings        []resource.Route
	ServiceBindings []resource.ServiceCredentialBinding
	Ports           []int
}
