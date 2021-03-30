package cloudfoundry

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
)

func schemaOldApp() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"service_binding": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"service_instance": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"params": &schema.Schema{
						Type:     schema.TypeMap,
						Optional: true,
					},
					"binding_id": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"route": &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"default_route": &schema.Schema{
						Type: schema.TypeString,
					},
					"default_route_mapping_id": &schema.Schema{
						Type: schema.TypeString,
					},
					"stage_route": &schema.Schema{
						Type: schema.TypeString,
					},
					"stage_route_mapping_id": &schema.Schema{
						Type: schema.TypeString,
					},
					"live_route": &schema.Schema{
						Type: schema.TypeString,
					},
					"live_route_mapping_id": &schema.Schema{
						Type: schema.TypeString,
					},
					"validation_script": &schema.Schema{
						Type: schema.TypeString,
					},
					"version": &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
		"routes": &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Set: func(v interface{}) int {
				elem := v.(map[string]interface{})
				var target string
				if v, ok := elem["route"]; ok {
					target = v.(string)
				} else if v, ok := elem["app"]; ok {
					target = v.(string)
				}
				return hashcode.String(target)
			},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"route": &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.NoZeroValues,
					},
					"port": &schema.Schema{
						Type: schema.TypeInt,
					},
					"mapping_id": &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func resourceAppMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found app Record State v0; migrating from v0 to v3")
		return migrateAppStateV2toV3(is, meta)
	case 2:
		log.Println("[INFO] Found app Record State v2; migrating from v2 to v3")
		return migrateAppStateV2toV3(is, meta)
	case 3:
		log.Println("[INFO] Found app Record State v3; migrating from v3 to v4")
		return migrateAppStateV3toV4(is, meta)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateAppStateV3toV4(is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	oldSchema := map[string]*schema.Schema{
		"service_binding": &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Set: func(v interface{}) int {
				elem := v.(map[string]interface{})
				return hashcode.String(elem["service_instance"].(string))
			},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"service_instance": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"params": &schema.Schema{
						Type:     schema.TypeMap,
						Optional: true,
					},
					"binding_id": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}

	reader := &schema.MapFieldReader{
		Schema: oldSchema,
		Map:    schema.BasicMapReader(is.Attributes),
	}
	result, err := reader.ReadField([]string{"service_binding"})
	if err != nil {
		return is, err
	}
	bindings := make([]map[string]interface{}, 0)
	result, err = reader.ReadField([]string{"service_binding"})
	if err == nil && result.Exists {
		oldBindings := getListOfStructs(result.Value)
		for _, b := range oldBindings {
			bindings = append(bindings, map[string]interface{}{
				"service_instance": b["service_instance"],
				"params":           b["params"],
				"params_json":      b["params_json"],
			})
		}
	}

	is.Attributes = cleanByKeyAttribute("service_binding", is.Attributes)

	writer := schema.MapFieldWriter{
		Schema: resourceApp().Schema,
	}

	err = writer.WriteField([]string{"service_binding"}, bindings)
	if err != nil {
		return is, err
	}
	attr := is.Attributes
	for k, v := range writer.Map() {
		attr[k] = v
	}
	is.Attributes = attr
	return is, nil
}

func migrateAppStateV2toV3(is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	reader := &schema.MapFieldReader{
		Schema: schemaOldApp(),
		Map:    schema.BasicMapReader(is.Attributes),
	}

	portReader := &schema.MapFieldReader{
		Schema: map[string]*schema.Schema{
			"ports": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{Type: schema.TypeInt},
				Set:  resourceIntegerSet,
			},
		},
		Map: schema.BasicMapReader(is.Attributes),
	}
	result, err := portReader.ReadField([]string{"ports"})
	if err != nil {
		return is, err
	}

	ports := make([]int, 0)
	if result.Exists {
		for _, p := range result.Value.(*schema.Set).List() {
			ports = append(ports, p.(int))
		}
	}

	routes := make([]map[string]interface{}, 0)

	result, err = reader.ReadField([]string{"route"})
	if err != nil {
		return is, err
	}
	if result.Exists {
		oldRoute := getListOfStructs(result.Value)
		if len(oldRoute) > 0 && oldRoute[0]["default_route_mapping_id"].(string) != "" {
			routes = append(routes, map[string]interface{}{
				"route": oldRoute[0]["default_route_mapping_id"].(string),
				"port":  ports[0],
			})
		}
	}

	result, err = reader.ReadField([]string{"routes"})
	if err == nil && result.Exists {
		oldRoutes := getListOfStructs(result.Value)
		for _, r := range oldRoutes {
			if port, ok := r["port"]; ok && port.(int) > 0 {
				routes = append(routes, map[string]interface{}{
					"route": r["route"],
					"port":  port,
				})
				continue
			}
			for _, port := range ports {
				routes = append(routes, map[string]interface{}{
					"route": r["route"],
					"port":  port,
				})
			}
		}
	}

	bindings := make([]map[string]interface{}, 0)
	result, err = reader.ReadField([]string{"service_binding"})
	if err == nil && result.Exists {
		oldBindings := getListOfStructs(result.Value)
		for _, b := range oldBindings {
			bindings = append(bindings, map[string]interface{}{
				"service_instance": b["service_instance"],
				"params":           b["params"],
				"params_json":      "",
			})
		}
	}

	is.Attributes = cleanByKeyAttribute("service_binding", is.Attributes)
	is.Attributes = cleanByKeyAttribute("routes", is.Attributes)
	is.Attributes = cleanByKeyAttribute("route", is.Attributes)

	writer := schema.MapFieldWriter{
		Schema: resourceApp().Schema,
	}

	err = writer.WriteField([]string{"service_binding"}, bindings)
	if err != nil {
		return is, err
	}

	err = writer.WriteField([]string{"routes"}, routes)
	if err != nil {
		return is, err
	}

	attr := is.Attributes
	for k, v := range writer.Map() {
		attr[k] = v
	}
	is.Attributes = attr
	return migrateBitsStateV2toV3(is, meta)
}
