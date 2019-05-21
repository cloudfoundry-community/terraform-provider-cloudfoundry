package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/types"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceRoute() *schema.Resource {

	return &schema.Resource{

		Create: resourceRouteCreate,
		Read:   resourceRouteRead,
		Update: resourceRouteUpdate,
		Delete: resourceRouteDelete,

		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
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
				ConflictsWith: []string{"random_port"},
			},
			"random_port": &schema.Schema{
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"port"},
			},
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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

func resourceRouteCreate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	port := types.NullInt{}
	if v, ok := d.GetOk("port"); ok {
		port.Value = v.(int)
		port.IsSet = true
	}

	route, _, err := session.ClientV2.CreateRoute(ccv2.Route{
		DomainGUID: d.Get("domain").(string),
		SpaceGUID:  d.Get("space").(string),
		Host:       d.Get("host").(string),
		Path:       d.Get("path").(string),
		Port:       port,
	}, d.Get("random_port").(bool))
	if err != nil {
		return err
	}
	// Delete route if an error occurs
	defer func() {
		e := &err
		if *e == nil {
			return
		}
		_, err = session.ClientV2.DeleteRoute(route.GUID)
		if err != nil {
			panic(err)
		}
	}()

	if err = setRouteArguments(session, route, d); err != nil {
		return err
	}

	if v, ok := d.GetOk("target"); ok {
		var t interface{}
		if t, err = addTargets(route.GUID, getListOfStructs(v.(*schema.Set).List()), session); err != nil {
			return err
		}
		d.Set("target", t)
	}

	d.SetId(route.GUID)
	return err
}

func resourceRouteRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	id := d.Id()

	route, _, err := session.ClientV2.GetRoute(id)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			d.SetId("")
			err = nil
		}
		return err
	}
	if err = setRouteArguments(session, route, d); err != nil {
		return err
	}

	if _, ok := d.GetOk("target"); !ok && !IsImportState(d) {
		return nil
	}
	mappingsTf := make([]map[string]interface{}, 0)
	tfTargets := d.Get("target").(*schema.Set).List()
	mappings, _, err := session.ClientV2.GetRouteMappings(ccv2.FilterEqual(constant.RouteGUIDFilter, d.Id()))
	if err != nil {
		return err
	}
	if IsImportState(d) {
		for _, mapping := range mappings {
			mappingsTf = append(mappingsTf, map[string]interface{}{
				"app":        mapping.AppGUID,
				"mapping_id": mapping.GUID,
				"port":       mapping.AppPort,
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
			if mapping.GUID == tmpT["mapping_id"] {
				inside = true
				tmpT["port"] = mapping.AppPort
				tmpT["app"] = mapping.AppGUID
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

func resourceRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	port := types.NullInt{}
	if v, ok := d.GetOk("port"); ok {
		port.Value = v.(int)
		port.IsSet = true
	}
	if d.HasChange("domain") || d.HasChange("space") || d.HasChange("hostname") {
		route, _, err := session.ClientV2.UpdateRoute(ccv2.Route{
			GUID:       d.Id(),
			DomainGUID: d.Get("domain").(string),
			SpaceGUID:  d.Get("space").(string),
			Host:       d.Get("host").(string),
			Path:       d.Get("path").(string),
			Port:       port,
		})
		if err != nil {
			return err
		}
		err = setRouteArguments(session, route, d)
		if err != nil {
			return err
		}
	}

	if d.HasChange("target") {
		old, new := d.GetChange("target")
		remove, _ := getListMapChanges(old, new, func(source, item map[string]interface{}) bool {
			return source["app"] == item["app"] && source["port"] == item["port"]
		})
		err := removeTargets(remove, session)
		if err != nil {
			return err
		}

		t, err := addTargets(d.Id(), getListOfStructs(d.Get("target").(*schema.Set).List()), session)
		if err != nil {
			return err
		}
		d.Set("target", t)
	}
	return nil
}

func resourceRouteDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	if targets, ok := d.GetOk("target"); ok {
		err := removeTargets(getListOfStructs(targets.(*schema.Set).List()), session)
		if err != nil {
			return err
		}
	}
	_, err := session.ClientV2.DeleteRoute(d.Id())
	return err
}

func setRouteArguments(session *managers.Session, route ccv2.Route, d *schema.ResourceData) (err error) {

	d.Set("domain", route.DomainGUID)
	d.Set("space", route.SpaceGUID)
	d.Set("hostname", route.Host)
	if route.Port.IsSet {
		d.Set("port", route.Port.Value)
	}
	d.Set("path", route.Path)

	domain, _, err := session.ClientV2.GetSharedDomain(route.DomainGUID)
	if err != nil || domain.GUID == "" {
		domain, _, err = session.ClientV2.GetPrivateDomain(route.DomainGUID)
		if err != nil {
			return err
		}
	}
	port := ""
	if route.Port.IsSet && route.Port.Value > 0 && domain.RouterGroupGUID != "" {
		port = fmt.Sprintf(":%d", route.Port.Value)
	}
	endpoint := fmt.Sprintf("%s.%s%s", route.Host, domain.Name, port)
	if route.Path != "" {
		endpoint += "/" + route.Path
	}
	d.Set("endpoint", endpoint)
	return nil
}

func addTargets(id string, add []map[string]interface{}, session *managers.Session) ([]map[string]interface{}, error) {
	targets := make([]map[string]interface{}, 0)

	for _, t := range add {
		if t["mapping_id"].(string) != "" {
			continue
		}
		appID := t["app"].(string)
		port := 8080
		if v, ok := t["port"]; ok {
			port = v.(int)
		}
		mapping, _, err := session.ClientV2.CreateRouteMapping(appID, id, port)
		if err != nil {
			return targets, err
		}
		t["mapping_id"] = mapping.GUID
		targets = append(targets, t)
	}
	return targets, nil
}

func removeTargets(delete []map[string]interface{}, session *managers.Session) error {

	for _, t := range delete {
		mappingID := t["mapping_id"].(string)
		if mappingID == "" {
			continue
		}
		if len(mappingID) > 0 {
			_, err := session.ClientV2.DeleteRouteMapping(mappingID)
			if err != nil {
				if IsErrNotFound(err) {
					continue
				}
				return err
			}
		}
	}
	return nil
}
