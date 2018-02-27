package cloudfoundry

import (
	"fmt"

	"encoding/json"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceRouteBinding() *schema.Resource {

	return &schema.Resource{
		Create: resourceRouteBindingCreate,
		Read:   resourceRouteBindingRead,
		Delete: resourceRouteBindingDelete,

		Importer: &schema.ResourceImporter{
			State: resourceRouteBindingImport,
		},

		Schema: map[string]*schema.Schema{
			"service_instance": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"json_params": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRouteBindingImport(d *schema.ResourceData, meta interface{}) (res []*schema.ResourceData, err error) {
	id := d.Id()
	if _, _, err = parseID(id); err != nil {
		return
	}
	return schema.ImportStatePassthrough(d, meta)
}

func resourceRouteBindingCreate(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		id   string
		data map[string]interface{}
	)

	serviceID := d.Get("service_instance").(string)
	routeID := d.Get("route").(string)
	params, okParams := d.GetOk("json_params")

	if okParams {
		if err = json.Unmarshal([]byte(params.(string)), &data); err != nil {
			return
		}
	}

	sm := session.ServiceManager()

	if err = sm.CreateRouteBinding(serviceID, routeID, data); err != nil {
		return
	}

	session.Log.DebugMessage("New Route Binding : %# v", id)
	d.SetId(computeID(serviceID, routeID))
	return
}

func resourceRouteBindingRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("Reading RouteBinding : %s", d.Id())

	serviceID, routeID, err := parseID(d.Id())
	if err != nil {
		return
	}

	sm := session.ServiceManager()

	found, err := sm.HasRouteBinding(serviceID, routeID)
	if err != nil {
		return
	}
	if found == false {
		d.SetId("")
		err = fmt.Errorf("Route '%s' not found in service instance '%s'", routeID, serviceID)
		return
	}

	d.Set("service_instance", serviceID)
	d.Set("route", routeID)
	session.Log.DebugMessage("Read Route Binding : %s", d.Id())
	return
}

func resourceRouteBindingDelete(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("begin resourceRouteBindingDelete")

	serviceID := d.Get("service_instance").(string)
	routeID := d.Get("route").(string)
	if err != nil {
		return
	}
	sm := session.ServiceManager()

	err = sm.DeleteRouteBinding(serviceID, routeID)
	if err != nil {
		return
	}

	session.Log.DebugMessage("Deleted RouteBinding : %s", d.Id())

	return
}
