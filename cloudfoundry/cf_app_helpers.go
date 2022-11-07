package cloudfoundry

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

// GetServiceBindingFromList returns the service credential binding and whether it's in the list
func GetServiceBindingFromList(guid string, bindings []resources.ServiceCredentialBinding) (resources.ServiceCredentialBinding, bool) {
	for _, binding := range bindings {
		if binding.ServiceInstanceGUID == guid {
			return binding, true
		}
	}
	return resources.ServiceCredentialBinding{}, false
}

// ReorderBindings -
func ReorderBindings(bindings []resources.ServiceCredentialBinding, currentBindings []interface{}) []resources.ServiceCredentialBinding {
	finalBindings := make([]resources.ServiceCredentialBinding, 0)
	for _, currentBinding := range currentBindings {
		if currentBinding == nil {
			continue
		}
		item := currentBinding.(map[string]interface{})
		if binding, ok := GetServiceBindingFromList(item["service_instance"].(string), bindings); ok {
			finalBindings = append(finalBindings, binding)
		}
	}
	for _, binding := range bindings {
		if _, ok := GetServiceBindingFromList(binding.ServiceInstanceGUID, finalBindings); ok {
			continue
		}
		finalBindings = append(finalBindings, binding)
	}
	return finalBindings
}

// RemoveStaleEnviromentVariables :
// Remove stale/externally set environment variables
func RemoveStaleEnviromentVariables(d *schema.ResourceData, session *managers.Session, appGUID string, currentEnv map[string]interface{}) error {
	var staleVars []string
	var vv map[string]interface{}
	if v, ok := d.GetOk("environment"); ok {
		vv = v.(map[string]interface{})
	}
	for s := range currentEnv {
		found := false
		for k := range vv {
			if k == s {
				found = true
				break
			}
		}
		if !found {
			staleVars = append(staleVars, s)
		}
	}
	if len(staleVars) > 0 {
		env := make(map[string]interface{})
		for _, e := range staleVars {
			env[e] = nil
		}
		_, _, err := session.BitsManager.UpdateAppEnvironment(appGUID, env)
		return err
	}
	return nil
}

// UnmapOldRoutes :
// unmap old routes / externally mapped routes
func UnmapOldRoutes(d *schema.ResourceData, client *ccv3.Client) error {
	oldRoutes, newRoutes := d.GetChange("routes")

	// getListMapChanges returns the routes to remove and the new routes to be added
	// new routes are handled later so we only remove deleted routes here
	remove, _ := getListMapChanges(oldRoutes, newRoutes, func(source, item map[string]interface{}) bool {
		return source["route"] == item["route"] && source["port"] == item["port"]
	})

	for _, r := range remove {
		// r contains route_id and port but port is always 0 as we don't support appPort in v3
		// Get list of destinations for each route to remove and delete the destination with matching app GUID
		routeGUID := r["route"].(string)
		destinations, _, err := client.GetRouteDestinations(routeGUID)
		if err != nil {
			return err
		}

		// Loop through the list of destinations to find the one with matching appGUID
		for _, destination := range destinations {
			if destination.App.GUID == d.Id() {
				_, err := client.UnmapRoute(routeGUID, destination.GUID)
				// If the destination is not found, we continue instead of raising the error
				if err != nil && !IsErrNotFound(err) {
					return err
				}
			}
		}
	}
	return nil
}

// UnbindServiceInstances :
// Unbind service instances that weren't set
func UnbindServiceInstances(d *schema.ResourceData, client *ccv3.Client) error {
	oldBindings, newBindings := d.GetChange("service_binding")

	// getListMapChanges returns the service credential bindings to remove and new ones to be added
	// new service credential bindings are handled later so we only remove deleted ones here
	remove, _ := getListMapChanges(oldBindings, newBindings, func(source, item map[string]interface{}) bool {
		matchID := source["service_instance"] == item["service_instance"]
		if !matchID {
			return false
		}
		// if binding parameters are different, delete the binding
		isDiff, err := isDiffAppParamsBinding(source, item)
		if err != nil {
			panic(err)
		}
		return !isDiff
	})

	for _, r := range remove {
		// r contains service_instance_id and params as map[string]interface{} or params_json as string
		// We simply get all service credential bindings between this app and the service instance then delete
		bindings, _, err := client.GetServiceCredentialBindings(
			ccv3.Query{
				Key:    ccv3.AppGUIDFilter,
				Values: []string{d.Id()},
			},
			ccv3.Query{
				Key:    ccv3.QueryKey("service_instance_guids"),
				Values: []string{r["service_instance"].(string)},
			},
		)
		if err != nil {
			return err
		}
		for _, binding := range bindings {
			// Delete service binding
			jobURL, _, err := client.DeleteServiceCredentialBinding(binding.GUID)
			if err != nil {
				return err
			}

			// Poll for service binding deletion
			err = common.PollingWithTimeout(func() (bool, error) {
				job, _, err := client.GetJob(jobURL)
				if err != nil {
					return true, err
				}

				// Stop polling and return error if job failed
				if job.State == constant.JobFailed {
					return true, fmt.Errorf(
						"Delete service binding failed, reason: %+v",
						job.Errors(),
					)
				}

				if job.State == constant.JobComplete {
					return true, nil
				}

				// Last operation initial or inprogress or job not completed, continue polling
				return false, nil
			}, 5*time.Second, 2*time.Minute)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

// func IsAppCodeChange(d ResourceChanger) bool {
// 	return d.HasChange("path") || d.HasChange("source_code_hash")
// }

// func IsAppUpdateOnly(d ResourceChanger) bool {
// 	if IsAppCodeChange(d) || IsAppRestageNeeded(d) || IsAppRestartNeeded(d) {
// 		return false
// 	}
// 	return d.HasChange("name") || d.HasChange("instances") ||
// 		d.HasChange("enable_ssh") || d.HasChange("stopped")
// }

// func IsAppRestageNeeded(d ResourceChanger) bool {
// 	return d.HasChange("buildpack") || d.HasChange("stack") ||
// 		d.HasChange("service_binding") || d.HasChange("environment")
// }

// func IsAppRestartNeeded(d ResourceChanger) bool {
// 	return d.HasChange("memory") || d.HasChange("disk_quota") ||
// 		d.HasChange("command") || d.HasChange("health_check_http_endpoint") ||
// 		d.HasChange("docker_image") || d.HasChange("health_check_type") ||
// 		d.HasChange("environment")
// }

// func isDiffAppParamsBinding(oldBinding, currentBinding map[string]interface{}) (bool, error) {
// 	if len(oldBinding["params"].(map[string]interface{})) != len(currentBinding["params"].(map[string]interface{})) {
// 		return true, nil
// 	}
// 	if len(oldBinding["params"].(map[string]interface{})) > 0 {
// 		oldParams := oldBinding["params"].(map[string]interface{})
// 		currentParams := currentBinding["params"].(map[string]interface{})
// 		return reflect.DeepEqual(oldParams, currentParams), nil
// 	}
// 	oldJson := oldBinding["params_json"].(string)
// 	currentJson := oldBinding["params_json"].(string)
// 	if oldJson == "" && currentJson == "" {
// 		return false, nil
// 	}
// 	if oldJson == "" && currentJson != "" || oldJson != "" && currentJson == "" {
// 		return true, nil
// 	}
// 	var oldParams map[string]interface{}
// 	var currentParams map[string]interface{}
// 	err := json.Unmarshal([]byte(oldJson), &oldParams)
// 	if err != nil {
// 		return false, err
// 	}
// 	err = json.Unmarshal([]byte(currentJson), &currentParams)
// 	if err != nil {
// 		return false, err
// 	}
// 	return reflect.DeepEqual(oldParams, currentParams), nil
// }
