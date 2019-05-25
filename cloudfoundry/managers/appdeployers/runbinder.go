package appdeployers

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/noaa"
	"time"
)

type RunBinder struct {
	client     *ccv2.Client
	noaaClient *noaa.NOAAClient
}

func NewRunBinder(client *ccv2.Client, noaaClient *noaa.NOAAClient) *RunBinder {
	return &RunBinder{
		client:     client,
		noaaClient: noaaClient,
	}
}

func (r RunBinder) MapRoutes(appDeploy AppDeploy) ([]ccv2.RouteMapping, error) {
	mappings := make([]ccv2.RouteMapping, 0)
	appGuid := appDeploy.App.GUID

	for _, mappingCur := range appDeploy.Mappings {
		exists, err := r.mappingExists(appGuid, mappingCur)
		if err != nil {
			return mappings, err
		}
		if exists {
			mappings = append(mappings, mappingCur)
			continue
		}
		for _, port := range appDeploy.App.Ports {
			mapping, _, err := r.client.CreateRouteMapping(appGuid, mappingCur.RouteGUID, port)
			if err != nil {
				return mappings, err
			}
			mappings = append(mappings, mapping)
		}
	}
	return mappings, nil
}

func (r RunBinder) mappingExists(appGuid string, curMapping ccv2.RouteMapping) (bool, error) {
	mappings, _, err := r.client.GetRouteMappings(
		ccv2.FilterEqual(constant.RouteGUIDFilter, curMapping.RouteGUID),
		ccv2.FilterEqual(constant.AppGUIDFilter, appGuid),
	)
	if err != nil {
		return false, err
	}
	if curMapping.AppPort <= 0 {
		return len(mappings) > 0, nil
	}
	for _, mapping := range mappings {
		if mapping.AppPort == curMapping.AppPort {
			return true, nil
		}
	}
	return false, nil
}

func (r RunBinder) bindingExists(appGuid string, binding ccv2.ServiceBinding) (bool, error) {
	bindings, _, err := r.client.GetServiceBindings(
		ccv2.FilterEqual(constant.ServiceInstanceGUIDFilter, binding.ServiceInstanceGUID),
		ccv2.FilterEqual(constant.AppGUIDFilter, appGuid),
	)
	if err != nil {
		return false, err
	}
	return len(bindings) > 0, nil
}

func (r RunBinder) BindServiceInstances(appDeploy AppDeploy) ([]ccv2.ServiceBinding, error) {
	bindings := make([]ccv2.ServiceBinding, 0)
	appGuid := appDeploy.App.GUID
	for _, binding := range appDeploy.ServiceBindings {
		exists, err := r.bindingExists(appGuid, binding)
		if err != nil {
			return bindings, err
		}
		if exists {
			bindings = append(bindings, binding)
			continue
		}
		binding, _, err := r.client.CreateServiceBinding(appGuid, binding.ServiceInstanceGUID, binding.Name, true, binding.Parameters)
		if err != nil {
			return bindings, err
		}
		bindings = append(bindings, binding)
		if binding.LastOperation.State == constant.LastOperationSucceeded {
			continue
		}
		err = common.PollingWithTimeout(func() (bool, error) {
			binding, _, err := r.client.GetServiceBinding(binding.GUID)
			if err != nil {
				return true, err
			}
			if binding.LastOperation.State == constant.LastOperationSucceeded {
				return true, nil
			}
			if binding.LastOperation.State == constant.LastOperationFailed {
				return true, fmt.Errorf(
					"Binding %s failed for app %s, reason: %s",
					binding.Name,
					appDeploy.App.Name,
					binding.LastOperation.Description,
				)
			}
			return false, nil
		}, 5*time.Second, appDeploy.BindTimeout)
		if err != nil {
			return bindings, err
		}
	}
	return bindings, nil
}

func (r RunBinder) WaitStaging(appDeploy AppDeploy) error {
	err := common.PollingWithTimeout(func() (bool, error) {
		app, _, err := r.client.GetApplication(appDeploy.App.GUID)
		if err != nil {
			return true, err
		}
		if app.PackageState == constant.ApplicationPackageStaged {
			return true, nil
		}
		if app.PackageState == constant.ApplicationPackageFailed {
			return true, fmt.Errorf(
				"Staging failed for app %s, reason: %s, description: %s",
				app.Name,
				app.StagingFailedReason,
				app.StagingFailedDescription,
			)
		}
		return false, nil
	}, 5*time.Second, appDeploy.StageTimeout)
	if err != nil {
		return r.processDeployErr(err, appDeploy)
	}
	return nil
}

func (r RunBinder) Start(appDeploy AppDeploy) error {
	_, _, err := r.client.UpdateApplication(ccv2.Application{
		GUID:  appDeploy.App.GUID,
		State: constant.ApplicationStarted,
	})
	if err != nil {
		return err
	}
	err = r.WaitStaging(appDeploy)
	if err != nil {
		return err
	}
	err = common.PollingWithTimeout(func() (bool, error) {
		appInstances, _, err := r.client.GetApplicationApplicationInstances(appDeploy.App.GUID)
		if err != nil {
			return true, err
		}
		if appDeploy.App.Instances.Value == 0 {
			return true, nil
		}
		for i, instance := range appInstances {
			if instance.State == constant.ApplicationInstanceStarting {
				continue
			}
			if instance.State == constant.ApplicationInstanceRunning {
				return true, nil
			}
			return true, fmt.Errorf("Instance %d failed with state %s for app %s", i, instance.State, appDeploy.App.Name)
		}

		return false, nil
	}, 5*time.Second, appDeploy.StartTimeout)
	if err != nil {
		return r.processDeployErr(err, appDeploy)
	}
	return nil
}

func (r RunBinder) Stop(appDeploy AppDeploy) error {
	_, _, err := r.client.UpdateApplication(ccv2.Application{
		GUID:  appDeploy.App.GUID,
		State: constant.ApplicationStopped,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r RunBinder) Restart(appDeploy AppDeploy, stageTimeout time.Duration) error {
	err := r.Stop(appDeploy)
	if err != nil {
		return err
	}
	err = r.Start(appDeploy)
	if err != nil {
		return err
	}
	return nil
}

func (r RunBinder) processDeployErr(origErr error, appDeploy AppDeploy) error {
	var err error
	var logs string
	logs, err = r.noaaClient.RecentLogs(appDeploy.App.GUID)
	if err != nil {
		logs = fmt.Sprintf("Error occured when recolting app %s logs: %s", appDeploy.App.Name, err.Error())
	}
	return fmt.Errorf("%s\n\nApp '%s' logs: \n%s", origErr.Error(), appDeploy.App.Name, logs)
}
