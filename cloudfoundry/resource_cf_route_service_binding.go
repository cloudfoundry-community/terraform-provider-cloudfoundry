package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	return ImportRead(resourceRouteServiceBindingRead)(d, meta)
}

func resourceRouteServiceBindingCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	var data map[string]interface{}

	serviceID := d.Get("service_instance").(string)
	routeID := d.Get("route").(string)
	params, okParams := d.GetOk("json_params")

	if okParams {
		if err := json.Unmarshal([]byte(params.(string)), &data); err != nil {
			return err
		}
	}
	_, err := session.ClientV2.CreateServiceBindingRoute(serviceID, routeID, data)
	if err != nil {
		return err
	}

	d.SetId(computeID(serviceID, routeID))
	return nil
}

func resourceRouteServiceBindingRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	serviceID, routeID, err := parseID(d.Id())
	if err != nil {
		return err
	}
	routes, _, err := session.ClientV2.GetServiceBindingRoutes(serviceID)
	if err != nil {
		return err
	}
	found := false
	for _, route := range routes {
		if route.GUID == routeID {
			found = true
			break
		}
	}
	if !found {
		d.SetId("")
		return fmt.Errorf("Route '%s' not found in service instance '%s'", routeID, serviceID)
	}

	d.Set("service_instance", serviceID)
	d.Set("route", routeID)
	return nil
}

func resourceRouteServiceBindingDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	serviceID := d.Get("service_instance").(string)
	routeID := d.Get("route").(string)
	_, err := session.ClientV2.DeleteServiceBindingRoute(serviceID, routeID)
	return err
}
