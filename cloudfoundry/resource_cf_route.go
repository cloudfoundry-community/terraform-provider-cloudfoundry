package cloudfoundry

import (
	"fmt"
	"net/http"

	"github.com/cenkalti/backoff/v4"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceRoute() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteCreate,
		Read:   resourceRouteRead,
		Update: resourceRouteUpdate,
		Delete: resourceRouteDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceRouteRead),
		},

		Schema: map[string]*schema.Schema{

			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"port": &schema.Schema{
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"random_port"},
			},
			"random_port": &schema.Schema{
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"port"},
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"target": &schema.Schema{
				Type: schema.TypeSet,
				Set: func(v interface{}) int {
					elem := v.(map[string]interface{})
					return hashcode.String(fmt.Sprintf(
						"%s-%d",
						elem["app"],
						elem["port"],
					))
				},
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"port": &schema.Schema{
							Type:       schema.TypeInt,
							ConfigMode: schema.SchemaConfigModeAttr,
							Optional:   true,
							Computed:   true,
						},
					},
				},
			},
		},
	}
}

func resourceRouteCreate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	port := types.NullInt{}
	if v, ok := d.GetOk("port"); ok {
		port.Value = v.(int)
		port.IsSet = true
	}
	var route ccv2.Route

	operation := func() error {
		var err error
		route, _, err = session.ClientV2.CreateRoute(ccv2.Route{
			DomainGUID: d.Get("domain").(string),
			SpaceGUID:  d.Get("space").(string),
			Host:       d.Get("hostname").(string),
			Path:       d.Get("path").(string),
			Port:       port,
		}, d.Get("random_port").(bool))
		if err != nil {
			if unexpected, ok := err.(ccerror.V2UnexpectedResponseError); ok && unexpected.ResponseCode == http.StatusInternalServerError {
				return err
			}
			return backoff.Permanent(err)
		}
		return nil
	}
	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return err
	}
	// Delete route if an error occurs
	defer func() {
		e := &err
		if *e == nil {
			return
		}
		_, err = session.ClientV2.DeleteRoute(route.GUID)
		if err != nil {
			panic(err)
		}
	}()

	if err = setRouteArguments(session, route, d); err != nil {
		return err
	}

	if v, ok := d.GetOk("target"); ok {
		var t interface{}
		if t, err = addTargets(route.GUID, getListOfStructs(v.(*schema.Set).List()), session); err != nil {
			return err
		}
		d.Set("target", t)
	}

	d.SetId(route.GUID)
	return err
}

func resourceRouteRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	id := d.Id()

	route, _, err := session.ClientV2.GetRoute(id)
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	if err = setRouteArguments(session, route, d); err != nil {
		return err
	}

	if _, ok := d.GetOk("target"); !ok && !IsImportState(d) {
		return nil
	}
	mappingsTf := make([]map[string]interface{}, 0)
	tfTargets := d.Get("target").(*schema.Set).List()
	mappings, _, err := session.ClientV2.GetRouteMappings(ccv2.FilterEqual(constant.RouteGUIDFilter, d.Id()))
	if err != nil {
		return err
	}
	if IsImportState(d) {
		for _, mapping := range mappings {
			// if 0 it mean app port has been set to null which means it takes the first port found in app port definition
			if mapping.AppPort <= 0 {
				appID := mapping.AppGUID
				app, _, err := session.ClientV2.GetApplication(appID)
				if err != nil {
					return err
				}
				mapping.AppPort = app.Ports[0]
			}
			mappingsTf = append(mappingsTf, map[string]interface{}{
				"app":  mapping.AppGUID,
				"port": mapping.AppPort,
			})
		}
		if len(mappingsTf) > 0 {
			d.Set("target", mappingsTf)
		}
		return nil
	}

	final := make([]map[string]interface{}, 0)
	for _, tfTarget := range tfTargets {
		inside := false
		tmpT := tfTarget.(map[string]interface{})
		for _, mapping := range mappings {
			// if 0 it mean app port has been set to null which means it takes the first port found in app port definition
			if mapping.AppPort <= 0 {
				appID := mapping.AppGUID
				app, _, err := session.ClientV2.GetApplication(appID)
				if err != nil {
					return err
				}
				mapping.AppPort = app.Ports[0]
			}
			if mapping.AppGUID == tmpT["app"] && mapping.AppPort == tmpT["port"] {
				inside = true
				tmpT["port"] = mapping.AppPort
				tmpT["app"] = mapping.AppGUID
				break
			}
		}
		if inside {
			final = append(final, tmpT)
		}
	}
	d.Set("target", final)
	return nil
}

func resourceRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	port := types.NullInt{}
	if v, ok := d.GetOk("port"); ok {
		port.Value = v.(int)
		port.IsSet = true
	}
	if d.HasChange("domain") || d.HasChange("space") || d.HasChange("hostname") {
		route, _, err := session.ClientV2.UpdateRoute(ccv2.Route{
			GUID:       d.Id(),
			DomainGUID: d.Get("domain").(string),
			SpaceGUID:  d.Get("space").(string),
			Host:       d.Get("hostname").(string),
			Path:       d.Get("path").(string),
			Port:       port,
		})
		if err != nil {
			return err
		}
		err = setRouteArguments(session, route, d)
		if err != nil {
			return err
		}
	}

	if d.HasChange("target") {
		old, new := d.GetChange("target")
		remove, add := getListMapChanges(old, new, func(source, item map[string]interface{}) bool {
			return source["app"] == item["app"] && source["port"] == item["port"]
		})
		err := removeTargets(d.Id(), remove, session)
		if err != nil {
			return err
		}

		t, err := addTargets(d.Id(), add, session)
		if err != nil {
			return err
		}
		d.Set("target", t)
	}
	return nil
}

func resourceRouteDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	if targets, ok := d.GetOk("target"); ok {
		err := removeTargets(d.Id(), getListOfStructs(targets.(*schema.Set).List()), session)
		if err != nil {
			return err
		}
	}
	_, err := session.ClientV2.DeleteRoute(d.Id())
	return err
}

func setRouteArguments(session *managers.Session, route ccv2.Route, d *schema.ResourceData) (err error) {

	d.Set("domain", route.DomainGUID)
	d.Set("space", route.SpaceGUID)
	d.Set("hostname", route.Host)
	if route.Port.IsSet {
		d.Set("port", route.Port.Value)
	}
	d.Set("path", route.Path)

	domain, _, err := session.ClientV2.GetSharedDomain(route.DomainGUID)
	if err != nil || domain.GUID == "" {
		domain, _, err = session.ClientV2.GetPrivateDomain(route.DomainGUID)
		if err != nil {
			return err
		}
	}
	port := ""
	if route.Port.IsSet && route.Port.Value > 0 && domain.RouterGroupGUID != "" {
		port = fmt.Sprintf(":%d", route.Port.Value)
	}
	endpoint := fmt.Sprintf("%s.%s%s", route.Host, domain.Name, port)
	if route.Path != "" {
		endpoint += "/" + route.Path
	}
	d.Set("endpoint", endpoint)
	return nil
}

func addTargets(id string, add []map[string]interface{}, session *managers.Session) ([]map[string]interface{}, error) {
	targets := make([]map[string]interface{}, 0)
	for _, t := range add {
		appID := t["app"].(string)
		var port int
		// if 0 it mean app port has been set to null which means it takes the first port found in app port definition
		if v, ok := t["port"]; ok && v.(int) > 0 {
			port = v.(int)
		} else {
			app, _, err := session.ClientV2.GetApplication(appID)
			if err != nil {
				return targets, err
			}
			port = app.Ports[0]
			t["port"] = port
		}
		_, _, err := session.ClientV2.CreateRouteMapping(appID, id, &port)
		if err != nil {
			return targets, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func removeTargets(id string, delete []map[string]interface{}, session *managers.Session) error {

	for _, t := range delete {
		appID := t["app"].(string)
		app, _, err := session.ClientV2.GetApplication(appID)
		if err != nil {
			return err
		}
		defaultAppPort := app.Ports[0]
		mappings, _, err := session.ClientV2.GetRouteMappings(filterAppGuid(t["app"].(string)), filterRouteGuid(id))
		if err != nil {
			return err
		}
		for _, mapping := range mappings {
			// if 0 it mean app port has been set to null which means it takes the first port found in app port definition
			if mapping.AppPort <= 0 {
				mapping.AppPort = defaultAppPort
			}
			if mapping.AppPort != t["port"] {
				continue
			}
			_, err := session.ClientV2.DeleteRouteMapping(mapping.GUID)
			if err != nil && !IsErrNotFound(err) {
				return err
			}
		}
	}
	return nil
}
