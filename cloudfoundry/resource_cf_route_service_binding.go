package cloudfoundry

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
		MigrateState:  resourceRouteServiceBindingMigrateState,
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

func resourceRouteServiceBindingMigrateState(v int, inst *terraform.InstanceState, meta any) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found Route Service Binding State v0; migrating to v1: change ID from routeID:serviceID to routeServiceBindingID.")
		return migrateRouteServiceBindingStateV0toV1ChangeID(inst, meta)
	default:
		return inst, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateRouteServiceBindingStateV0toV1ChangeID(inst *terraform.InstanceState, meta any) (*terraform.InstanceState, error) {
	session := meta.(*managers.Session)

	if inst.Empty() {
		log.Println("[DEBUG] Empty RouteServiceBinding; nothing to migrate.")
		return inst, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", inst.Attributes)
	options := client.NewServiceRouteBindingListOptions()
	options.ServiceInstanceGUIDs = client.Filter{Values: []string{inst.Attributes["service_instance"]}}
	options.RouteGUIDs = client.Filter{Values: []string{inst.Attributes["route"]}}

	routeBinding, err := session.ClientGo.ServiceRouteBindings.Single(context.Background(), options)

	if err != nil {
		log.Println("[DEBUG] Failed to migrate RouteServiceBinding id: did not find the route service binding.")
		return inst, err
	}

	inst.Attributes["id"] = routeBinding.GUID

	log.Printf("[DEBUG] Attributes after migration: %#v", inst.Attributes)

	return inst, nil
}
