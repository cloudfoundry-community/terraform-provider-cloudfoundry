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
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          1,
				DiffSuppressFunc: diffSuppressOnStoppedApps,
			},
			"memory": &schema.Schema{
				Type:             schema.TypeInt,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: diffSuppressOnStoppedApps,
			},
			"disk_quota": &schema.Schema{
				Type:             schema.TypeInt,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: diffSuppressOnStoppedApps,
			},
			"stack": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"buildpack": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"buildpacks"},
			},
			"buildpacks": &schema.Schema{
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"buildpack"},
			},
			"command": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
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
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{"path"},
				DiffSuppressFunc: diffSuppressOnStoppedApps,
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: diffSuppressOnStoppedApps,
			},
			"health_check_type": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "port",
				ValidateFunc:     validateAppV3HealthCheckType,
				DiffSuppressFunc: diffSuppressOnStoppedApps,
			},
			"health_check_timeout": &schema.Schema{
				Type:             schema.TypeInt,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: diffSuppressOnStoppedApps,
			},
			"health_check_invocation_timeout": &schema.Schema{
				Type:             schema.TypeInt,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: diffSuppressOnStoppedApps,
			},
			"id_bg": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			session := meta.(*managers.Session)
			deployer := session.Deployer.Strategy(diff.Get("strategy").(string))

			if (diff.HasChange("docker_image") || diff.HasChange("path")) && !deployer.IsCreateNewApp() {
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

			if IsAppRestageNeeded(diff) ||
				(deployer.IsCreateNewApp() && IsAppRestartNeeded(diff)) ||
				(deployer.IsCreateNewApp() && IsAppCodeChange(diff)) {
				if stopped, ok := diff.GetOk("stopped"); ok {
					if !stopped.(bool) {
						diff.SetNewComputed("id_bg")
					}
				}
			}

			if diff.HasChange("stack") && !deployer.IsCreateNewApp() {
				// Not b/g
				diff.ForceNew("stack")
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

// suppress diff on process/droplet related attributes
func diffSuppressOnStoppedApps(k, old, new string, d *schema.ResourceData) bool {
	if stopped, ok := d.GetOk("stopped"); ok {
		if stopped.(bool) {
			return true
		}
	}
	return false
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	deployer := session.V3Deployer.Strategy(d.Get("strategy").(string))
	log.Printf("[INFO] Use deploy strategy %s", deployer.Names()[0])

	appDeploy, err := ResourceDataToAppDeployV3(d)
	if err != nil {
		return diag.FromErr(err)
	}

	appResp, err := deployer.Deploy(appDeploy)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ports are set to 8080 by default
	appResp.Ports = appDeploy.Ports

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

	// Get enabled_ssh
	enableSSH, _, err := session.ClientV3.GetAppFeature(d.Id(), "ssh")
	if err != nil {
		return diag.FromErr(err)
	}

	// Get environment variables
	env, err := session.BitsManager.GetAppEnvironmentVariables(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Get route mapping
	mappings, _, err := session.ClientV3.GetApplicationRoutes(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Get service bindings
	bindings, _, err := session.ClientV3.GetServiceCredentialBindings(ccv3.Query{
		Key:    ccv3.AppGUIDFilter,
		Values: []string{d.Id()},
	})
	if err != nil {
		return diag.FromErr(err)
	}
	bindings = ReorderBindings(bindings, d.Get("service_binding").([]interface{}))

	// Fetch process information
	proc, _, err := session.ClientV3.GetApplicationProcessByType(d.Id(), constant.ProcessTypeWeb)
	if err != nil {
		return diag.FromErr(err)
	}
	// ProcessToResourceData(d, proc)

	// droplet sync through V3 API
	// Do nothing if droplet is not found
	droplet, _, err := session.ClientV3.GetApplicationDropletCurrent(d.Id())
	if err != nil && !strings.Contains(err.Error(), "Droplet not found") {
		return diag.FromErr(err)
	}

	appDeployResponse := v3appdeployers.AppDeployResponse{
		App:             app,
		Mappings:        mappings,
		ServiceBindings: bindings,
		EnableSSH:       v3appdeployers.AppFeatureToNullBool(enableSSH),
		EnvVars:         env,
		Process:         proc,
		// Set docker image
		AppPackage: resources.Package{
			DockerImage: droplet.Image,
		},
		// Buildpacks apps have a default port set to 8080. Custom port for app process and route is not yet supported
		// Default port for docker apps is 9000, this case is not handled
	}

	if appDeployResponse.AppPackage.DockerImage == "" {
		appDeployResponse.Ports = []int{DefaultAppPort}
	}

	AppDeployV3ToResourceData(d, appDeployResponse)

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

	// Parse tfstate to appDeploy struct for deployer
	appDeploy, err := ResourceDataToAppDeployV3(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// we are on the case where app code change so we can run directly deploy
	// which will do all mapping and binding and update the app
	// If the application uses b/g deployment method (IsCreateNewApp), and the app has changes that require restart/restage, simply use b/g deployment instead of simple updating the application.
	if IsAppCodeChange(d) || (deployer.IsCreateNewApp() && IsAppRestartNeeded(d)) || (deployer.IsCreateNewApp() && IsAppRestageNeeded(d)) {
		appResp, err := deployer.Deploy(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Partial(false)
		AppDeployV3ToResourceData(d, appResp)
		return nil
	}

	// Route mappings change
	if d.HasChange("routes") {
		err := UnmapOldRoutes(d, session.ClientV3)
		if err != nil {
			return diag.FromErr(err)
		}

		// Redo route mappings
		mappings, err := session.V3RunBinder.MapRoutes(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		appDeploy.Mappings = mappings
	}

	// Update application
	appUpdate := resources.Application{
		GUID: appDeploy.App.GUID,
	}
	if d.HasChange("name") {
		appUpdate.Name = d.Get("name").(string)
	}
	if d.HasChange("buildpack") {
		appUpdate.LifecycleBuildpacks = []string{d.Get("buildpack").(string)}
	}
	if d.HasChange("buildpacks") {
		buildpacks := make([]string, 0)
		for _, bpkg := range d.Get("buildpacks").([]interface{}) {
			buildpacks = append(buildpacks, bpkg.(string))
		}
		appUpdate.LifecycleBuildpacks = buildpacks
	}
	if d.HasChange("stack") {
		appUpdate.StackName = d.Get("stack").(string)
	}

	// Ports - custom ports are not yet supported in v3 provider
	if d.HasChange("ports") {
		ports := make([]int, 0)
		for _, vv := range d.Get("ports").(*schema.Set).List() {
			ports = append(ports, vv.(int))
		}
		log.Printf("[WARN] Ports have changed but not yet supported in v3 provider: %+v", ports)
	}

	// Process scale
	processScaleRequired := d.HasChange("instances") || d.HasChange("memory") || d.HasChange("disk_quota")
	processScale := resources.Process{
		Type: constant.ProcessTypeWeb,
	}
	if d.HasChange("instances") {
		processScale.Instances = IntToNullInt(d.Get("instances").(int))
	}
	if d.HasChange("memory") {
		processScale.MemoryInMB = IntToNullUint64Zero(d.Get("memory").(int))
	}
	if d.HasChange("disk_quota") {
		processScale.DiskInMB = IntToNullUint64Zero(d.Get("disk_quota").(int))
	}

	// Process update
	processUpdateRequired := d.HasChange("command") || d.HasChange("health_check_http_endpoint") || d.HasChange("health_check_type") || d.HasChange("health_check_timeout")
	processUpdate := resources.Process{
		Type: constant.ProcessTypeWeb,
	}
	if d.HasChange("command") {
		processUpdate.Command = StringToFilteredString(d.Get("command").(string))
	}
	if d.HasChange("health_check_http_endpoint") {
		processUpdate.HealthCheckEndpoint = d.Get("health_check_http_endpoint").(string)
	}
	if d.HasChange("health_check_type") {
		processUpdate.HealthCheckType = constant.HealthCheckType(d.Get("health_check_type").(string))
	}
	if d.HasChange("health_check_timeout") {
		processUpdate.HealthCheckTimeout = int64(d.Get("health_check_timeout").(int))
	}

	if d.HasChange("docker_image") {
		appDeploy.AppPackage.DockerImage = d.Get("docker_image").(string)
		if v, ok := d.GetOk("docker_credentials"); ok {
			vv := v.(map[string]interface{})
			appDeploy.AppPackage.DockerUsername = vv["username"].(string)
			appDeploy.AppPackage.DockerPassword = vv["password"].(string)
		}
	}

	// If b/g, only update if restage/restart not required
	if IsAppUpdateOnly(d) || (IsAppRestageNeeded(d) && !deployer.IsCreateNewApp()) || (IsAppRestartNeeded(d) && !deployer.IsCreateNewApp()) {
		log.Printf("\n--------------\n Updating app \n--------------\n")
		// Update application
		app, _, err := session.ClientV3.UpdateApplication(appUpdate)
		if err != nil {
			return diag.FromErr(err)
		}

		// Update process scale
		if processScaleRequired {
			proc, _, err := session.ClientV3.CreateApplicationProcessScale(appUpdate.GUID, processScale)
			if err != nil {
				return diag.FromErr(err)
			}

			appDeploy.Process = proc
		}

		// Service bindings change
		if d.HasChange("service_binding") {
			err := UnbindServiceInstances(d, session.ClientV3)
			if err != nil {
				return diag.FromErr(err)
			}

			// Rebind service bindings
			bindings, err := session.V3RunBinder.BindServiceInstances(appDeploy)
			if err != nil {
				return diag.FromErr(err)
			}
			appDeploy.ServiceBindings = bindings
		}

		// TODO: poll for app state after starting/stopping
		if d.HasChange("stopped") {
			// Update application state
			stopApplication := d.Get("stopped").(bool)

			var f func(appGUID string) (resources.Application, ccv3.Warnings, error)
			if stopApplication {
				f = session.ClientV3.UpdateApplicationStop
			} else {
				f = session.ClientV3.UpdateApplicationStart
			}
			app, _, err = f(appUpdate.GUID)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if d.HasChange("enable_ssh") {
			// Update enable SSH feature
			enableSSH := BoolToNullBool(d.Get("enable_ssh").(bool))
			_, err := session.ClientV3.UpdateAppFeature(appUpdate.GUID, enableSSH.Value, "ssh")
			if err != nil {
				return diag.FromErr(err)
			}

			// Set to appdeploy
			enabled, _, err := session.ClientV3.GetAppFeature(appUpdate.GUID, "ssh")
			if err != nil {
				return diag.FromErr(err)
			}
			appDeploy.EnableSSH = v3appdeployers.AppFeatureToNullBool(enabled)
		}

		if d.HasChange("environment") {
			// Update application's environment variables
			// Remove stale / externally set variables
			if currentEnv, err := session.BitsManager.GetAppEnvironmentVariables(appDeploy.App.GUID); err == nil {
				err := RemoveStaleEnviromentVariables(d, session, appDeploy.App.GUID, currentEnv)
				if err != nil {
					return diag.FromErr(err)
				}
			}

			// Set new environment variables
			createdEnv, _, err := session.BitsManager.UpdateAppEnvironment(app.GUID, appDeploy.EnvVars)
			if err != nil {
				return diag.FromErr(err)
			}

			appDeploy.EnvVars = createdEnv
		}

		if processUpdateRequired {
			// Get process guid
			currentAppProcess, _, err := session.ClientV3.GetApplicationProcessByType(appDeploy.App.GUID, constant.ProcessTypeWeb)
			if err != nil {
				return diag.FromErr(err)
			}

			processUpdate.GUID = currentAppProcess.GUID

			// Update application process
			proc, _, err := session.ClientV3.UpdateProcess(processUpdate)
			if err != nil {
				return diag.FromErr(err)
			}

			appDeploy.Process = proc
		}

		appDeploy.App = app
	}

	if IsAppRestageNeeded(d) || (deployer.IsCreateNewApp() && IsAppRestartNeeded(d)) {
		log.Printf("\n--------------\n Restaging app \n--------------\n")

		appResp, err := deployer.Restage(appDeploy)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Partial(false)
		AppDeployV3ToResourceData(d, appResp)
		return nil
	}

	if IsAppRestartNeeded(d) {
		log.Printf("\n--------------\n Restarting app \n--------------\n")

		var err error
		if d, ok := deployer.(v3appdeployers.CustomRestartStrategy); ok {
			err = d.Restart(appDeploy)
		} else {
			_, _, err = session.V3RunBinder.Restart(appDeploy)
		}
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
	return d.HasChange("path") || d.HasChange("source_code_hash") || d.HasChange("docker_image")
}

func IsAppUpdateOnly(d ResourceChanger) bool {
	if IsAppCodeChange(d) || IsAppRestageNeeded(d) || IsAppRestartNeeded(d) {
		return false
	}
	return d.HasChange("name") || d.HasChange("instances") ||
		d.HasChange("enable_ssh") || d.HasChange("stopped")
}

func IsAppRestageNeeded(d ResourceChanger) bool {
	return d.HasChange("buildpack") || d.HasChange("buildpacks") || d.HasChange("stack") ||
		d.HasChange("service_binding") || d.HasChange("environment")
}

func IsAppRestartNeeded(d ResourceChanger) bool {
	return d.HasChange("memory") || d.HasChange("disk_quota") ||
		d.HasChange("command") || d.HasChange("health_check_http_endpoint") || d.HasChange("health_check_type") ||
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
	if err != nil && !IsErrNotFound(err) {
		return diag.FromErr(err)
	}

	err = PollAsyncJob(PollingConfig{
		Session: session,
		JobURL:  jobURL,
	})
	return diag.FromErr(err)
}
