package cloudfoundry

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceServiceKey() *schema.Resource {

	return &schema.Resource{

		CreateContext: resourceServiceKeyCreate,
		ReadContext:   resourceServiceKeyRead,
		DeleteContext: resourceServiceKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: ImportReadContext(resourceServiceKeyRead),
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_instance": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"params": &schema.Schema{
				Type:          schema.TypeMap,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"params_json"},
			},
			"params_json": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"params"},
				ValidateFunc:  validation.StringIsJSON,
			},
			"credentials": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceServiceKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	serviceInstance := d.Get("service_instance").(string)
	params := d.Get("params").(map[string]interface{})
	paramJson := d.Get("params_json").(string)
	if len(params) == 0 && paramJson != "" {
		err := json.Unmarshal([]byte(paramJson), &params)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	serviceKey, _, err := session.ClientV2.CreateServiceKey(serviceInstance, name, params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("credentials", normalizeMap(serviceKey.Credentials, make(map[string]interface{}), "", "_"))
	d.SetId(serviceKey.GUID)
	return nil
}

func resourceServiceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	serviceKey, _, err := session.ClientV2.GetServiceKey(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	d.Set("name", serviceKey.Name)
	d.Set("service_instance", serviceKey.ServiceInstanceGUID)
	d.Set("credentials", normalizeMap(serviceKey.Credentials, make(map[string]interface{}), "", "_"))
	return nil
}

func resourceServiceKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	_, err := session.ClientV2.DeleteServiceKey(d.Id())
	return diag.FromErr(err)
}
