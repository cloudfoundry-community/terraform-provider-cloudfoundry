package cloudfoundry

import (
	"fmt"

	"encoding/json"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceRouteServiceBinding() *schema.Resource {

	return &schema.Resource{
		Create: resourceRouteServiceBindingCreate,
		Read:   resourceRouteServiceBindingRead,
		Delete: resourceRouteServiceBindingDelete,

		Importer: &schema.ResourceImporter{
			State: resourceRouteServiceBindingImport,
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

func resourceRouteServiceBindingImport(d *schema.ResourceData, meta interface{}) (res []*schema.ResourceData, err error) {
	id := d.Id()
	if _, _, err = parseID(id); err != nil {
		return
	}
	return schema.ImportStatePassthrough(d, meta)
}

func resourceRouteServiceBindingCreate(d *schema.ResourceData, meta interface{}) (err error) {
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
			return err
		}
	}

	sm := session.ServiceManager()

	if err = sm.CreateRouteServiceBinding(serviceID, routeID, data); err != nil {
		return err
	}

	session.Log.DebugMessage("New Route Binding : %# v", id)
	d.SetId(computeID(serviceID, routeID))
	return nil
}

func resourceRouteServiceBindingRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("Reading RouteServiceBinding : %s", d.Id())

	serviceID, routeID, err := parseID(d.Id())
	if err != nil {
		return err
	}

	sm := session.ServiceManager()

	found, err := sm.HasRouteServiceBinding(serviceID, routeID)
	if err != nil {
		return err
	}
	if !found {
		d.SetId("")
		return fmt.Errorf("Route '%s' not found in service instance '%s'", routeID, serviceID)
	}

	d.Set("service_instance", serviceID)
	d.Set("route", routeID)
	session.Log.DebugMessage("Read Route Binding : %s", d.Id())
	return nil
}

func resourceRouteServiceBindingDelete(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("begin resourceRouteServiceBindingDelete")

	serviceID := d.Get("service_instance").(string)
	routeID := d.Get("route").(string)
	sm := session.ServiceManager()

	if err = sm.DeleteRouteServiceBinding(serviceID, routeID); err != nil {
		return err
	}

	session.Log.DebugMessage("Deleted RouteServiceBinding : %s", d.Id())
	return nil
}
