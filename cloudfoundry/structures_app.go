package cloudfoundry

import (
	"encoding/json"
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/appdeployers"
)

func ResourceDataToAppDeploy(d *schema.ResourceData) (appdeployers.AppDeploy, error) {
	enableSSH := d.Get("enable_ssh").(bool)
	if d.IsNewResource() {
		// if user does not explicitly set allow_ssh
		// it set allow ssh to true only during creation
		if _, ok := d.GetOk("enable_ssh"); !ok {
			enableSSH = true
		}
	}

	app := ccv2.Application{
		GUID:                    d.Id(),
		Name:                    d.Get("name").(string),
		Instances:               IntToNullInt(d.Get("instances").(int)),
		Memory:                  IntToNullByteSizeZero(d.Get("memory").(int)),
		DiskQuota:               IntToNullByteSizeZero(d.Get("disk_quota").(int)),
		StackGUID:               d.Get("stack").(string),
		Buildpack:               StringToFilteredString(d.Get("buildpack").(string)),
		Command:                 StringToFilteredString(d.Get("command").(string)),
		EnableSSH:               BoolToNullBool(enableSSH),
		State:                   constant.ApplicationStarted,
		DockerImage:             d.Get("docker_image").(string),
		HealthCheckHTTPEndpoint: d.Get("health_check_http_endpoint").(string),
		HealthCheckType:         constant.ApplicationHealthCheckType(d.Get("health_check_type").(string)),
		HealthCheckTimeout:      uint64(d.Get("health_check_timeout").(int)),
		SpaceGUID:               d.Get("space").(string),
	}
	if d.Get("stopped").(bool) {
		app.State = constant.ApplicationStopped
	}
	ports := make([]int, 0)
	for _, vv := range d.Get("ports").(*schema.Set).List() {
		ports = append(ports, vv.(int))
	}
	if len(ports) == 0 && app.DockerImage == "" {
		ports = []int{8080}
	}
	app.Ports = ports

	if v, ok := d.GetOk("docker_credentials"); ok {
		vv := v.(map[string]interface{})
		app.DockerCredentials = ccv2.DockerCredentials{
			Username: vv["username"].(string),
			Password: vv["password"].(string),
		}
	}
	if v, ok := d.GetOk("environment"); ok {
		vv := v.(map[string]interface{})
		envVars := make(map[string]string)
		for k, v := range vv {
			envVars[k] = fmt.Sprint(v)
		}
		app.EnvironmentVariables = envVars
	}

	mappings := make([]ccv2.RouteMapping, 0)
	for _, r := range getListOfStructs(d.Get("routes")) {
		mappings = append(mappings, ccv2.RouteMapping{
			RouteGUID: r["route"].(string),
			AppPort:   r["port"].(int),
		})
	}

	bindings := make([]ccv2.ServiceBinding, 0)
	for _, r := range getListOfStructs(d.Get("service_binding")) {
		params := r["params"].(map[string]interface{})
		paramJson := r["params_json"].(string)
		if len(params) == 0 && paramJson != "" {
			params = make(map[string]interface{})
			err := json.Unmarshal([]byte(paramJson), &params)
			if err != nil {
				return appdeployers.AppDeploy{}, err
			}
		}
		bindings = append(bindings, ccv2.ServiceBinding{
			ServiceInstanceGUID: r["service_instance"].(string),
			Parameters:          params,
		})
	}
	return appdeployers.AppDeploy{
		App:             app,
		ServiceBindings: bindings,
		Mappings:        mappings,
		Path:            d.Get("path").(string),
		StartTimeout:    time.Duration(d.Get("timeout").(int)) * time.Second,
		BindTimeout:     DefaultBindTimeout,
		StageTimeout:    DefaultStageTimeout,
	}, nil
}

func AppDeployToResourceData(d *schema.ResourceData, appDeploy appdeployers.AppDeployResponse) {
	d.SetId(appDeploy.App.GUID)
	d.Set("name", appDeploy.App.Name)
	d.Set("space", appDeploy.App.SpaceGUID)
	d.Set("ports", appDeploy.App.Ports)
	d.Set("instances", appDeploy.App.Instances.Value)
	d.Set("memory", appDeploy.App.Memory.Value)
	d.Set("disk_quota", appDeploy.App.DiskQuota.Value)
	d.Set("stack", appDeploy.App.StackGUID)
	d.Set("buildpack", appDeploy.App.Buildpack.Value)
	d.Set("command", appDeploy.App.Command.Value)
	d.Set("enable_ssh", appDeploy.App.EnableSSH.Value)
	d.Set("stopped", appDeploy.App.State == constant.ApplicationStopped)
	d.Set("docker_image", appDeploy.App.DockerImage)
	d.Set("health_check_http_endpoint", appDeploy.App.HealthCheckHTTPEndpoint)
	d.Set("health_check_type", string(appDeploy.App.HealthCheckType))
	d.Set("health_check_timeout", int(appDeploy.App.HealthCheckTimeout))
	d.Set("environment", appDeploy.App.EnvironmentVariables)

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
			if len(binding.Parameters) > 0 && len(curBinding["params"].(map[string]interface{})) > 0 {
				curBinding["params"] = binding.Parameters
			}
			if len(binding.Parameters) > 0 && (curBinding["params_json"].(string) != "" || len(curBinding["params"].(map[string]interface{})) == 0) {
				// error can't happen and skip it when sure there is no error is the way of life in go
				b, _ := json.Marshal(binding.Parameters)
				curBinding["params_json"] = string(b)
			}
			curBinding["service_instance"] = binding.ServiceInstanceGUID
			finalBindings = append(finalBindings, curBinding)
		}
	}
	d.Set("service_binding", finalBindings)

	mappingsTf := getListOfStructs(d.Get("routes"))
	finalMappings := make([]map[string]interface{}, 0)
	for _, mapping := range appDeploy.RouteMapping {
		// if 0 it mean app port has been set to null which means it takes the first port found in app port definition
		if mapping.AppPort <= 0 {
			mapping.AppPort = appDeploy.App.Ports[0]
		}
		if IsImportState(d) {
			finalMappings = append(finalMappings, map[string]interface{}{
				"route": mapping.RouteGUID,
				"port":  mapping.AppPort,
			})
			continue
		}
		curMappingsRaw, ok := getInSlice(mappingsTf, func(object interface{}) bool {
			objMap := object.(map[string]interface{})
			if objMap["port"].(int) <= 0 {
				return objMap["route"] == mapping.RouteGUID
			}
			return objMap["route"] == mapping.RouteGUID && objMap["port"] == mapping.AppPort
		})
		if !ok {
			continue
		}
		for _, curMappingRaw := range curMappingsRaw {
			curMapping := curMappingRaw.(map[string]interface{})
			curMapping["route"] = mapping.RouteGUID
			curMapping["port"] = mapping.AppPort
			finalMappings = append(finalMappings, curMapping)
		}

	}
	d.Set("routes", finalMappings)

}

type ResourceChanger interface {
	HasChange(key string) bool
}
