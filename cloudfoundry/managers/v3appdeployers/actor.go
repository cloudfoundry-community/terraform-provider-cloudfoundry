package v3appdeployers

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/bits"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/raw"
)

// Actor handles deployment
type Actor struct {
	bitsManager *bits.BitsManager
	client      *ccv3.Client
	rawClient   *raw.RawClient
	runBinder   *RunBinder
}

// NewActor initializes a deployer
func NewActor(bitsManager *bits.BitsManager, client *ccv3.Client, rawClient *raw.RawClient, runBinder *RunBinder) *Actor {
	return &Actor{
		bitsManager: bitsManager,
		client:      client,
		rawClient:   rawClient,
		runBinder:   runBinder,
	}
}

// Initialize : First in sequence, Initialize context
func (a Actor) Initialize(appDeploy AppDeploy, _ FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			ctx["app_response"] = AppDeployResponse{
				App:             appDeploy.App,
				EnableSSH:       appDeploy.EnableSSH,
				AppPackage:      appDeploy.AppPackage,
				Process:         appDeploy.Process,
				EnvVars:         appDeploy.EnvVars,
				Mappings:        appDeploy.Mappings,
				ServiceBindings: appDeploy.ServiceBindings,
				Ports:           appDeploy.Ports,
			}
			return ctx, nil
		},
		ReversePrevious: a.ReverseActionBlank,
	}
}

// FallbackFunction : Action to reverse to previous state if error
type FallbackFunction func(Context) error

// ChangeApplicationFunction is used in Actor.Execute
type ChangeApplicationFunction func(AppDeploy, FallbackFunction) Action
