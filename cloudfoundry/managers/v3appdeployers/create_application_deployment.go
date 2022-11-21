package v3appdeployers

import (
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
)

// CreateApplicationDeployment : Create deployment for application
func (a Actor) CreateApplicationDeployment(appDeploy AppDeploy, reverse FallbackFunction) Action {
	return Action{
		Forward: func(ctx Context) (Context, error) {
			appResp := ctx["app_response"].(AppDeployResponse)

			var dropletGUID string
			dropletCtx := ctx["droplet"]
			if dropletCtx == nil {
				droplet, _, err := a.client.GetApplicationDropletCurrent(appResp.App.GUID)
				if err != nil {
					return ctx, err
				}
				dropletGUID = droplet.GUID
			} else {
				dropletGUID = dropletCtx.(resources.Droplet).GUID
			}

			deploymentGUID, _, err := a.client.CreateApplicationDeployment(appResp.App.GUID, dropletGUID)
			if err != nil {
				return ctx, err
			}

			err = a.PollStartRolling(appResp.App, deploymentGUID, appDeploy.StartTimeout)
			if err != nil {
				return ctx, err
			}

			app, _, err := a.client.UpdateApplicationStart(appResp.App.GUID)

			if err != nil {
				return ctx, err
			}

			appResp.App = app
			ctx["app_response"] = appResp
			ctx["deployment"] = deploymentGUID

			return ctx, nil
		},
		ReversePrevious: reverse,
	}
}

// PollStartRolling polls a deploying application's processes until some are started, accounting for rolling deployments and whether
// they have failed or been canceled during polling.
func (a Actor) PollStartRolling(app resources.Application, deploymentGUID string, startTimeout time.Duration) error {
	return common.PollingWithTimeout(func() (bool, error) {

		deployment, _, err := a.client.GetDeployment(deploymentGUID)
		if err != nil {
			return true, err
		}

		if isDeployed(deployment) {
			err := a.runBinder.WaitStart(AppDeploy{
				App:          app,
				StartTimeout: startTimeout,
			})
			if err != nil {
				return true, err
			}
			return true, nil
		}

		return false, nil
	}, 5*time.Second, startTimeout)
}

func isDeployed(d resources.Deployment) bool {
	return d.StatusValue == constant.DeploymentStatusValueFinalized && d.StatusReason == constant.DeploymentStatusReasonDeployed
}

// func (a Actor) getProcesses(deployment resources.Deployment, appGUID string) ([]resources.Process, ccv3.Warnings, error) {
// 	// if the deployment is deployed we know web are all running and PollProcesses will see those as stable
// 	// so just getting all processes is equivalent to just getting non-web ones and polling those
// 	if isDeployed(deployment) {
// 		processes, warnings, err := a.client.GetApplicationProcesses(appGUID)
// 		if err != nil {
// 			return processes, warnings, err
// 		}
// 		return processes, warnings, nil
// 	}

// 	return nil, nil, nil
// }

// PollProcesses - return true if there's no need to keep polling
func (a Actor) PollProcesses(processes []resources.Process) (bool, error) {
	numProcesses := len(processes)
	numStableProcesses := 0

	for _, process := range processes {
		instances, _, err := a.client.GetProcessInstances(process.GUID)
		if err != nil {
			return true, err
		}

		if Empty(instances) || AnyRunning(instances) {
			numStableProcesses++
			continue
		}

		if AllCrashed(instances) {
			return true, actionerror.AllInstancesCrashedError{}
		}

		// precondition: !instances.Empty() && no instances are running
		// do not increment numStableProcesses
		return false, nil
	}
	return numStableProcesses == numProcesses, nil
}

// AllCrashed : return true if all process instances crashed
func AllCrashed(pi []ccv3.ProcessInstance) bool {
	for _, instance := range pi {
		if instance.State != constant.ProcessInstanceCrashed {
			return false
		}
	}
	return true
}

// AnyRunning : return true if at least one instance is running
func AnyRunning(pi []ccv3.ProcessInstance) bool {
	for _, instance := range pi {
		if instance.State == constant.ProcessInstanceRunning {
			return true
		}
	}
	return false
}

// Empty : return true no instances
func Empty(pi []ccv3.ProcessInstance) bool {
	return len(pi) == 0
}
