package cloudfoundry

import (
	"context"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRouteServiceBinding() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceRouteServiceBindingCreate,
		ReadContext:   resourceRouteServiceBindingRead,
		DeleteContext: resourceRouteServiceBindingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRouteServiceBindingImport,
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

func resourceRouteServiceBindingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (res []*schema.ResourceData, err error) {
	id := d.Id()
	if _, _, err = parseID(id); err != nil {
		return
	}
	return ImportReadContext(resourceRouteServiceBindingRead)(ctx, d, meta)
}

func resourceRouteServiceBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	var data map[string]interface{}

	serviceID := d.Get("service_instance").(string)
	routeID := d.Get("route").(string)
	params, okParams := d.GetOk("json_params")

	if okParams {
		if err := json.Unmarshal([]byte(params.(string)), &data); err != nil {
			return diag.FromErr(err)
		}
	}
	jobGUID, _, err := session.ClientGo.ServiceRouteBindings.Create(context.Background(), &resource.ServiceRouteBindingCreate{
		Relationships: resource.ServiceRouteBindingRelationships{
			// ServiceInstance ToOneRelationship `json:"service_instance"`
			// // The route that the service instance is bound to
			// Route ToOneRelationship `json:"route"`
			ServiceInstance: resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: serviceID,
				},
			},
			Route: resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: routeID,
				},
			},
		},
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if jobGUID != "" {
		err = session.ClientGo.Jobs.PollComplete(context.Background(), jobGUID, nil)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(computeID(serviceID, routeID))
	return nil
}

func resourceRouteServiceBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	serviceID, routeID, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	options := client.NewServiceRouteBindingListOptions()
	options.ServiceInstanceGUIDs = client.Filter{Values: []string{serviceID}}

	routeBindings, err := session.ClientGo.ServiceRouteBindings.ListAll(context.Background(), options)
	if err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, routeBinding := range routeBindings {
		if routeBinding.Relationships.Route.Data.GUID == routeID {
			found = true
			break
		}
	}
	if !found {
		d.SetId("")
		return diag.Errorf("Route '%s' not found in service instance '%s'", routeID, serviceID)
	}

	d.Set("service_instance", serviceID)
	d.Set("route", routeID)
	return nil
}

func resourceRouteServiceBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	serviceID := d.Get("service_instance").(string)
	routeID := d.Get("route").(string)

	var err error

	options := client.NewServiceRouteBindingListOptions()
	options.ServiceInstanceGUIDs = client.Filter{Values: []string{serviceID}}
	options.RouteGUIDs = client.Filter{Values: []string{routeID}}
	routeBindings, err := session.ClientGo.ServiceRouteBindings.ListAll(context.Background(), options)
	if err != nil {
		return diag.FromErr(err)
	}

	routeBindingID := ""
	for _, routeBinding := range routeBindings {
		if routeBinding.Relationships.Route.Data.GUID == routeID &&
			routeBinding.Relationships.ServiceInstance.Data.GUID == serviceID {
			routeBindingID = routeBinding.GUID
		}
	}
	if routeBindingID != "" {
		var jobGUID string
		jobGUID, err = session.ClientGo.ServiceRouteBindings.Delete(context.Background(), routeBindingID)
		if err != nil {
			return diag.FromErr(err)
		}
		if jobGUID != "" {
			err = session.ClientGo.Jobs.PollComplete(context.Background(), jobGUID, nil)
		}
	}

	return diag.FromErr(err)
}
