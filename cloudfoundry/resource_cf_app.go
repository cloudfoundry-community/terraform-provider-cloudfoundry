package cloudfoundry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/appdeployers"
)

// schema.BasicMapReader
// DefaultAppTimeout - Timeout (in seconds) when pushing apps to CF
const (
	DefaultAppTimeout   = 60
	DefaultBindTimeout  = 5 * time.Minute
	DefaultStageTimeout = 15 * time.Minute
	DefaultAppPort      = 8080
)

func resourceApp() *schema.Resource {
	return &schema.Resource{

		Create: resourceAppCreate,
		Read:   resourceAppRead,
		Update: resourceAppUpdate,
		Delete: resourceAppDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAppImport,
		},
		SchemaVersion: 3,
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
				Default:  DefaultAppTimeout,
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
				ValidateFunc: validateStrategy,
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
						port = DefaultAppPort
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
				ValidateFunc: validateAppHealthCheckType,
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

func validateAppHealthCheckType(v interface{}, k string) (ws []string, errs []error) {
	value := v.(string)
	if value != "port" && value != "process" && value != "http" && value != "none" {
		errs = append(errs, fmt.Errorf("%q must be one of 'port', 'process', 'http' or 'none'", k))
	}
	return ws, errs
}

func validateStrategy(v interface{}, k string) (ws []string, errs []error) {
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

func resourceAppCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	deployer := session.Deployer.Strategy(d.Get("strategy").(string))
	log.Printf("[INFO] Use deploy strategy %s", deployer.Names()[0])

	appDeploy, err := ResourceDataToAppDeploy(d)
	if err != nil {
		return err
	}

	appResp, err := deployer.Deploy(appDeploy)
	if err != nil {
		return err
	}
	AppDeployToResourceData(d, appResp)
	err = metadataCreate(appMetadata, d, meta)
	if err != nil {
		return err
	}
	return nil
}

func resourceAppRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	app, _, err := session.ClientV2.GetApplication(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	if idBg, ok := d.GetOk("id_bg"); !ok || idBg == "" {
		d.Set("id_bg", d.Id())
	}
	mappings, _, err := session.ClientV2.GetRouteMappings(ccv2.FilterEqual(constant.AppGUIDFilter, d.Id()))
	if err != nil {
		return err
	}
	bindings, _, err := session.ClientV2.GetServiceBindings(ccv2.FilterEqual(constant.AppGUIDFilter, d.Id()))
	if err != nil {
		return err
	}
	AppDeployToResourceData(d, appdeployers.AppDeployResponse{
		App:             app,
		RouteMapping:    mappings,
		ServiceBindings: bindings,
	})
	err = metadataRead(appMetadata, d, meta, false)
	if err != nil {
		return err
	}
	return nil
}

func resourceAppUpdate(d *schema.ResourceData, meta interface{}) error {
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
				return err
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
					return err
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
				return err
			}
			for _, binding := range bindings {
				_, _, err := session.ClientV2.DeleteServiceBinding(binding.GUID, true)
				if err != nil && !IsErrNotFound(err) {
					return err
				}
			}

		}
	}

	appDeploy, err := ResourceDataToAppDeploy(d)
	if err != nil {
		return err
	}

	// we are on the case where app code change so we can run directly deploy
	// which will do all mapping and binding and update the app
	if IsAppCodeChange(d) {
		appResp, err := deployer.Deploy(appDeploy)
		if err != nil {
			return err
		}
		d.Partial(false)
		AppDeployToResourceData(d, appResp)
		return err
	}

	if d.HasChange("routes") {
		mappings, err := session.RunBinder.MapRoutes(appDeploy)
		if err != nil {
			return err
		}
		appDeploy.Mappings = mappings
	}

	if d.HasChange("service_binding") {
		bindings, err := session.RunBinder.BindServiceInstances(appDeploy)
		if err != nil {
			return err
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
	}

	if IsAppUpdateOnly(d) || (IsAppRestageNeeded(d) && !deployer.IsCreateNewApp()) || (IsAppRestartNeeded(d) && !deployer.IsCreateNewApp()) {
		app, _, err := session.ClientV2.UpdateApplication(appUpdate)
		if err != nil {
			return err
		}
		appDeploy.App = app
	}

	if !appDeploy.IsDockerImage() && (IsAppRestageNeeded(d) || (deployer.IsCreateNewApp() && IsAppRestartNeeded(d))) {
		appResp, err := deployer.Restage(appDeploy)
		if err != nil {
			return err
		}
		d.Partial(false)
		AppDeployToResourceData(d, appResp)
		return nil
	}

	if IsAppRestartNeeded(d) {
		err := session.RunBinder.Restart(appDeploy, DefaultStageTimeout)
		if err != nil {
			return err
		}
		d.Partial(false)
		return nil
	}
	err = metadataUpdate(appMetadata, d, meta)
	if err != nil {
		return err
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

func resourceAppDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	_, err := session.ClientV2.DeleteApplication(d.Id())
	return err
}
