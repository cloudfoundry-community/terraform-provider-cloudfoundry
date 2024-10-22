package cloudfoundry

import (
	"context"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
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
	return ImportReadContext(resourceRouteServiceBindingRead)(ctx, d, meta)
}

func findRouteServiceBinding(session *managers.Session, query ...ccv3.Query) (*resources.RouteBinding, diag.Diagnostics) {
	var diags diag.Diagnostics
	routeBindings, _, warnings, err := session.ClientV3.GetRouteBindings(query...)

	if len(warnings) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "API warnings reading Route Service Binding Resource",
			Detail:   strings.Join(warnings, "\n"),
		})
	}
	if err != nil {
		return nil, append(diags, diag.FromErr(err)...)
	}
	if len(routeBindings) == 0 {
		return nil, append(diags, diag.Errorf("Route Service Binding not found")...)
	}
	if len(routeBindings) > 1 {
		return nil, append(diags, diag.Errorf("Something bad happened: there are multiple similar route service bindings")...)
	}

	return &routeBindings[0], diags
}

func resourceRouteServiceBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	var diags diag.Diagnostics

	var data map[string]interface{}

	serviceID := d.Get("service_instance").(string)
	routeID := d.Get("route").(string)
	params, okParams := d.GetOk("json_params")

	if okParams {
		if err := json.Unmarshal([]byte(params.(string)), &data); err != nil {
			return diag.FromErr(err)
		}
	}

	jobURL, warnings, err := session.ClientV3.CreateRouteBinding(resources.RouteBinding{
		ServiceInstanceGUID: serviceID,
		RouteGUID:           routeID,
		Parameters:          types.OptionalObject{IsSet: okParams, Value: data},
	})
	if len(warnings) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "API warnings creating Route Service Binding Resource",
			Detail:   strings.Join(warnings, "\n"),
		})
	}
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}

	if jobURL != "" {
		err = PollAsyncJob(PollingConfig{
			Session: session,
			JobURL:  jobURL,
		})
		if err != nil {
			return append(diags, diag.FromErr(err)...)
		}
	}

	createdRouteBinding, findDiags := findRouteServiceBinding(session,
		ccv3.Query{Key: ccv3.ServiceInstanceGUIDFilter, Values: []string{serviceID}},
		ccv3.Query{Key: ccv3.RouteGUIDFilter, Values: []string{routeID}},
	)

	if findDiags.HasError() {
		return append(diags, findDiags...)
	}

	d.SetId(createdRouteBinding.GUID)
	return append(diags, resourceRouteServiceBindingRead(ctx, d, meta)...)
}

func resourceRouteServiceBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	var diags diag.Diagnostics
	id := d.Id()

	routeBindings, _, warnings, err := session.ClientV3.GetRouteBindings(
		ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{id}},
	)

	if len(warnings) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "API warnings reading Route Service Binding Resource",
			Detail:   strings.Join(warnings, "\n"),
		})
	}
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}
	var foundRouteBinding *resources.RouteBinding = nil
	for _, binding := range routeBindings {
		if binding.GUID == id {
			foundRouteBinding = &binding
			break
		}
	}
	if foundRouteBinding == nil {
		d.SetId("")
		return diags
	}

	d.Set("service_instance", foundRouteBinding.ServiceInstanceGUID)
	d.Set("route", foundRouteBinding.RouteGUID)
	if foundRouteBinding.Parameters.IsSet {
		d.Set("json_params", foundRouteBinding.Parameters.Value)
	}
	return diags
}

func resourceRouteServiceBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	var diags diag.Diagnostics

	jobURL, warnings, err := session.ClientV3.DeleteRouteBinding(d.Id())

	if len(warnings) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "API warnings deleting Route Service Binding Resource",
			Detail:   strings.Join(warnings, "\n"),
		})
	}

	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}

	if jobURL != "" {
		err = PollAsyncJob(PollingConfig{
			Session: session,
			JobURL:  jobURL,
		})
	}

	return append(diags, diag.FromErr(err)...)
}
