package cloudfoundry

import (
	"context"
	"log"
	"strings"

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
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceRouteServiceBindingResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: upgradeStateRouteServiceBindingStateV0toV1ChangeID,
				Version: 0,
			},
		},
	}
}

func resourceRouteServiceBindingResourceV0() *schema.Resource {
	return &schema.Resource{
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

func resourceRouteServiceBindingImport(ctx context.Context, d *schema.ResourceData, meta any) (res []*schema.ResourceData, err error) {
	return ImportReadContext(resourceRouteServiceBindingRead)(ctx, d, meta)
}

func resourceRouteServiceBindingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	options := client.NewServiceRouteBindingListOptions()
	options.ServiceInstanceGUIDs = client.Filter{Values: []string{serviceID}}
	options.RouteGUIDs = client.Filter{Values: []string{routeID}}

	routeBinding, err := session.ClientGo.ServiceRouteBindings.Single(context.Background(), options)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(routeBinding.GUID)
	return nil
}

func resourceRouteServiceBindingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	session := meta.(*managers.Session)

	routeServiceBinding, err := session.ClientGo.ServiceRouteBindings.Get(context.Background(), d.Id())

	if err != nil {
		if strings.Contains(err.Error(), "CF-ResourceNotFound") {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Error when reading routeServiceBinding with id '%s': %s", d.Id(), err)
	}

	d.Set("service_instance", routeServiceBinding.Relationships.ServiceInstance.Data.GUID)
	d.Set("route", routeServiceBinding.Relationships.Route.Data.GUID)

	return nil
}

func resourceRouteServiceBindingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	session := meta.(*managers.Session)

	jobGUID, err := session.ClientGo.ServiceRouteBindings.Delete(context.Background(), d.Id())

	if err != nil {
		return diag.FromErr(err)
	}
	if jobGUID != "" {
		err = session.ClientGo.Jobs.PollComplete(context.Background(), jobGUID, nil)
	}
	return diag.FromErr(err)
}

func upgradeStateRouteServiceBindingStateV0toV1ChangeID(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	session := meta.(*managers.Session)

	if len(rawState) == 0 {
		log.Println("[DEBUG] Empty RouteServiceBinding; nothing to migrate.")
		return rawState, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", rawState)
	options := client.NewServiceRouteBindingListOptions()
	options.ServiceInstanceGUIDs = client.Filter{Values: []string{rawState["service_instance"].(string)}}
	options.RouteGUIDs = client.Filter{Values: []string{rawState["route"].(string)}}

	routeBinding, err := session.ClientGo.ServiceRouteBindings.Single(context.Background(), options)

	if err != nil {
		if strings.Contains(err.Error(), "expected exactly 1 result, but got less or more than 1") {
			rawState["id"] = ""
			return rawState, nil
		}
		log.Println("[DEBUG] Failed to migrate RouteServiceBinding id: did not find the route service binding.")
		return rawState, err
	}

	rawState["id"] = routeBinding.GUID

	log.Printf("[DEBUG] Attributes after migration: %#v", rawState)

	return rawState, nil
}
