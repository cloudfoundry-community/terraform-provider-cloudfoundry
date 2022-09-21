package cloudfoundry

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/v3appdeployers"
)

// ResourceDataToAppDeployV3 convert tfstate to AppDeploy structure
func ResourceDataToAppDeployV3(d *schema.ResourceData) (v3appdeployers.AppDeploy, error) {
	app := resources.Application{
		GUID:                d.Id(),
		Name:                d.Get("name").(string),
		State:               constant.ApplicationStarted,
		SpaceGUID:           d.Get("space").(string),
		LifecycleBuildpacks: []string{d.Get("buildpack").(string)},
		StackName:           d.Get("stack").(string),
	}

	proc := resources.Process{
		Command:             StringToFilteredString(d.Get("command").(string)),
		Instances:           IntToNullInt(d.Get("instances").(int)),
		MemoryInMB:          IntToNullUint64Zero(d.Get("memory").(int)),
		DiskInMB:            IntToNullUint64Zero(d.Get("disk_quota").(int)),
		HealthCheckEndpoint: d.Get("health_check_http_endpoint").(string),
		HealthCheckType:     constant.HealthCheckType((d.Get("health_check_type").(string))),
		HealthCheckTimeout:  int64(d.Get("health_check_timeout").(int)),
	}

	pkg := resources.Package{
		DockerImage: d.Get("docker_image").(string),
	}

	enableSSH := resources.ApplicationFeature{
		Name:    "enable_ssh",
		Enabled: d.Get("enable_ssh").(bool),
	}

	if d.Get("stopped").(bool) {
		app.State = constant.ApplicationStopped
	}
	ports := make([]int, 0)
	for _, vv := range d.Get("ports").(*schema.Set).List() {
		ports = append(ports, vv.(int))
	}
	if len(ports) == 0 && pkg.DockerImage == "" {
		ports = []int{8080}
	}

	if v, ok := d.GetOk("docker_credentials"); ok {
		vv := v.(map[string]interface{})
		pkg.DockerUsername = vv["username"].(string)
		pkg.DockerPassword = vv["password"].(string)
	}

	envVars := make(resources.EnvironmentVariables)
	if v, ok := d.GetOk("environment"); ok {
		vv := v.(map[string]interface{})
		for k, v := range vv {
			envVars[k] = *types.NewFilteredString(fmt.Sprint(v))
		}
	}

	mappings := make([]resources.Route, 0)
	for _, r := range getListOfStructs(d.Get("routes")) {
		mappings = append(mappings, resources.Route{
			GUID: r["route"].(string),
			// AppPort is not supported in ccv3 route mapping
			Port: r["port"].(int),
		})
	}

	bindings := make([]resources.ServiceCredentialBinding, 0)
	for _, r := range getListOfStructs(d.Get("service_binding")) {
		params := r["params"].(map[string]interface{})
		paramJSON := r["params_json"].(string)
		if len(params) == 0 && paramJSON != "" {
			params = make(map[string]interface{})
			err := json.Unmarshal([]byte(paramJSON), &params)
			if err != nil {
				return v3appdeployers.AppDeploy{}, err
			}
		}

		bindings = append(bindings, resources.ServiceCredentialBinding{
			ServiceInstanceGUID: r["service_instance"].(string),
			Parameters:          types.NewOptionalObject(params),
		})
	}

	appDeploy := v3appdeployers.AppDeploy{
		App:             app,
		Process:         proc,
		AppPackage:      pkg,
		ServiceBindings: bindings,
		EnableSSH:       enableSSH,
		Mappings:        mappings,
		Path:            d.Get("path").(string),
		StartTimeout:    time.Duration(d.Get("timeout").(int)) * time.Second,
		BindTimeout:     DefaultBindTimeout,
		StageTimeout:    DefaultStageTimeout,
	}

	log.Printf("--------- [INFO] Parsed app deploy %+v", appDeploy)

	return appDeploy, nil
}

// AppDeployV3ToResourceData convert AppDeploy structure to tfstate
func AppDeployV3ToResourceData(d *schema.ResourceData, appDeploy v3appdeployers.AppDeployResponse) {
	d.SetId(appDeploy.App.GUID)
	log.Printf("--------- [INFO] Appdeploy resp to parse %+v", appDeploy)

	_ = d.Set("name", appDeploy.App.Name)
	_ = d.Set("space", appDeploy.App.SpaceGUID)
	// _ = d.Set("ports", appDeploy.App.Ports)
	_ = d.Set("instances", appDeploy.Process.Instances.Value)
	_ = d.Set("memory", appDeploy.Process.MemoryInMB.Value)
	_ = d.Set("disk_quota", appDeploy.Process.DiskInMB.Value)
	_ = d.Set("stack", appDeploy.App.StackName)

	// Ensure buildpacks has an elem
	if bpkg := appDeploy.App.LifecycleBuildpacks; len(bpkg) > 0 {
		_ = d.Set("buildpack", bpkg[0])
	}

	_ = d.Set("command", appDeploy.Process.Command.Value)
	_ = d.Set("enable_ssh", appDeploy.EnableSSH.Enabled)
	_ = d.Set("stopped", appDeploy.App.State == constant.ApplicationStopped)
	_ = d.Set("docker_image", appDeploy.AppPackage.DockerImage)
	_ = d.Set("health_check_http_endpoint", appDeploy.Process.HealthCheckEndpoint)
	_ = d.Set("health_check_type", string(appDeploy.Process.HealthCheckType))
	_ = d.Set("health_check_timeout", int(appDeploy.Process.HealthCheckTimeout))
	_ = d.Set("environment", appDeploy.EnvVars)
	// Ensure id_bg is set
	if idBg, ok := d.GetOk("id_bg"); !ok || idBg == "" {
		_ = d.Set("id_bg", d.Id())
	}

	bindingsTf := getListOfStructs(d.Get("service_binding"))
	finalBindings := make([]map[string]interface{}, 0)
	for _, binding := range appDeploy.ServiceBindings {
		if IsImportState(d) {
			b, _ := json.Marshal(binding.Parameters)
			finalBindings = append(finalBindings, map[string]interface{}{
				"service_instance": binding.ServiceInstanceGUID,
				"params_json":      string(b),
			})
			continue
		}
		curBindingsRaw, ok := getInSlice(bindingsTf, func(object interface{}) bool {
			objMap := object.(map[string]interface{})
			return objMap["service_instance"] == binding.ServiceInstanceGUID
		})
		if !ok {
			continue
		}
		for _, curBindingRaw := range curBindingsRaw {
			curBinding := curBindingRaw.(map[string]interface{})
			if len(binding.Parameters.Value) > 0 && len(curBinding["params"].(map[string]interface{})) > 0 {
				curBinding["params"] = binding.Parameters
			}
			if len(binding.Parameters.Value) > 0 && (curBinding["params_json"].(string) != "" || len(curBinding["params"].(map[string]interface{})) == 0) {
				// error can't happen and skip it when sure there is no error is the way of life in go
				b, _ := json.Marshal(binding.Parameters)
				curBinding["params_json"] = string(b)
			}
			curBinding["service_instance"] = binding.ServiceInstanceGUID
			finalBindings = append(finalBindings, curBinding)
		}
	}
	_ = d.Set("service_binding", finalBindings)

	mappingsTf := getListOfStructs(d.Get("routes"))
	finalMappings := make([]map[string]interface{}, 0)
	for _, mapping := range appDeploy.Mappings {
		// if 0 it mean app port has been set to null which means it takes the first port found in app port definition
		// if mapping.AppPort <= 0 {
		// 	mapping.AppPort = appDeploy.App.Ports[0]
		// }
		if IsImportState(d) {
			finalMappings = append(finalMappings, map[string]interface{}{
				"route": mapping.GUID,
				"port":  mapping.Port,
			})
			continue
		}
		curMappingsRaw, ok := getInSlice(mappingsTf, func(object interface{}) bool {
			objMap := object.(map[string]interface{})
			return objMap["route"] == mapping.GUID
		})
		if !ok {
			continue
		}
		for _, curMappingRaw := range curMappingsRaw {
			curMapping := curMappingRaw.(map[string]interface{})
			curMapping["route"] = mapping.GUID
			curMapping["port"] = mapping.Port
			finalMappings = append(finalMappings, curMapping)
		}

	}
	_ = d.Set("routes", finalMappings)

}

// ProcessToResourceData convert an app's process information to terraform state
func ProcessToResourceData(d *schema.ResourceData, proc resources.Process) {
	log.Printf("---------- [READ] proc info : %+v", proc)
	_ = d.Set("instances", proc.Instances.Value)
	_ = d.Set("memory", proc.MemoryInMB.Value)
	_ = d.Set("disk_quota", proc.DiskInMB.Value)
	_ = d.Set("health_check_type", proc.HealthCheckType)
	_ = d.Set("health_check_http_endpoint", proc.HealthCheckEndpoint)
	_ = d.Set("health_check_timeout", proc.HealthCheckTimeout)
}
