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
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/v3appdeployers"
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

		CreateContext: resourceAppCreate,
		ReadContext:   resourceAppRead,
		UpdateContext: resourceAppUpdate,
		DeleteContext: resourceAppDelete,

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
	if names, ok := v3appdeployers.ValidStrategy(value); !ok {
		errs = append(errs,
			fmt.Errorf("%q must be one of '%s' or 'none'", k, strings.Join(names, "', '")))
	}
	return ws, errs
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	deployer := session.V3Deployer.Strategy(d.Get("strategy").(string))
	// log.Printf("[INFO] Use deploy strategy %s", deployer.Names()[0])

	appDeploy, err := ResourceDataToAppDeployV3(d)
	if err != nil {
		return diag.FromErr(err)
	}

	appResp, err := deployer.Deploy(appDeploy)
	if err != nil {
		return diag.FromErr(err)
	}
	AppDeployV3ToResourceData(d, appResp)
	err = metadataCreate(appMetadata, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	query := ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: []string{d.Id()},
	}
	apps, _, err := session.ClientV3.GetApplications(query)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(apps) == 0 {
		d.SetId("")
		return nil
	}

	app := apps[0]

	if idBg, ok := d.GetOk("id_bg"); !ok || idBg == "" {
		_ = d.Set("id_bg", d.Id())
	}
	mappings, _, err := session.ClientV3.GetApplicationRoutes(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	bindings, _, err := session.ClientV3.GetServiceCredentialBindings(ccv3.Query{
		Key:    ccv3.AppGUIDFilter,
		Values: []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	bindings = ReorderBindings(bindings, d.Get("service_binding").([]interface{}))
	AppDeployV3ToResourceData(d, v3appdeployers.AppDeployResponse{
		App:             app,
		Mappings:        mappings,
		ServiceBindings: bindings,
		// Buildpacks apps have a default port set to 8080. Custom port for app process and route is not yet supported
		// Default port for docker apps is 9000, this case is not handled
		Ports: []int{DefaultAppPort},
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

	// Fetch process information
	proc, _, err := session.ClientV3.GetApplicationProcessByType(d.Id(), constant.ProcessTypeWeb)
	ProcessToResourceData(d, proc)

	err = metadataRead(appMetadata, d, meta, false)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.Partial(true)
	session := meta.(*managers.Session)
	defer func() {
		d.Set("id_bg", d.Id())
	}()
	deployer := session.V3Deployer.Strategy(d.Get("strategy").(string))

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

	// Route mappings change
	if d.HasChange("routes") {
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
			destinations, _, err := session.ClientV3.GetRouteDestinations(routeGUID)
			if err != nil {
				return diag.FromErr(err)
			}

			// Loop through the list of destinations to find the one with matching appGUID
			for _, destination := range destinations {
				if destination.App.GUID == d.Id() {
					_, err := session.ClientV3.UnmapRoute(routeGUID, destination.GUID)
					// If the destination is not found, we continue instead of raising the error
					if err != nil && !IsErrNotFound(err) {
						return diag.FromErr(err)
					}
				}
			}
		}
	}

	// Service bindings change
	if d.HasChange("service_binding") {
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
			bindings, _, err := session.ClientV3.GetServiceCredentialBindings(
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
				return diag.FromErr(err)
			}
			for _, binding := range bindings {
				// We don't wait for async job to complete, TODO: discuss on this
				_, _, err := session.ClientV3.DeleteServiceCredentialBinding(binding.GUID)
				// If the binding is not found, we continue instead of raising the error
				if err != nil && !IsErrNotFound(err) {
					return diag.FromErr(err)
				}
			}

		}
	}

	// Parse tfstate to appDeploy struct for deployer
	appDeploy, err := ResourceDataToAppDeployV3(d)
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
		AppDeployV3ToResourceData(d, appResp)
		return nil
	}

	// Redo route mappings
	if d.HasChange("routes") {
		mappings, err := session.V3RunBinder.MapRoutes(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		appDeploy.Mappings = mappings
	}

	if d.HasChange("service_binding") {
		bindings, err := session.V3RunBinder.BindServiceInstances(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		appDeploy.ServiceBindings = bindings
	}

	// appUpdate := resources.Application{
	// 	GUID: appDeploy.App.GUID,
	// }

	// packageUpdate := resources.Package{}
	// processUpdate := resources.Process{}
	// sshUpdate := resources.ApplicationFeature{}
	// envVarUpdate := make(resources.EnvironmentVariables)

	if d.HasChange("name") {
		appDeploy.App.Name = d.Get("name").(string)
	}
	if d.HasChange("ports") {
		ports := make([]int, 0)
		for _, vv := range d.Get("ports").(*schema.Set).List() {
			ports = append(ports, vv.(int))
		}
		appDeploy.Ports = ports
		log.Printf("[WARN] Ports have changed but not yet supported in v3 provider")
	}
	if d.HasChange("instances") {
		appDeploy.Process.Instances = IntToNullInt(d.Get("instances").(int))
	}
	if d.HasChange("memory") {
		appDeploy.Process.MemoryInMB = IntToNullUint64Zero(d.Get("memory").(int))
	}
	if d.HasChange("disk_quota") {
		appDeploy.Process.DiskInMB = IntToNullUint64Zero(d.Get("disk_quota").(int))
	}
	if d.HasChange("stack") {
		appDeploy.App.StackName = d.Get("stack").(string)
	}
	if d.HasChange("buildpack") {
		appDeploy.App.LifecycleBuildpacks = []string{d.Get("buildpack").(string)}
	}
	if d.HasChange("command") {
		appDeploy.Process.Command = StringToFilteredString(d.Get("command").(string))
	}
	if d.HasChange("enable_ssh") {
		appDeploy.EnableSSH.Enabled = d.Get("enable_ssh").(bool)
	}
	if d.HasChange("stopped") {
		state := constant.ApplicationStarted
		if d.Get("stopped").(bool) {
			state = constant.ApplicationStopped
		}
		appDeploy.App.State = state
	}
	if d.HasChange("docker_image") {
		appDeploy.AppPackage.DockerImage = d.Get("docker_image").(string)
		if v, ok := d.GetOk("docker_credentials"); ok {
			vv := v.(map[string]interface{})
			appDeploy.AppPackage.DockerUsername = vv["username"].(string)
			appDeploy.AppPackage.DockerPassword = vv["password"].(string)
		}
	}
	if d.HasChange("health_check_http_endpoint") {
		appDeploy.Process.HealthCheckEndpoint = d.Get("health_check_http_endpoint").(string)
	}
	if d.HasChange("health_check_type") {
		appDeploy.Process.HealthCheckType = constant.HealthCheckType(d.Get("health_check_type").(string))
	}
	if d.HasChange("health_check_timeout") {
		appDeploy.Process.HealthCheckTimeout = int64(d.Get("health_check_timeout").(int))
	}
	if d.HasChange("environment") {
		if v, ok := d.GetOk("environment"); ok {
			vv := v.(map[string]interface{})
			for k, v := range vv {
				appDeploy.EnvVars[k] = *types.NewFilteredString(fmt.Sprint(v))
			}
		}

		// Remove stale / externally set variables
		// Already implemented in v3 so we don't touch
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

	// appDeploy.AppPackage = packageUpdate
	// appDeploy.Process = processUpdate
	// appDeploy.EnableSSH = sshUpdate
	// appDeploy.EnvVars = envVarUpdate

	if IsAppUpdateOnly(d) || (IsAppRestageNeeded(d) && !deployer.IsCreateNewApp()) || (IsAppRestartNeeded(d) && !deployer.IsCreateNewApp()) {
		app, _, err := session.ClientV3.UpdateApplication(appDeploy.App)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("instances") {
			// Scale only web process type
			procScale := resources.Process{
				Type:      constant.ProcessTypeWeb,
				Instances: IntToNullInt(d.Get("instances").(int)),
			}
			// log.Printf("scale proc : %+v", procScale)
			_, _, err := session.ClientV3.CreateApplicationProcessScale(d.Id(), procScale)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("stopped") {
			var f func(appGUID string) (resources.Application, ccv3.Warnings, error)
			if d.Get("stopped").(bool) {
				f = session.ClientV3.UpdateApplicationStart
			} else {
				f = session.ClientV3.UpdateApplicationStop
			}
			app, _, err = f(d.Id())
			if err != nil {
				return diag.FromErr(err)
			}
		}

		appDeploy.App = app
	}

	if IsAppRestageNeeded(d) || (deployer.IsCreateNewApp() && IsAppRestartNeeded(d)) {
		appResp, err := deployer.Restage(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Partial(false)
		AppDeployV3ToResourceData(d, appResp)
		return nil
	}

	if IsAppRestartNeeded(d) {
		err := session.V3RunBinder.Restart(appDeploy, DefaultStageTimeout)
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

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	appGUID := d.Id()
	_, _, err := session.ClientV3.UpdateApplicationStop(appGUID)
	if err != nil {
		return diag.FromErr(err)
	}

	jobURL, _, err := session.ClientV3.DeleteApplication(appGUID)

	err = PollAsyncJob(PollingConfig{
		session: session,
		jobURL:  jobURL,
	})
	return diag.FromErr(err)
}
