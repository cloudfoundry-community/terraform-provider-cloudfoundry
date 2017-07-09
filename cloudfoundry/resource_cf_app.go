package cloudfoundry

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/cf/terminal"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/repo"
)

func resourceApp() *schema.Resource {

	return &schema.Resource{

		Create: resourceAppCreate,
		Read:   resourceAppRead,
		Update: resourceAppUpdate,
		Delete: resourceAppDelete,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
			},
			"stopped": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"url": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"git", "github_release"},
			},
			"git": &schema.Schema{
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"url", "github_release"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"branch": &schema.Schema{
							Type:          schema.TypeString,
							Optional:      true,
							Default:       "master",
							ConflictsWith: []string{"git.tag"},
						},
						"tag": &schema.Schema{
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"git.branch"},
						},
						"user": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"password": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"key": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"github_release": &schema.Schema{
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"url", "git"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"owner": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"repo": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"token": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"version": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"filename": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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
							Type:     schema.TypeMap,
							Optional: true,
						},
						"credentials": &schema.Schema{
							Type:     schema.TypeMap,
							Computed: true,
						},
						"binding_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"route": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_route": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"default_route_mapping_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"stage_route": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"stage_route_mapping_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"live_route": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"live_route_mapping_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"validation_script": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"version": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"environment": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
			"health_check_http_endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"health_check_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateAppHealthCheckType,
			},
			"health_check_timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func serviceBindingHash(d interface{}) int {
	return hashcode.String(d.(map[string]interface{})["service_instance"].(string))
}

func validateAppHealthCheckType(v interface{}, k string) (ws []string, errs []error) {
	value := v.(string)
	if value != "port" && value != "process" && value != "http" && value != "none" {
		errs = append(errs, fmt.Errorf("%q must be one of 'port', 'process', 'http' or 'none'", k))
	}
	return
}

func resourceAppCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.AppManager()
	rm := session.RouteManager()

	var (
		v  interface{}
		ok bool

		app cfapi.CCApp

		appPath string

		defaultRoute, stageRoute, liveRoute string
		isBlueGreen                         bool

		serviceBindings    []map[string]interface{}
		hasServiceBindings bool

		routeConfig    map[string]interface{}
		hasRouteConfig bool
	)

	app.Name = d.Get("name").(string)
	app.SpaceGUID = d.Get("space").(string)
	if v, ok = d.GetOk("ports"); ok {
		p := []int{}
		for _, vv := range v.(*schema.Set).List() {
			p = append(p, vv.(int))
		}
		app.Ports = &p
	}
	if v, ok = d.GetOk("instances"); ok {
		vv := v.(int)
		app.Instances = &vv
	}
	if v, ok = d.GetOk("memory"); ok {
		vv := v.(int)
		app.Memory = &vv
	}
	if v, ok = d.GetOk("disk_quota"); ok {
		vv := v.(int)
		app.DiskQuota = &vv
	}
	if v, ok = d.GetOk("stack"); ok {
		vv := v.(string)
		app.StackGUID = &vv
	}
	if v, ok = d.GetOk("buildpack"); ok {
		vv := v.(string)
		app.Buildpack = &vv
	}
	if v, ok = d.GetOk("command"); ok {
		vv := v.(string)
		app.Command = &vv
	}
	if v, ok = d.GetOk("enable_ssh"); ok {
		vv := v.(bool)
		app.EnableSSH = &vv
	}
	if v, ok = d.GetOk("health_check_http_endpoint"); ok {
		vv := v.(string)
		app.HealthCheckHTTPEndpoint = &vv
	}
	if v, ok = d.GetOk("health_check_type"); ok {
		vv := v.(string)
		app.HealthCheckType = &vv
	}
	if v, ok = d.GetOk("health_check_timeout"); ok {
		vv := v.(int)
		app.HealthCheckTimeout = &vv
	}
	if v, ok = d.GetOk("environment"); ok {
		vv := v.(map[string]interface{})
		app.Environment = &vv
	}

	// Download application binary / source asynchronously
	prepare := make(chan error)
	go func() {
		appPath, err = prepareApp(app, d, session.Log)
		prepare <- err
	}()

	if v, hasRouteConfig = d.GetOk("route"); hasRouteConfig {

		routeConfig = v.([]interface{})[0].(map[string]interface{})
		isBlueGreen = false

		if defaultRoute, err = validateRoute(routeConfig, "default_route", rm); err != nil {
			return
		}
		if stageRoute, err = validateRoute(routeConfig, "stage_route", rm); err != nil {
			return
		}
		if liveRoute, err = validateRoute(routeConfig, "live_route", rm); err != nil {
			return
		}

		if len(stageRoute) > 0 && len(liveRoute) > 0 {
			isBlueGreen = true
		} else if len(stageRoute) > 0 || len(liveRoute) > 0 {
			err = fmt.Errorf("both 'stage_route' and 'live_route' need to be provided to deploy the app using blue-green routing")
			return
		}
	}

	// Create application
	if app, err = am.CreateApp(app); err != nil {
		return
	}
	// Delete application if an error occurs
	defer func() {
		e := &err
		if *e != nil {
			am.DeleteApp(app.ID, true)
		}
	}()

	// Upload application binary / source
	// asynchronously once download has completed
	if err = <-prepare; err != nil {
		return
	}
	upload := make(chan error)
	go func() {
		err = am.UploadApp(app, appPath)
		upload <- err
	}()

	// Bind services
	if v, hasServiceBindings = d.GetOk("service_binding"); hasServiceBindings {
		if serviceBindings, err = addServiceBindings(app.ID, getListOfStructs(v), am, session.Log); err != nil {
			return
		}
	}

	// Bind default route
	if len(defaultRoute) > 0 {
		var mappingID string
		if mappingID, err = rm.CreateRouteMapping(defaultRoute, app.ID, nil); err != nil {
			return
		}
		routeConfig["default_route_mapping_id"] = mappingID
	}

	timeout := time.Second * time.Duration(d.Get("timeout").(int))
	stopped := d.Get("stopped").(bool)

	// Start application if not stopped
	// state once upload has completed
	if err = <-upload; err != nil {
		return
	}
	if !stopped {
		if err = am.StartApp(app.ID, timeout); err != nil {
			return
		}

		// Execute blue-green validation
		if isBlueGreen {
		}
	}

	if app, err = am.ReadApp(app.ID); err != nil {
		return
	}
	d.SetId(app.ID)

	session.Log.DebugMessage("Created app state: %# v", app)
	setAppArguments(app, d)

	if hasServiceBindings {
		d.Set("service_binding", serviceBindings)
		session.Log.DebugMessage("Created service bindings: %# v", d.Get("service_binding"))
	}
	if hasRouteConfig {
		d.Set("route", []map[string]interface{}{routeConfig})
		session.Log.DebugMessage("Created routes: %# v", d.Get("route"))
	}

	return
}

func resourceAppRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id := d.Id()
	am := session.AppManager()

	var app cfapi.CCApp
	if app, err = am.ReadApp(id); err != nil {
		return
	}
	setAppArguments(app, d)
	return
}

func resourceAppUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.AppManager()
	rm := session.RouteManager()

	app := cfapi.CCApp{
		ID: d.Id(),
	}
	update := false
	restage := false
	restart := false

	app.Name = *getChangedValueString("name", &update, d)
	app.SpaceGUID = *getChangedValueString("space", &update, d)
	app.Ports = getChangedValueIntList("ports", &update, d)
	app.Instances = getChangedValueInt("instances", &update, d)
	app.Memory = getChangedValueInt("memory", &update, d)
	app.DiskQuota = getChangedValueInt("disk_quota", &update, d)
	app.Command = getChangedValueString("command", &update, d)
	app.EnableSSH = getChangedValueBool("enable_ssh", &update, d)
	app.HealthCheckHTTPEndpoint = getChangedValueString("health_check_http_endpoint", &update, d)
	app.HealthCheckType = getChangedValueString("health_check_type", &update, d)
	app.HealthCheckTimeout = getChangedValueInt("health_check_timeout", &update, d)

	app.Buildpack = getChangedValueString("buildpack", &restage, d)
	app.Environment = getChangedValueMap("environment", &restage, d)

	if update || restart || restage {
		if app, err = am.UpdateApp(app); err != nil {
			return
		}
		setAppArguments(app, d)
	}

	if d.HasChange("service_binding") {

		old, new := d.GetChange("service_binding")
		session.Log.DebugMessage("Old service bindings state:: %# v", old)
		session.Log.DebugMessage("New service bindings state:: %# v", new)

		bindingsToDelete, bindingsToAdd := getListChangedSchemaLists(old.([]interface{}), new.([]interface{}))
		session.Log.DebugMessage("Service bindings to be deleted: %# v", bindingsToDelete)
		session.Log.DebugMessage("Service bindings to be added: %# v", bindingsToAdd)

		if err = removeServiceBindings(bindingsToDelete, am, session.Log); err != nil {
			return
		}

		var added []map[string]interface{}
		if added, err = addServiceBindings(app.ID, bindingsToAdd, am, session.Log); err != nil {
			return
		}
		if len(added) > 0 {
			if new != nil {
				for _, b := range new.([]interface{}) {
					bb := b.(map[string]interface{})

					for _, a := range added {
						if bb["service_instance"] == a["service_instance"] {
							bb["binding_id"] = a["binding_id"]
							bb["credentials"] = a["credentials"]
							break
						}
					}
				}
				d.Set("service_binding", new)
			}
		}
		restage = true
	}

	if d.HasChange("route") {
		old, new := d.GetChange("route")

		var (
			oldRouteConfig, newRouteConfig map[string]interface{}
			mappingID                      string
		)

		oldA := old.([]interface{})
		if len(oldA) == 1 {
			oldRouteConfig = oldA[0].(map[string]interface{})
		} else {
			oldRouteConfig = make(map[string]interface{})
		}
		newA := new.([]interface{})
		if len(newA) == 1 {
			newRouteConfig = newA[0].(map[string]interface{})
		} else {
			newRouteConfig = make(map[string]interface{})
		}

		for _, r := range []string{
			"default_route",
			"stage_route",
			"live_route",
		} {
			if _, err = validateRoute(newRouteConfig, r, rm); err != nil {
				return
			}
			if mappingID, err = updateMapping(oldRouteConfig, newRouteConfig, r, app.ID, rm); err != nil {
				return
			}
			if len(mappingID) > 0 {
				newRouteConfig[r+"_mapping_id"] = mappingID
			}
		}
	}

	if d.HasChange("url") || d.HasChange("git") || d.HasChange("github_release") {

		var appPath string
		if appPath, err = prepareApp(app, d, session.Log); err != nil {
			return
		}
		if err = am.UploadApp(app, appPath); err != nil {
			return err
		}
		restage = true
	}

	timeout := time.Second * time.Duration(d.Get("timeout").(int))

	if restage {
		if err = am.RestageApp(app.ID, timeout); err != nil {
			return
		}
	}
	if d.HasChange("stopped") {

		if d.Get("stopped").(bool) {
			if err = am.StopApp(app.ID, timeout); err != nil {
				return
			}
		} else {
			if err = am.StartApp(app.ID, timeout); err != nil {
				return
			}
		}
	} else if restage {
		err = am.WaitForAppToStart(app, timeout)
	}
	return
}

func resourceAppDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.AppManager()
	rm := session.RouteManager()

	if v, ok := d.GetOk("service_binding"); ok {
		if err = removeServiceBindings(getListOfStructs(v), am, session.Log); err != nil {
			return
		}
	}
	if v, ok := d.GetOk("route"); ok {

		routeConfig := v.([]interface{})[0].(map[string]interface{})

		for _, r := range []string{
			"default_route_mapping_id",
			"stage_route_mapping_id",
			"live_route_mapping_id",
		} {
			if v, ok := routeConfig[r]; ok {
				mappingID := v.(string)
				if len(mappingID) > 0 {
					if err = rm.DeleteRouteMapping(v.(string)); err != nil {
						return
					}
				}
			}
		}
	}
	err = am.DeleteApp(d.Id(), false)
	return
}

func setAppArguments(app cfapi.CCApp, d *schema.ResourceData) {

	d.Set("name", app.Name)
	d.Set("space", app.SpaceGUID)
	if app.Instances != nil {
		d.Set("instances", app.Instances)
	}
	if app.Memory != nil {
		d.Set("memory", app.Memory)
	}
	if app.DiskQuota != nil {
		d.Set("disk_quota", app.DiskQuota)
	}
	if app.StackGUID != nil {
		d.Set("stack", app.StackGUID)
	}
	if app.Buildpack != nil {
		d.Set("buildpack", app.Buildpack)
	}
	if app.Command != nil {
		d.Set("command", app.Command)
	}
	if app.EnableSSH != nil {
		d.Set("enable_ssh", app.EnableSSH)
	}
	if app.HealthCheckHTTPEndpoint != nil {
		d.Set("health_check_http_endpoint", app.HealthCheckHTTPEndpoint)
	}
	if app.HealthCheckType != nil {
		d.Set("health_check_type", app.HealthCheckType)
	}
	if app.HealthCheckTimeout != nil {
		d.Set("health_check_timeout", app.HealthCheckTimeout)
	}
	if app.Environment != nil {
		d.Set("environment", app.Environment)
	}

	d.Set("timeout", d.Get("timeout"))
	d.Set("stopped", d.Get("stopped"))

	ports := []interface{}{}
	for _, p := range *app.Ports {
		ports = append(ports, p)
	}
	d.Set("ports", schema.NewSet(resourceIntegerSet, ports))
}

func prepareApp(app cfapi.CCApp, d *schema.ResourceData, log *cfapi.Logger) (path string, err error) {

	if v, ok := d.GetOk("url"); ok {
		url := v.(string)

		if strings.HasPrefix(url, "file://") {
			path = url[7:]
		} else {

			var (
				resp *http.Response

				in  io.ReadCloser
				out *os.File
			)

			if out, err = ioutil.TempFile("", "cfapp"); err != nil {
				return
			}

			log.UI.Say("Downloading application %s from url %s.", terminal.EntityNameColor(app.Name), url)

			if resp, err = http.Get(url); err != nil {
				return
			}
			in = resp.Body
			if _, err = io.Copy(out, in); err != nil {
				return
			}
			if err = out.Close(); err != nil {
				return
			}

			path = out.Name()
		}

	} else {
		log.UI.Say("Retrieving application %s source / binary.", terminal.EntityNameColor(app.Name))

		var repository repo.Repository
		if repository, err = getRepositoryFromConfig(d); err != nil {
			return
		}

		if _, ok := d.GetOk("github_release"); ok {
			path = filepath.Dir(repository.GetPath())
		} else {
			path = repository.GetPath()
		}
	}
	if err != nil {
		return "", err
	}

	log.UI.Say("Application downloaded to: %s", path)
	return
}

func validateRoute(routeConfig map[string]interface{}, route string, rm *cfapi.RouteManager) (routeID string, err error) {

	if v, ok := routeConfig[route]; ok {

		routeID = v.(string)

		var mappings []map[string]interface{}
		if mappings, err = rm.ReadRouteMappingsByRoute(routeID); err == nil && len(mappings) > 0 {
			err = fmt.Errorf(
				"route with id %s is already mapped. routes specificed in the 'route' argument can only be mapped to one 'cf_app' resource",
				routeID)
		}
	}
	return
}

func updateMapping(old map[string]interface{}, new map[string]interface{},
	route, appID string, rm *cfapi.RouteManager) (mappingID string, err error) {

	var (
		oldRouteID, newRouteID string
	)

	if v, ok := old[route]; ok {
		oldRouteID = v.(string)
	}
	if v, ok := new[route]; ok {
		newRouteID = v.(string)
	}

	if oldRouteID != newRouteID {
		if len(oldRouteID) > 0 {
			if v, ok := old[route+"_mapping_id"]; ok {
				if err = rm.DeleteRouteMapping(v.(string)); err != nil {
					return
				}
			}
		}
		if len(newRouteID) > 0 {
			if mappingID, err = rm.CreateRouteMapping(newRouteID, appID, nil); err != nil {
				return
			}
		}
	}
	return
}

func addServiceBindings(id string, add []map[string]interface{},
	am *cfapi.AppManager, log *cfapi.Logger) (bindings []map[string]interface{}, err error) {

	var (
		serviceInstanceID, bindingID string
		params                       *map[string]interface{}

		credentials        map[string]interface{}
		bindingCredentials map[string]interface{}
	)

	for _, b := range add {

		serviceInstanceID = b["service_instance"].(string)
		params = nil
		if v, ok := b["params"]; ok {
			vv := v.(map[string]interface{})
			params = &vv
		}
		if bindingID, bindingCredentials, err = am.CreateServiceBinding(id, serviceInstanceID, params); err != nil {
			return
		}
		b["binding_id"] = bindingID

		credentials = b["credentials"].(map[string]interface{})
		for k, v := range bindingCredentials {
			credentials[k] = fmt.Sprintf("%v", v)
		}

		bindings = append(bindings, b)
		log.DebugMessage("Created binding with id '%s' for service instance '%s'.", bindingID, serviceInstanceID)
	}
	return
}

func removeServiceBindings(delete []map[string]interface{},
	am *cfapi.AppManager, log *cfapi.Logger) error {

	for _, b := range delete {

		serviceInstanceID := b["service_instance"].(string)
		bindingID := b["binding_id"].(string)

		if len(bindingID) > 0 {
			log.DebugMessage("Deleting binding with id '%s' for service instance '%s'.", bindingID, serviceInstanceID)
			if err := am.DeleteServiceBinding(bindingID); err != nil {
				return err
			}
		} else {
			log.DebugMessage("Ignoring binding for service instance '%s' as no corresponding binding id was found.", serviceInstanceID)
		}
	}
	return nil
}
