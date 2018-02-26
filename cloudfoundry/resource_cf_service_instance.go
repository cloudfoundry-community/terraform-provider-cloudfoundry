package cloudfoundry

import (
	"fmt"

	"encoding/json"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceServiceInstance() *schema.Resource {

	return &schema.Resource{

		Create: resourceServiceInstanceCreate,
		Read:   resourceServiceInstanceRead,
		Update: resourceServiceInstanceUpdate,
		Delete: resourceServiceInstanceDelete,

		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"service_plan": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"routes": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      resourceStringHash,
				Optional: true,
			},
			"json_params": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceServiceInstanceCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		id     string
		tags   []string
		params map[string]interface{}
	)
	name := d.Get("name").(string)
	servicePlan := d.Get("service_plan").(string)
	space := d.Get("space").(string)
	jsonParameters := d.Get("json_params").(string)

	for _, v := range d.Get("tags").([]interface{}) {
		tags = append(tags, v.(string))
	}

	if len(jsonParameters) > 0 {
		if err = json.Unmarshal([]byte(jsonParameters), &params); err != nil {
			return
		}
	}

	sm := session.ServiceManager()

	if id, err = sm.CreateServiceInstance(name, servicePlan, space, params, tags); err != nil {
		return
	}

	if err = serviceInstanceRouteCreate(d, sm, id, params); err != nil {
		sm.DeleteServiceInstance(id)
		return
	}

	session.Log.DebugMessage("New Service Instance : %# v", id)

	// TODO deal with asynchronous responses

	d.SetId(id)

	return
}

func resourceServiceInstanceRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("Reading Service Instance : %s", d.Id())

	sm := session.ServiceManager()
	var serviceInstance cfapi.CCServiceInstance

	serviceInstance, err = sm.ReadServiceInstance(d.Id())
	if err != nil {
		return
	}

	d.Set("name", serviceInstance.Name)
	d.Set("service_plan", serviceInstance.ServicePlanGUID)
	d.Set("space", serviceInstance.SpaceGUID)

	if serviceInstance.Tags != nil {
		tags := make([]interface{}, len(serviceInstance.Tags))
		for i, v := range serviceInstance.Tags {
			tags[i] = v
		}
		d.Set("tags", tags)
	} else {
		d.Set("tags", nil)
	}

	if err = serviceInstanceRouteRead(d, sm); err != nil {
		return
	}

	session.Log.DebugMessage("Read Service Instance : %# v", serviceInstance)

	return
}

func resourceServiceInstanceUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	sm := session.ServiceManager()

	session.Log.DebugMessage("begin resourceServiceInstanceUpdate")

	var (
		id, name string
		tags     []string
		params   map[string]interface{}
	)

	id = d.Id()
	name = d.Get("name").(string)
	servicePlan := d.Get("service_plan").(string)
	jsonParameters := d.Get("json_params").(string)

	if len(jsonParameters) > 0 {
		if err = json.Unmarshal([]byte(jsonParameters), &params); err != nil {
			return
		}
	}

	for _, v := range d.Get("tags").([]interface{}) {
		tags = append(tags, v.(string))
	}

	if _, err = sm.UpdateServiceInstance(id, name, servicePlan, params, tags); err != nil {
		return
	}
	if err != nil {
		return
	}

	if err = serviceInstanceRouteUpdate(d, sm, params); err != nil {
		return
	}

	return
}

func resourceServiceInstanceDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("begin resourceServiceInstanceDelete")

	sm := session.ServiceManager()
	if err = serviceInstanceRouteDelete(d, sm); err != nil {
		return
	}

	if err = sm.DeleteServiceInstance(d.Id()); err != nil {
		return
	}

	session.Log.DebugMessage("Deleted Service Instance : %s", d.Id())
	return
}

func serviceInstanceRouteCreate(
	d *schema.ResourceData,
	sm *cfapi.ServiceManager,
	serviceInstanceID string,
	params map[string]interface{}) (err error) {

	for _, r := range d.Get("routes").(*schema.Set).List() {
		routeGUID := r.(string)
		if err = sm.AddRouteServiceInstance(serviceInstanceID, routeGUID, params); err != nil {
			return
		}
	}
	return
}

func serviceInstanceRouteRead(d *schema.ResourceData, sm *cfapi.ServiceManager) (err error) {
	var routes []string
	var iroutes []interface{}

	if routes, err = sm.ReadServiceInstanceRoutes(d.Id()); err != nil {
		return
	}

	for _, r := range routes {
		iroutes = append(iroutes, r)
	}
	d.Set("routes", schema.NewSet(resourceStringHash, iroutes))

	return
}

func serviceInstanceRouteUpdate(
	d *schema.ResourceData,
	sm *cfapi.ServiceManager,
	params map[string]interface{}) (err error) {

	var cur_routes schema.Set
	var routes []string
	routes, err = sm.ReadServiceInstanceRoutes(d.Id())
	if err != nil {
		return
	}
	for _, v := range routes {
		cur_routes.Add(v)
	}
	rm, add := getListChanges(&cur_routes, d.Get("routes"))

	for _, r := range rm {
		if err = sm.DeleteRouteServiceInstance(d.Id(), r); err != nil {
			return
		}
	}
	for _, r := range add {
		if err = sm.AddRouteServiceInstance(d.Id(), r, params); err != nil {
			return
		}
	}

	return
}

func serviceInstanceRouteDelete(d *schema.ResourceData, sm *cfapi.ServiceManager) (err error) {
	var routes []string

	if routes, err = sm.ReadServiceInstanceRoutes(d.Id()); err != nil {
		return
	}

	for _, r := range routes {
		if err = sm.DeleteRouteServiceInstance(d.Id(), r); err != nil {
			return
		}
	}

	return
}
