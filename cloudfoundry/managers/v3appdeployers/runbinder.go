package v3appdeployers

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/noaa"
)

type RunBinder struct {
	client     *ccv3.Client
	noaaClient *noaa.NOAAClient
}

func NewRunBinder(client *ccv3.Client, noaaClient *noaa.NOAAClient) *RunBinder {
	return &RunBinder{
		client:     client,
		noaaClient: noaaClient,
	}
}

// MapRoutes create a v3 destination for each declared route during app deployment
// Handling of custom appPort not supported (default is 8080)
func (r RunBinder) MapRoutes(appDeploy AppDeploy) ([]resources.Route, error) {
	mappings := make([]resources.Route, 0)
	appGUID := appDeploy.App.GUID
	for _, mappingCur := range appDeploy.Mappings {
		exists, err := r.mappingExists(appGUID, mappingCur)
		if err != nil {
			return mappings, err
		}
		if exists {
			mappings = append(mappings, mappingCur)
			continue
		}

		_, err = r.client.MapRoute(mappingCur.GUID, appGUID)
		if err != nil {
			return mappings, err
		}
		mappingCreated := false
		// we wait one second after mapping because mapping is an async operation which can take time to complete
		// mostly due to route emitter to perform its action inside diego
		time.Sleep(1 * time.Second)

		routeMappings, _, err := r.client.GetRouteDestinations(mappingCur.GUID)
		if err != nil {
			return mappings, err
		}

		for _, mapping := range routeMappings {
			if mapping.App.GUID == appGUID {
				mappings = append(mappings, mappingCur)
				mappingCreated = true
			}
		}

		if !mappingCreated {
			return mappings, fmt.Errorf("Failed to map route %s", mappingCur.GUID)
		}

	}
	return mappings, nil
}

func (r RunBinder) mappingExists(appGUID string, curMapping resources.Route) (bool, error) {
	mappings, _, err := r.client.GetRouteDestinations(curMapping.GUID)

	if err != nil {
		return false, err
	}

	for _, mapping := range mappings {
		if mapping.App.GUID == appGUID {
			return true, nil
		}
	}

	return false, nil
}

func (r RunBinder) bindingExists(appGUID string, binding resources.ServiceCredentialBinding) (bool, error) {
	bindings, _, err := r.client.GetServiceCredentialBindings(
		ccv3.Query{
			Key:    ccv3.QueryKey("service_instance_guids"),
			Values: []string{binding.ServiceInstanceGUID},
		}, ccv3.Query{
			Key:    ccv3.AppGUIDFilter,
			Values: []string{appGUID},
		},
	)
	if err != nil {
		return false, err
	}
	return len(bindings) > 0, nil
}

// BindServiceInstances creates service credential bindings resources for each definied service bindings in terrafrom
func (r RunBinder) BindServiceInstances(appDeploy AppDeploy) ([]resources.ServiceCredentialBinding, error) {
	bindings := make([]resources.ServiceCredentialBinding, 0)

	// log.Printf("Bind service instances : %+v", appDeploy.ServiceBindings)
	appGUID := appDeploy.App.GUID
	for _, binding := range appDeploy.ServiceBindings {
		exists, err := r.bindingExists(appGUID, binding)
		if err != nil {
			return bindings, err
		}
		if exists {
			bindings = append(bindings, binding)
			continue
		}

		// Define type = app for backing service bindings
		binding.Type = resources.AppBinding
		binding.AppGUID = appGUID

		si, _, _, err := r.client.GetServiceInstances(ccv3.Query{
			Key:    ccv3.GUIDFilter,
			Values: []string{binding.ServiceInstanceGUID},
		})
		if err != nil {
			return bindings, err
		}
		if len(si) != 1 {
			return bindings, fmt.Errorf("Error querying for the type of the service instance")
		}

		siType := si[0].Type

		// Specific binding action for user-provided service instance
		if siType == resources.UserProvidedServiceInstance {
			// Force parameters to disappear for user-provided services as this is not allowed and don't have any effect
			if len(binding.Parameters.Value) == 0 {
				binding.Parameters = types.OptionalObject{
					IsSet: false,
				}
			}

			// Create binding using specific endpoint
			createdBinding, _, err := r.client.CreateUserProvidedServiceCredentialBinding(binding)
			if err != nil {
				return bindings, err
			}

			bindings = append(bindings, createdBinding)
			continue
		}

		jobURL, _, err := r.client.CreateServiceCredentialBinding(binding)
		if err != nil {
			return bindings, err
		}

		// Poll the state of the async job
		err = common.PollingWithTimeout(func() (bool, error) {
			job, _, err := r.client.GetJob(jobURL)
			if err != nil {
				return true, err
			}

			// Stop polling and return error if job failed
			if job.State == constant.JobFailed {
				return true, fmt.Errorf(
					"Binding %s failed for app %s, reason: %+v",
					binding.Name,
					appDeploy.App.Name,
					job.Errors(),
				)
			}

			// Check binding state if job completed
			if job.State == constant.JobComplete {
				createdBindings, _, err := r.client.GetServiceCredentialBindings(
					ccv3.Query{
						Key:    ccv3.QueryKey("service_instance_guids"),
						Values: []string{binding.ServiceInstanceGUID},
					}, ccv3.Query{
						Key:    ccv3.AppGUIDFilter,
						Values: []string{appGUID},
					},
				)
				if err != nil {
					return true, err
				}

				if len(createdBindings) == 0 {
					// Job completed but service binding not found, retry
					return false, nil
				}

				if createdBindings[0].LastOperation.State == resources.OperationSucceeded {
					bindings = append(bindings, createdBindings[0])
					return true, nil
				}

				if createdBindings[0].LastOperation.State == resources.OperationFailed {
					return true, fmt.Errorf(
						"Binding %s failed for app %s, reason: %s",
						binding.Name,
						appDeploy.App.Name,
						binding.LastOperation.Description,
					)
				}
			}

			// Last operation initial or inprogress or job not completed, continue polling
			return false, nil
		}, 5*time.Second, appDeploy.BindTimeout)

		if err != nil {
			return bindings, err
		}
	}
	return bindings, nil
}

// WaitStart checks the state of each process instance
func (r RunBinder) WaitStart(appDeploy AppDeploy) error {
	return common.PollingWithTimeout(func() (bool, error) {
		process, _, err := r.client.GetApplicationProcessByType(appDeploy.App.GUID, constant.ProcessTypeWeb)
		if err != nil {
			return true, err
		}
		if appDeploy.Process.Instances.Value == 0 {
			return true, nil
		}

		instances, _, err := r.client.GetProcessInstances(process.GUID)
		if err != nil {
			return false, err
		}
		for i, instance := range instances {
			if instance.State == constant.ProcessInstanceStarting {
				continue
			}
			if instance.State == constant.ProcessInstanceRunning {
				return true, nil
			}
			if instance.State == constant.ProcessInstanceDown {
				return false, fmt.Errorf("Instance %d failed with state %s for app %s", i, instance.State, appDeploy.App.Name)
			}
			return true, fmt.Errorf("Instance %d failed with state %s for app %s", i, instance.State, appDeploy.App.Name)
		}

		return false, nil
	}, 5*time.Second, appDeploy.StartTimeout)
}

func (r RunBinder) WaitStaging(appDeploy AppDeploy) error {
	err := common.PollingWithTimeout(func() (bool, error) {
		appGUID := appDeploy.App.GUID
		appPackages, _, err := r.client.GetPackages(ccv3.Query{
			Key:    ccv3.AppGUIDFilter,
			Values: []string{appGUID},
		})
		if err != nil {
			return true, err
		}
		appPackage := appPackages[0]
		if appPackage.State == constant.PackageState(constant.BuildStaged) {
			return true, nil
		}
		if appPackage.State == constant.PackageState(constant.BuildFailed) {
			return true, fmt.Errorf(
				"Staging failed for app %s",
				appGUID,
			)
		}
		return false, nil
	}, 5*time.Second, appDeploy.StageTimeout)
	if err != nil {
		return r.processDeployErr(err, appDeploy)
	}
	return nil
}

// Start performs staging of the most recent ready package, update the app's current droplet and start the application
// http://v3-apidocs.cloudfoundry.org/version/3.124.0/#starting-apps
func (r RunBinder) Start(appDeploy AppDeploy) (resources.Application, resources.Process, error) {

	// Get newest READY package for an app
	packages, _, err := r.client.GetPackages(ccv3.Query{
		Key:    ccv3.AppGUIDFilter,
		Values: []string{appDeploy.App.GUID},
	}, ccv3.Query{
		Key:    ccv3.StatesFilter,
		Values: []string{"READY"},
	}, ccv3.Query{
		Key:    ccv3.OrderBy,
		Values: []string{"-created_at"},
	})
	if err != nil {
		return resources.Application{}, resources.Process{}, err
	}
	if len(packages) < 1 {
		return resources.Application{}, resources.Process{}, fmt.Errorf("No READY package found")
	}

	// Get first package in the list
	pkg := packages[0]
	packageGUID := pkg.GUID

	// Check for package droplets
	// If package staged, skip
	droplets, _, err := r.client.GetPackageDroplets(packageGUID,
		ccv3.Query{
			Key:    ccv3.StatesFilter,
			Values: []string{"STAGED"},
		},
	)
	if err != nil {
		return resources.Application{}, resources.Process{}, err
	}

	// Package not staged
	if len(droplets) == 0 {
		// Stage the package
		build, _, err := r.client.CreateBuild(resources.Build{
			PackageGUID: packageGUID,
		})
		if err != nil {
			return resources.Application{}, resources.Process{}, err
		}

		// Poll once every 5 sec, timeout ${stageTimeout} fixed by appDeploy
		err = common.PollingWithTimeout(func() (bool, error) {

			ccBuild, _, err := r.client.GetBuild(build.GUID)
			if err != nil {
				return true, err
			}

			if ccBuild.State == constant.BuildStaged {
				return true, nil
			}

			if ccBuild.State == constant.BuildFailed {
				return true, fmt.Errorf("Package staging failed: %s", ccBuild.Error)
			}

			return false, nil
		}, 5*time.Second, appDeploy.StageTimeout)

		if err != nil {
			return resources.Application{}, resources.Process{}, err
		}

		// Poll staged package
		completedBuild, _, err := r.client.GetBuild(build.GUID)
		if err != nil {
			return resources.Application{}, resources.Process{}, err
		}
		dropletGUID := completedBuild.DropletGUID

		// Set current droplet
		_, _, err = r.client.SetApplicationDroplet(appDeploy.App.GUID, dropletGUID)
		if err != nil {
			return resources.Application{}, resources.Process{}, err
		}
	}

	// Define process information and add to payload if set in terraform
	processScaleInfo := resources.Process{
		Type: constant.ProcessTypeWeb,
	}

	if appDeploy.Process.Instances.IsSet {
		processScaleInfo.Instances = appDeploy.Process.Instances
	}

	if appDeploy.Process.MemoryInMB.IsSet && appDeploy.Process.MemoryInMB.Value > 0 {
		processScaleInfo.MemoryInMB = appDeploy.Process.MemoryInMB
	}

	if appDeploy.Process.DiskInMB.IsSet && appDeploy.Process.DiskInMB.Value > 0 {
		processScaleInfo.DiskInMB = appDeploy.Process.DiskInMB
	}

	scaledProcess, _, err := r.client.CreateApplicationProcessScale(appDeploy.App.GUID, processScaleInfo)
	if err != nil {
		return resources.Application{}, resources.Process{}, err
	}

	// Process update info
	updatedProcess, _, err := r.client.UpdateProcess(resources.Process{
		GUID:                         scaledProcess.GUID,
		Command:                      appDeploy.Process.Command,
		HealthCheckType:              appDeploy.Process.HealthCheckType,
		HealthCheckEndpoint:          appDeploy.Process.HealthCheckEndpoint,
		HealthCheckTimeout:           appDeploy.Process.HealthCheckTimeout,
		HealthCheckInvocationTimeout: appDeploy.Process.HealthCheckInvocationTimeout,
	})
	if err != nil {
		return resources.Application{}, resources.Process{}, err
	}
	// If command = null or "", the command field in updatedProcess will be the default generated by buildpack
	// In that case, we don't set it in tfstate
	if !appDeploy.Process.Command.IsSet {
		updatedProcess.Command = appDeploy.Process.Command
	}

	// Start application
	_, _, err = r.client.UpdateApplicationStart(appDeploy.App.GUID)
	if err != nil {
		return resources.Application{}, resources.Process{}, err
	}

	err = r.WaitStart(appDeploy)
	if err != nil {
		return resources.Application{}, resources.Process{}, r.processDeployErr(err, appDeploy)
	}

	app, _, err := r.client.GetApplications(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: []string{appDeploy.App.GUID},
	})
	if err != nil {
		return resources.Application{}, resources.Process{}, err
	}
	return app[0], updatedProcess, nil
}

// Stop simply stops the application
func (r RunBinder) Stop(appDeploy AppDeploy) error {
	_, _, err := r.client.UpdateApplicationStop(appDeploy.App.GUID)
	if err != nil {
		return err
	}

	return nil
}

// Restart simply runs Stop and Start.
func (r RunBinder) Restart(appDeploy AppDeploy) (resources.Application, resources.Process, error) {
	err := r.Stop(appDeploy)
	if err != nil {
		return resources.Application{}, resources.Process{}, err
	}
	app, proc, err := r.Start(appDeploy)
	if err != nil {
		return resources.Application{}, resources.Process{}, err
	}
	return app, proc, nil
}

func (r RunBinder) processDeployErr(origErr error, appDeploy AppDeploy) error {
	var err error
	var logs string
	logs, err = r.noaaClient.RecentLogs(appDeploy.App.GUID)
	if err != nil {
		logs = fmt.Sprintf("Error occurred when recolting app %s logs: %s", appDeploy.App.Name, err.Error())
	}
	return fmt.Errorf("%s\n\nApp '%s' logs: \n%s", origErr.Error(), appDeploy.App.Name, logs)
}
