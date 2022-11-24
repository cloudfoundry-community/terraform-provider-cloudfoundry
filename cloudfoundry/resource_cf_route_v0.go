package cloudfoundry

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
)

func patchRouteV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		rawState = map[string]interface{}{}
	}
	// random_port is no more
	if _, ok := rawState["random_port"]; ok {
		delete(rawState, "random_port")
	}
	return rawState, nil
}

func ResourceRouteV0() *schema.Resource {
	return &schema.Resource{
		// This is only used for state migration, so the CRUD
		// callbacks are no longer relevant
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
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"random_port"},
			},
			"random_port": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"port"},
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
