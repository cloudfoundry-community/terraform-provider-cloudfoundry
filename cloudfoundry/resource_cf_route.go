package cloudfoundry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func ResourceRoute() *schema.Resource {

	return &schema.Resource{
		SchemaVersion: 1,
		CreateContext: resourceRouteCreate,
		ReadContext:   resourceRouteRead,
		UpdateContext: resourceRouteUpdate,
		DeleteContext: resourceRouteDelete,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    ResourceRouteV0().CoreConfigSchema().ImpliedType(),
				Upgrade: patchRouteV0,
				Version: 0,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: ImportReadContext(resourceRouteRead),
		},

		Schema: map[string]*schema.Schema{

			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"space": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"target": {
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
						"app": {
							Type:     schema.TypeString,
							Required: true,
						},
						"port": {
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

func setRouteStateV3(session *managers.Session, route resources.Route, d *schema.ResourceData) (err error) {
	_ = d.Set("domain", route.DomainGUID)
	_ = d.Set("space", route.SpaceGUID)
	_ = d.Set("hostname", route.Host)

	if route.Port != 0 {
		_ = d.Set("port", route.Port)
	}

	_ = d.Set("path", route.Path)

	// In v3 shared domains and private domains are managed by the same endpoint, differenciating on whether
	// a relationship with an org is set
	domain, _, err := session.ClientV3.GetDomain(route.DomainGUID)
	if err != nil || domain.GUID == "" {
		return err
	}

	port := ""
	if route.Port > 0 && domain.RouterGroup != "" {
		port = fmt.Sprintf(":%d", route.Port)
	}
	endpoint := fmt.Sprintf("%s.%s%s", route.Host, domain.Name, port)
	if route.Path != "" {
		endpoint += "/" + route.Path
	}
	d.Set("endpoint", endpoint)
	return nil
}

func addRouteDestinationV3(id string, add []map[string]interface{}, session *managers.Session) ([]map[string]interface{}, error) {
	targets := make([]map[string]interface{}, 0)
	for _, t := range add {
		appID := t["app"].(string)

		_, err := session.ClientV3.MapRoute(id, appID)
		if err != nil {
			return targets, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func removeRouteDestinationV3(id string, delete []map[string]interface{}, session *managers.Session) error {
	mappings, _, err := session.ClientV3.GetRouteDestinations(id)
	if err != nil {
		return err
	}

	for _, t := range delete {
		appID := t["app"].(string)
		for _, mapping := range mappings {
			// if 0 it mean app port has been set to null which means it takes the first port found in app port definition
			if mapping.App.GUID == appID {
				_, err := session.ClientV3.UnmapRoute(id, mapping.GUID)
				if err != nil && !IsErrNotFound(err) {
					return err
				}
			}
		}
	}
	return nil
}

func resourceRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	var route = resources.Route{}

	// Call create route API
	operation := func() error {
		var err error
		route, _, err = session.ClientV3.CreateRoute(resources.Route{
			DomainGUID: d.Get("domain").(string),
			SpaceGUID:  d.Get("space").(string),
			Host:       d.Get("hostname").(string),
			Path:       d.Get("path").(string),
			Port:       d.Get("port").(int),
		})

		if v, ok := d.GetOk("port"); ok {
			route.Port = v.(int)
		}
		if err != nil {
			if unexpected, ok := err.(ccerror.V3UnexpectedResponseError); ok && unexpected.ResponseCode == http.StatusInternalServerError {
				return err
			}
			return backoff.Permanent(err)
		}
		return nil
	}

	// Retry the API call
	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return diag.FromErr(err)
	}

	// Delete route if an error occurs, defer will always run at the end of the block
	defer func() {
		e := &err
		if *e == nil {
			return
		}
		_, _, err = session.ClientV3.DeleteRoute(route.GUID)
		if err != nil {
			panic(err)
		}
	}()

	// set fields in tfstate, calculate URL field
	if err = setRouteStateV3(session, route, d); err != nil {
		return diag.FromErr(err)
	}

	// Separate call to add destinations
	if v, ok := d.GetOk("target"); ok {
		var t interface{}
		if t, err = addRouteDestinationV3(route.GUID, GetListOfStructs(v.(*schema.Set).List()), session); err != nil {
			return diag.FromErr(err)
		}
		_ = d.Set("target", t)
	}

	d.SetId(route.GUID)
	return diag.FromErr(err)
}

func resourceRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	id := d.Id()

	routes, _, err := session.ClientV3.GetRoutes(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: []string{id},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if len(routes) == 0 {
		d.SetId("")
		return nil
	}
	if len(routes) != 1 {
		return diag.FromErr(fmt.Errorf("Unexpected error reading route (different than 1 match)"))
	}

	route := routes[0]

	if err = setRouteStateV3(session, route, d); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("target"); !ok && !IsImportState(d) {
		return nil
	}
	mappingsTf := make([]map[string]interface{}, 0)
	tfTargets := d.Get("target").(*schema.Set).List()
	mappings, _, err := session.ClientV3.GetRouteDestinations(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if IsImportState(d) {
		for _, mapping := range mappings {
			mappingsTf = append(mappingsTf, map[string]interface{}{
				"app": mapping.App.GUID,
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
			if mapping.App.GUID == tmpT["app"] {
				inside = true
				tmpT["app"] = mapping.App.GUID
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

func resourceRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	if d.HasChange("domain") || d.HasChange("space") || d.HasChange("hostname") || d.HasChange("target") {
		// Delete and recreate
		var route = resources.Route{}

		if targets, ok := d.GetOk("target"); ok {
			err := removeRouteDestinationV3(d.Id(), GetListOfStructs(targets.(*schema.Set).List()), session)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		_, _, err := session.ClientV3.DeleteRoute(d.Id())
		if err != nil && !IsErrNotFound(err) {
			return diag.FromErr(err)
		}

		operation := func() error {
			var err error
			route, _, err = session.ClientV3.CreateRoute(resources.Route{
				DomainGUID: d.Get("domain").(string),
				SpaceGUID:  d.Get("space").(string),
				Host:       d.Get("hostname").(string),
				Path:       d.Get("path").(string),
				Port:       d.Get("port").(int),
			})
			if err != nil {
				if unexpected, ok := err.(ccerror.V3UnexpectedResponseError); ok && unexpected.ResponseCode == http.StatusInternalServerError {
					return err
				}
				return backoff.Permanent(err)
			}
			return nil
		}

		// Retry the API call
		err = backoff.Retry(operation, backoff.NewExponentialBackOff())
		if err != nil {
			return diag.FromErr(err)
		}

		// Delete route if an error occurs, defer will always run at the end of the block
		defer func() {
			e := &err
			if *e == nil {
				return
			}
			_, _, err = session.ClientV3.DeleteRoute(d.Id())
			if err != nil {
				panic(err)
			}
		}()

		// set fields in tfstate, calculate URL field
		if err = setRouteStateV3(session, route, d); err != nil {
			return diag.FromErr(err)
		}

		// Separate call to add destinations
		if v, ok := d.GetOk("target"); ok {
			var t interface{}
			if t, err = addRouteDestinationV3(route.GUID, GetListOfStructs(v.(*schema.Set).List()), session); err != nil {
				return diag.FromErr(err)
			}
			_ = d.Set("target", t)
		}

		d.SetId(route.GUID)
		return diag.FromErr(err)
	}
	return nil
}

func resourceRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	if targets, ok := d.GetOk("target"); ok {
		err := removeRouteDestinationV3(d.Id(), GetListOfStructs(targets.(*schema.Set).List()), session)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	jobURL, _, err := session.ClientV3.DeleteRoute(d.Id())
	if err != nil && !IsErrNotFound(err) {
		return diag.FromErr(err)
	}

	err = PollAsyncJob(PollingConfig{
		Session: session,
		JobURL:  jobURL,
	})
	return diag.FromErr(err)
}
