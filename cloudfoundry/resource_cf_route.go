package cloudfoundry

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceRoute() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteCreate,
		Read:   resourceRouteRead,
		Update: resourceRouteUpdate,
		Delete: resourceRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceRouteImport,
		},

		Schema: map[string]*schema.Schema{

			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
				ConflictsWith: []string{"path", "random_port"},
			},
			"random_port": &schema.Schema{
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"path", "port"},
			},
			"path": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"port"},
			},
			"endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"target": &schema.Schema{
				Type:     schema.TypeSet,
				Set:      routeTargetHash,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"port": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  8080,
						},
						"mapping_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func routeTargetHash(d interface{}) int {

	a := d.(map[string]interface{})["app"].(string)

	p := ""
	if v, ok := d.(map[string]interface{})["port"]; ok {
		p = strconv.Itoa(v.(int))
	}

	return hashcode.String(a + p)
}

func resourceRouteCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	route := cfapi.CCRoute{
		DomainGUID: d.Get("domain").(string),
		SpaceGUID:  d.Get("space").(string),
	}

	if v, ok := d.GetOk("hostname"); ok {
		vv := v.(string)
		route.Hostname = &vv
	}
	if v, ok := d.GetOk("port"); ok {
		vv := v.(int)
		route.Port = &vv
	}
	if v, ok := d.GetOk("path"); ok {
		vv := v.(string)
		route.Path = &vv
	}

	randomPort := false
	if v, ok := d.GetOk("random_port"); ok {
		randomPort = v.(bool)
	}

	rm := session.RouteManager()

	// Create route
	if route, err = rm.CreateRoute(route, randomPort); err != nil {
		return err
	}
	// Delete route if an error occurs
	defer func() {
		e := &err
		if *e != nil {
			rm.DeleteRoute(route.ID)
		}
	}()

	if err = setRouteArguments(session, route, d); err != nil {
		return
	}

	if v, ok := d.GetOk("target"); ok {
		var t interface{}
		if t, err = addTargets(route.ID, getListOfStructs(v.(*schema.Set).List()), rm, session.Log); err != nil {
			return
		}
		d.Set("target", t)
		session.Log.DebugMessage("Mapped route targets: %# v", d.Get("target"))
	}

	d.SetId(route.ID)
	return
}

func resourceRouteRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id := d.Id()
	rm := session.RouteManager()

	var route cfapi.CCRoute
	if route, err = rm.ReadRoute(id); err != nil {
		return
	}
	if err = setRouteArguments(session, route, d); err != nil {
		return
	}

	if _, ok := d.GetOk("target"); ok || IsImportState(d) {
		var mappings []map[string]interface{}
		if mappings, err = rm.ReadRouteMappingsByRoute(id); err != nil {
			return
		}
		if len(mappings) > 0 {
			d.Set("target", mappings)
		}
	}
	return
}

func resourceRouteUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	rm := session.RouteManager()

	route := cfapi.CCRoute{
		ID: d.Id(),
	}

	update := false
	route.DomainGUID = *getChangedValueString("domain", &update, d)
	route.SpaceGUID = *getChangedValueString("space", &update, d)
	route.Hostname = getChangedValueString("hostname", &update, d)

	if update {
		if route, err = rm.UpdateRoute(route); err != nil {
			return err
		}
		if err = setRouteArguments(session, route, d); err != nil {
			return
		}
	}

	if d.HasChange("target") {
		old, new := d.GetChange("target")
		session.Log.DebugMessage("Old route mappings state:: %# v", old)
		session.Log.DebugMessage("New route mappings state:: %# v", new)

		if err = removeTargets(getListOfStructs(old.(*schema.Set).List()), rm, session.Log); err != nil {
			return
		}

		var t interface{}
		if t, err = addTargets(route.ID, getListOfStructs(new.(*schema.Set).List()), rm, session.Log); err != nil {
			return
		}
		d.Set("target", t)
		session.Log.DebugMessage("Updated route target mappings: %# v", d.Get("target"))
	}
	return
}

func resourceRouteDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	rm := session.RouteManager()

	if targets, ok := d.GetOk("target"); ok {
		err = removeTargets(getListOfStructs(targets.(*schema.Set).List()), rm, session.Log)
	}
	if err = rm.DeleteRoute(d.Id()); err != nil {
		return
	}
	return
}

func setRouteArguments(session *cfapi.Session, route cfapi.CCRoute, d *schema.ResourceData) (err error) {

	d.Set("domain", route.DomainGUID)
	d.Set("space", route.SpaceGUID)
	if route.Hostname != nil {
		d.Set("hostname", route.Hostname)
	}
	if route.Port != nil {
		d.Set("port", route.Port)
	}
	if route.Path != nil {
		d.Set("path", route.Path)
	}

	domain, err := session.DomainManager().FindDomain(route.DomainGUID)
	if err != nil {
		return
	}
	if route.Port != nil && *route.Port > 0 {
		d.Set("endpoint", fmt.Sprintf("%s:%d", domain.Name, *route.Port))
	} else if route.Path == nil || len(*route.Path) == 0 {
		d.Set("endpoint", fmt.Sprintf("%s.%s", *route.Hostname, domain.Name))
	} else {
		d.Set("endpoint", fmt.Sprintf("%s.%s/%s", *route.Hostname, domain.Name, *route.Path))
	}

	return
}

func addTargets(id string, add []map[string]interface{},
	rm *cfapi.RouteManager, log *cfapi.Logger) (targets []map[string]interface{}, err error) {

	var (
		appID, mappingID string
		port             *int
	)

	for _, t := range add {

		appID = t["app"].(string)
		port = nil
		if v, ok := t["port"]; ok {
			vv := v.(int)
			port = &vv
		}
		if mappingID, err = rm.CreateRouteMapping(id, appID, port); err != nil {
			return
		}
		t["mapping_id"] = mappingID
		targets = append(targets, t)

		log.DebugMessage("Created route mapping with id '%s' to app instance '%s'.", mappingID, appID)
	}
	return
}

func removeTargets(delete []map[string]interface{},
	rm *cfapi.RouteManager, log *cfapi.Logger) error {

	for _, t := range delete {

		appID := t["app"].(string)
		mappingID := t["mapping_id"].(string)
		log.DebugMessage("Deleting route mapping with id '%s' to app instance '%s'.", mappingID, appID)

		if len(mappingID) > 0 {
			log.DebugMessage("Deleting route mapping with id '%s' to app instance '%s'.", mappingID, appID)
			if err := rm.DeleteRouteMapping(mappingID); err != nil {
				return err
			}
		} else {
			log.DebugMessage("Ignoring mapping app instance '%s' as no corresponding mapping id was found.", appID)
		}
	}
	return nil
}

func resourceRouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	session := meta.(*cfapi.Session)

	if session == nil {
		return nil, fmt.Errorf("client is nil")
	}

	rm := session.RouteManager()

	route, err := rm.ReadRoute(d.Id())

	if err != nil {
		return nil, err
	}

	d.Set("domain", route.DomainGUID)
	d.Set("space", route.SpaceGUID)
	d.Set("path", route.Path)
	d.Set("hostname", route.Hostname)
	d.Set("id", route.ID)

	domain, err := session.DomainManager().FindDomain(route.DomainGUID)
	if err != nil {
		return nil, err
	}

	if route.Port != nil && *route.Port > 0 {
		d.Set("endpoint", fmt.Sprintf("%s:%d", domain.Name, *route.Port))
	} else if route.Path == nil || len(*route.Path) == 0 {
		d.Set("endpoint", fmt.Sprintf("%s.%s", *route.Hostname, domain.Name))
	} else {
		d.Set("endpoint", fmt.Sprintf("%s.%s/%s", *route.Hostname, domain.Name, *route.Path))
	}

	return []*schema.ResourceData{d}, nil
}
