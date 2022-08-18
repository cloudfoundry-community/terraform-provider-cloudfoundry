package cloudfoundry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/appdeployers"
)

// schema.BasicMapReader
// DefaultAppTimeout - Timeout (in seconds) when pushing apps to CF
const (
	DefaultAppV3Timeout   = 60
	DefaultBindV3Timeout  = 5 * time.Minute
	DefaultStageV3Timeout = 15 * time.Minute
	DefaultAppV3Port      = 8080
)

func resourceAppV3() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceAppV3Create,
		ReadContext:   resourceAppV3Read,
		UpdateContext: resourceAppV3Update,
		DeleteContext: resourceAppV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAppImport,
		},
		SchemaVersion: 4,
		MigrateState:  resourceAppMigrateState,
		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ports": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Set:      resourceIntegerSet,
			},
			"instances": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"disk_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"stack": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"buildpack": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"command": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"enable_ssh": &schema.Schema{
				Type:       schema.TypeBool,
				ConfigMode: schema.SchemaConfigModeAttr,
				Optional:   true,
				Computed:   true,
			},
			"timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  DefaultAppV3Timeout,
			},
			"stopped": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"strategy": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				Description:  "Deployment strategy, default to none but accept blue-green strategy",
				ValidateFunc: validateV3Strategy,
			},
			"path": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Path to an app zip in the form of unix path or http url",
				ConflictsWith: []string{"docker_image", "docker_credentials"},
			},
			"source_code_hash": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"docker_image": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"path"},
			},
			"docker_credentials": &schema.Schema{
				Type:          schema.TypeMap,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"path"},
			},
			"service_binding": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_instance": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"params": &schema.Schema{
							Type:      schema.TypeMap,
							Optional:  true,
							Sensitive: true,
						},
						"params_json": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsJSON,
						},
					},
				},
			},
			"routes": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Set: func(v interface{}) int {
					elem := v.(map[string]interface{})
					port := elem["port"].(int)
					if port == 0 {
						port = DefaultAppV3Port
					}
					return hashcode.String(fmt.Sprintf(
						"%s-%d",
						elem["route"],
						port,
					))
				},
				Elem: &schema.Resource{
					CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {

						if diff.HasChange("port") {
							return nil
						}
						oldPort, newPort := diff.GetChange("port")
						if oldPort != "" && newPort == "" {
							return diff.SetNew("port", oldPort)
						}
						return nil
					},
					Schema: map[string]*schema.Schema{
						"route": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"port": &schema.Schema{
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
					},
				},
			},
			"environment": &schema.Schema{
				Type:      schema.TypeMap,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"health_check_http_endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"health_check_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "port",
				ValidateFunc: validateAppV3HealthCheckType,
			},
			"health_check_timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"id_bg": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			if diff.HasChange("docker_image") || diff.HasChange("path") {
				oldImg, newImg := diff.GetChange("docker_image")
				oldPath, newPath := diff.GetChange("path")
				if oldImg == "" && newImg != "" && newPath == "" {
					return diff.ForceNew("docker_image")
				}
				if oldPath == "" && newPath != "" && newImg == "" {
					return diff.ForceNew("path")
				}
			}
			if diff.Id() == "" {
				return nil
			}
			session := meta.(*managers.Session)
			deployer := session.Deployer.Strategy(diff.Get("strategy").(string))
			if IsAppRestageNeeded(diff) ||
				(deployer.IsCreateNewApp() && IsAppRestartNeeded(diff)) ||
				(deployer.IsCreateNewApp() && IsAppCodeChange(diff)) {
				diff.SetNewComputed("id_bg")
			}

			return nil
		},
	}
}

func validateAppV3HealthCheckType(v interface{}, k string) (ws []string, errs []error) {
	value := v.(string)
	if value != "port" && value != "process" && value != "http" && value != "none" {
		errs = append(errs, fmt.Errorf("%q must be one of 'port', 'process', 'http' or 'none'", k))
	}
	return ws, errs
}

func validateV3Strategy(v interface{}, k string) (ws []string, errs []error) {
	value := strings.ToLower(v.(string))
	if value == "none" {
		return ws, errs
	}
	if names, ok := appdeployers.ValidStrategy(value); !ok {
		errs = append(errs,
			fmt.Errorf("%q must be one of '%s' or 'none'", k, strings.Join(names, "', '")))
	}
	return ws, errs
}

func resourceAppV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	deployer := session.Deployer.Strategy(d.Get("strategy").(string))
	log.Printf("[INFO] Use deploy strategy %s", deployer.Names()[0])

	appDeploy, err := ResourceDataToAppDeploy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	appResp, err := deployer.Deploy(appDeploy)
	if err != nil {
		return diag.FromErr(err)
	}
	AppDeployToResourceData(d, appResp)
	err = metadataCreate(appMetadata, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAppV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)


	query := ccv3.Query{
		Key:    ccv3.NameFilter,
		Values: []string{d.Id()},
	}
	apps, _, err := session.ClientV3.GetApplications(query)
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if idBg, ok := d.GetOk("id_bg"); !ok || idBg == "" {
		_ = d.Set("id_bg", d.Id())
	}
	mappings, _, err := session.ClientV2.GetRouteMappings(ccv3.FilterEqual(constant.AppGUIDFilter, d.Id()))
	if err != nil {
		return diag.FromErr(err)
	}
	bindings, _, err := session.ClientV3.GetServiceBindings(ccv3.FilterEqual(constant.AppGUIDFilter, d.Id()))
	if err != nil {
		return diag.FromErr(err)
	}
	bindings = reorderBindings(bindings, d.Get("service_binding").([]interface{}))
	AppDeployToResourceData(d, appdeployers.AppDeployResponse{
		App:             app,
		RouteMapping:    mappings,
		ServiceBindings: bindings,
	})
	// droplet sync through V3 API
	droplet, _, err := session.ClientV3.GetApplicationDropletCurrent(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	DropletToResourceData(d, droplet)

	err = metadataRead(appMetadata, d, meta, false)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func getServiceBindingFromList(guid string, bindings []ccv2.ServiceBinding) (ccv2.ServiceBinding, bool) {
	for _, binding := range bindings {
		if binding.ServiceInstanceGUID == guid {
			return binding, true
		}
	}
	return ccv2.ServiceBinding{}, false
}

func reorderBindings(bindings []ccv2.ServiceBinding, currentBindings []interface{}) []ccv2.ServiceBinding {
	finalBindings := make([]ccv2.ServiceBinding, 0)
	for _, currentBindings := range currentBindings {
		if currentBindings == nil {
			continue
		}
		item := currentBindings.(map[string]interface{})
		if binding, ok := getServiceBindingFromList(item["service_instance"].(string), bindings); ok {
			finalBindings = append(finalBindings, binding)
		}
	}
	for _, binding := range bindings {
		if _, ok := getServiceBindingFromList(binding.ServiceInstanceGUID, finalBindings); ok {
			continue
		}
		finalBindings = append(finalBindings, binding)
	}
	return finalBindings
}

func resourceAppV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.Partial(true)
	session := meta.(*managers.Session)
	defer func() {
		d.Set("id_bg", d.Id())
	}()
	deployer := session.Deployer.Strategy(d.Get("strategy").(string))

	// sanitize any empty port under 1024
	// this means that we are using not predefined port by user
	// push back to empty list to make blue-green happy with api
	ports := d.Get("ports").(*schema.Set).List()
	finalPorts := make([]int, 0)
	for _, port := range ports {
		if port.(int) <= 1024 {
			continue
		}
		finalPorts = append(finalPorts, port.(int))
	}
	d.Set("ports", finalPorts)

	if d.HasChange("routes") {
		oldRoutes, newRoutes := d.GetChange("routes")
		remove, _ := getListMapChanges(oldRoutes, newRoutes, func(source, item map[string]interface{}) bool {
			return source["route"] == item["route"] && source["port"] == item["port"]
		})
		for _, r := range remove {
			mappings, _, err := session.ClientV2.GetRouteMappings(filterAppGuid(d.Id()), filterRouteGuid(r["route"].(string)))
			if err != nil {
				return diag.FromErr(err)
			}
			for _, mapping := range mappings {
				// if 0 it mean app port has been set to null which means it takes the first port found in app port definition
				if mapping.AppPort <= 0 {
					mapping.AppPort = (d.Get("ports").(*schema.Set).List()[0]).(int)
				}
				if mapping.AppPort != r["port"] {
					continue
				}
				_, err := session.ClientV2.DeleteRouteMapping(mapping.GUID)
				if err != nil && !IsErrNotFound(err) {
					return diag.FromErr(err)
				}
			}
		}
	}

	if d.HasChange("service_binding") {
		oldBindings, newBindings := d.GetChange("service_binding")
		remove, _ := getListMapChanges(oldBindings, newBindings, func(source, item map[string]interface{}) bool {
			matchId := source["service_instance"] == item["service_instance"]
			if !matchId {
				return false
			}
			isDiff, err := isDiffAppParamsBinding(source, item)
			if err != nil {
				panic(err)
			}
			return !isDiff
		})

		for _, r := range remove {
			bindings, _, err := session.ClientV2.GetServiceBindings(filterAppGuid(d.Id()), filterServiceInstanceGuid(r["service_instance"].(string)))
			if err != nil {
				return diag.FromErr(err)
			}
			for _, binding := range bindings {
				_, _, err := session.ClientV2.DeleteServiceBinding(binding.GUID, true)
				if err != nil && !IsErrNotFound(err) {
					return diag.FromErr(err)
				}
			}

		}
	}

	appDeploy, err := ResourceDataToAppDeploy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// we are on the case where app code change so we can run directly deploy
	// which will do all mapping and binding and update the app
	if IsAppCodeChange(d) {
		appResp, err := deployer.Deploy(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Partial(false)
		AppDeployToResourceData(d, appResp)
		return nil
	}

	if d.HasChange("routes") {
		mappings, err := session.RunBinder.MapRoutes(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		appDeploy.Mappings = mappings
	}

	if d.HasChange("service_binding") {
		bindings, err := session.RunBinder.BindServiceInstances(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		appDeploy.ServiceBindings = bindings
	}

	appUpdate := ccv2.Application{
		GUID: appDeploy.App.GUID,
	}

	if d.HasChange("name") {
		appUpdate.Name = d.Get("name").(string)
	}
	if d.HasChange("ports") {
		ports := make([]int, 0)
		for _, vv := range d.Get("ports").(*schema.Set).List() {
			ports = append(ports, vv.(int))
		}
		appUpdate.Ports = ports
	}
	if d.HasChange("instances") {
		appUpdate.Instances = IntToNullInt(d.Get("instances").(int))
	}
	if d.HasChange("memory") {
		appUpdate.Memory = IntToNullByteSizeZero(d.Get("memory").(int))
	}
	if d.HasChange("disk_quota") {
		appUpdate.DiskQuota = IntToNullByteSizeZero(d.Get("disk_quota").(int))
	}
	if d.HasChange("stack") {
		appUpdate.StackGUID = d.Get("stack").(string)
	}
	if d.HasChange("buildpack") {
		appUpdate.Buildpack = StringToFilteredString(d.Get("buildpack").(string))
	}
	if d.HasChange("command") {
		appUpdate.Command = StringToFilteredString(d.Get("command").(string))
	}
	if d.HasChange("enable_ssh") {
		appUpdate.EnableSSH = BoolToNullBool(d.Get("enable_ssh").(bool))
	}
	if d.HasChange("stopped") {
		state := constant.ApplicationStarted
		if d.Get("stopped").(bool) {
			state = constant.ApplicationStopped
		}
		appUpdate.State = state
	}
	if d.HasChange("docker_image") {
		appUpdate.DockerImage = d.Get("docker_image").(string)
		if v, ok := d.GetOk("docker_credentials"); ok {
			vv := v.(map[string]interface{})
			appUpdate.DockerCredentials = ccv2.DockerCredentials{
				Username: vv["username"].(string),
				Password: vv["password"].(string),
			}
		}
	}
	if d.HasChange("health_check_http_endpoint") {
		appUpdate.HealthCheckHTTPEndpoint = d.Get("health_check_http_endpoint").(string)
	}
	if d.HasChange("health_check_type") {
		appUpdate.HealthCheckType = constant.ApplicationHealthCheckType(d.Get("health_check_type").(string))
	}
	if d.HasChange("health_check_timeout") {
		appUpdate.HealthCheckTimeout = uint64(d.Get("health_check_timeout").(int))
	}
	if d.HasChange("environment") {
		if v, ok := d.GetOk("environment"); ok {
			vv := v.(map[string]interface{})
			envVars := make(map[string]string)
			for k, v := range vv {
				envVars[k] = fmt.Sprint(v)
			}
			appUpdate.EnvironmentVariables = envVars
		}
		// Remove stale / externally set variables
		if currentEnv, _, err := session.ClientV3.GetApplicationEnvironment(appDeploy.App.GUID); err == nil {
			var staleVars []string
			var vv map[string]interface{}
			if v, ok := d.GetOk("environment"); ok {
				vv = v.(map[string]interface{})
			}
			for s := range currentEnv.EnvironmentVariables {
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
				_ = session.BitsManager.SetAppEnvironmentVariables(appDeploy.App.GUID, env)
			}
		}
	}

	if IsAppUpdateOnly(d) || (IsAppRestageNeeded(d) && !deployer.IsCreateNewApp()) || (IsAppRestartNeeded(d) && !deployer.IsCreateNewApp()) {
		app, _, err := session.ClientV2.UpdateApplication(appUpdate)
		if err != nil {
			return diag.FromErr(err)
		}
		appDeploy.App = app
	}

	if IsAppRestageNeeded(d) || (deployer.IsCreateNewApp() && IsAppRestartNeeded(d)) {
		appResp, err := deployer.Restage(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Partial(false)
		AppDeployToResourceData(d, appResp)
		return nil
	}

	if IsAppRestartNeeded(d) {
		err := session.RunBinder.Restart(appDeploy, DefaultStageV3Timeout)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Partial(false)
		return nil
	}
	err = metadataUpdate(appMetadata, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)
	return nil
}

func IsAppCodeChange(d ResourceChanger) bool {
	return d.HasChange("path") || d.HasChange("source_code_hash")
}

func IsAppUpdateOnly(d ResourceChanger) bool {
	if IsAppCodeChange(d) || IsAppRestageNeeded(d) || IsAppRestartNeeded(d) {
		return false
	}
	return d.HasChange("name") || d.HasChange("instances") ||
		d.HasChange("enable_ssh") || d.HasChange("stopped")
}

func IsAppRestageNeeded(d ResourceChanger) bool {
	return d.HasChange("buildpack") || d.HasChange("stack") ||
		d.HasChange("service_binding") || d.HasChange("environment")
}

func IsAppRestartNeeded(d ResourceChanger) bool {
	return d.HasChange("memory") || d.HasChange("disk_quota") ||
		d.HasChange("command") || d.HasChange("health_check_http_endpoint") ||
		d.HasChange("docker_image") || d.HasChange("health_check_type") ||
		d.HasChange("environment")
}

func isDiffAppParamsBinding(oldBinding, currentBinding map[string]interface{}) (bool, error) {
	if len(oldBinding["params"].(map[string]interface{})) != len(currentBinding["params"].(map[string]interface{})) {
		return true, nil
	}
	if len(oldBinding["params"].(map[string]interface{})) > 0 {
		oldParams := oldBinding["params"].(map[string]interface{})
		currentParams := currentBinding["params"].(map[string]interface{})
		return reflect.DeepEqual(oldParams, currentParams), nil
	}
	oldJson := oldBinding["params_json"].(string)
	currentJson := oldBinding["params_json"].(string)
	if oldJson == "" && currentJson == "" {
		return false, nil
	}
	if oldJson == "" && currentJson != "" || oldJson != "" && currentJson == "" {
		return true, nil
	}
	var oldParams map[string]interface{}
	var currentParams map[string]interface{}
	err := json.Unmarshal([]byte(oldJson), &oldParams)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal([]byte(currentJson), &currentParams)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(oldParams, currentParams), nil
}

func resourceAppV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	_, err := session.ClientV2.DeleteApplication(d.Id())
	return diag.FromErr(err)
}
