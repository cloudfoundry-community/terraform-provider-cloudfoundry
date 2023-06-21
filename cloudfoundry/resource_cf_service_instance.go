package cloudfoundry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

const ManagedServiceInstance = "managed"

func resourceServiceInstance() *schema.Resource {

	return &schema.Resource{

		CreateContext: resourceServiceInstanceCreate,
		ReadContext:   resourceServiceInstanceRead,
		UpdateContext: resourceServiceInstanceUpdate,
		DeleteContext: resourceServiceInstanceDelete,

		SchemaVersion: 1,

		MigrateState: resourceServiceInstanceMigrateState,

		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceInstanceImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"service_plan": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"json_params": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringIsJSON,
			},
			"replace_on_service_plan_change": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"replace_on_params_change": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"recursive_delete": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIf(
				"service_plan", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
					if ok := d.Get("replace_on_service_plan_change").(bool); ok {
						return true
					}
					return false
				}),
			customdiff.ForceNewIf(
				"json_params", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
					if ok := d.Get("replace_on_params_change").(bool); ok {
						return true
					}
					return false
				},
			),
		),
	}
}

func resourceServiceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	servicePlan := d.Get("service_plan").(string)
	space := d.Get("space").(string)
	jsonParameters := d.Get("json_params").(string)
	tags := make([]string, 0)

	for _, v := range d.Get("tags").([]interface{}) {
		tags = append(tags, v.(string))
	}
	tagsFormatted := types.OptionalStringSlice{
		IsSet: true,
		Value: tags,
	}

	params := make(map[string]interface{})
	if len(jsonParameters) > 0 {
		err := json.Unmarshal([]byte(jsonParameters), &params)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	paramsFormatted := types.OptionalObject{
		IsSet: true,
		Value: params,
	}

	serviceInstance := resources.ServiceInstance{}
	serviceInstance.Type = ManagedServiceInstance
	serviceInstance.Name = name
	serviceInstance.SpaceGUID = space
	serviceInstance.ServicePlanGUID = servicePlan
	serviceInstance.Tags = tagsFormatted
	serviceInstance.Parameters = paramsFormatted

	jobURL, _, err := session.ClientV3.CreateServiceInstance(serviceInstance)
	if err != nil {
		return diag.FromErr(err)
	}

	// Poll the state of the async job
	err = common.PollingWithTimeout(func() (bool, error) {
		job, _, err := session.ClientV3.GetJob(jobURL)
		if err != nil {
			return true, err
		}

		// Stop polling and return error if job failed
		if job.State == constant.JobFailed {
			return true, fmt.Errorf(
				"Service Instance %s failed %s, reason: %+v",
				name,
				space,
				job.Errors(),
			)
		}
		// If job completed, check if the service instance is created
		if job.State == constant.JobComplete {
			si, _, _, err := session.ClientV3.GetServiceInstanceByNameAndSpace(name, space)
			if err != nil {
				return true, err
			}
			d.SetId(si.GUID)
			return true, nil
		}
		// Last operation initial or inprogress or job not completed, continue polling
		return false, nil
	}, 5*time.Second, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	serviceInstances, _, _, err := session.ClientV3.GetServiceInstances(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: []string{d.Id()},
	})

	if len(serviceInstances) > 1 {
		err = fmt.Errorf("GUID filter for service instance %s returned more than one result", d.Id())
		return diag.FromErr(err)
	}

	if len(serviceInstances) == 0 {
		d.SetId("")
		return nil
	}

	serviceInstance := serviceInstances[0]

	d.Set("name", serviceInstance.Name)
	d.Set("service_plan", serviceInstance.ServicePlanGUID)
	d.Set("space", serviceInstance.SpaceGUID)

	if serviceInstance.Tags.IsSet {
		tags := make([]interface{}, len(serviceInstance.Tags.Value))
		for i, v := range serviceInstance.Tags.Value {
			tags[i] = v
		}
		d.Set("tags", tags)
	} else {
		d.Set("tags", nil)
	}
	if serviceInstance.Parameters.IsSet {
		//params := make(map[string]interface{})
		params, err := json.Marshal(serviceInstance.Parameters.Value)
		if err != nil {
			return diag.FromErr(err)
		}

		d.Set("json_params", params)
	}
	// Keep state as-is if the cloudcontroller does not return any Parameters (99% of the time)

	return nil
}

func resourceServiceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	// Nothing to be done
	if !isServiceInstanceUpdateRequired(d) {
		return nil
	}

	// Enable partial state mode
	// We need to explicitly set state updates ourselves or
	// tell terraform when a state change is applied and thus okay to persist
	// In particular this is necessary for params since we cannot query CF for
	// the current value of this field
	d.Partial(true)

	var (
		id, name string
		tags     []string
		params   map[string]interface{}
	)

	name = d.Get("name").(string)
	id = d.Id()
	jsonParameters := d.Get("json_params").(string)
	space := d.Get("space").(string)

	if len(jsonParameters) > 0 {
		err := json.Unmarshal([]byte(jsonParameters), &params)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	tags = make([]string, 0)
	// log.Printf("Tags : %+v", tags)

	for _, v := range d.Get("tags").([]interface{}) {
		tags = append(tags, v.(string))
	}

	tagsFormatted := types.OptionalStringSlice{
		IsSet: true,
		Value: tags,
	}
	params = make(map[string]interface{})
	if len(jsonParameters) > 0 {
		err := json.Unmarshal([]byte(jsonParameters), &params)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	paramsFormatted := types.OptionalObject{
		IsSet: true,
		Value: params,
	}

	serviceInstanceUpdate := resources.ServiceInstance{
		Name:       name,
		Parameters: paramsFormatted,
		Tags:       tagsFormatted,
	}
	// Some services don't support changing service plan, so we only add it to request body only if changed by user
	if d.HasChange("service_plan") {
		serviceInstanceUpdate.ServicePlanGUID = d.Get("service_plan").(string)
	}

	jobURL, _, err := session.ClientV3.UpdateServiceInstance(id, serviceInstanceUpdate)
	// log.Printf("Service Instance Object Job URL : %+v", JobURL)
	if err != nil {
		return diag.FromErr(err)
	}

	// Poll the state of the async job
	err = common.PollingWithTimeout(func() (bool, error) {
		job, _, err := session.ClientV3.GetJob(jobURL)
		if err != nil {
			return true, err
		}

		// Stop polling and return error if job failed
		if job.State == constant.JobFailed {
			return true, fmt.Errorf(
				"Instance %s failed %s, reason: %+v",
				name,
				space,
				job.Errors(),
			)
		}
		// If job completed, check if the service instance exists
		if job.State == constant.JobComplete {
			si, _, _, err := session.ClientV3.GetServiceInstanceByNameAndSpace(name, space)
			if err != nil {
				return true, err
			}
			d.SetId(si.GUID)
			return true, nil
		}

		// Last operation initial or inprogress or job not completed, continue polling
		return false, nil
	}, 5*time.Second, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}
	// We succeeded, disable partial mode
	d.Partial(false)
	return nil
}

func resourceServiceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	id := d.Id()

	jobURL, _, err := session.ClientV3.DeleteServiceInstance(id)

	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	space := d.Get("space").(string)

	// Poll the state of the async job
	err = common.PollingWithTimeout(func() (bool, error) {
		job, _, err := session.ClientV3.GetJob(jobURL)
		if err != nil {
			return true, err
		}

		// Stop polling and return error if job failed
		if job.State == constant.JobFailed {
			return true, fmt.Errorf(
				"Instance %s failed %s, reason: %+v",
				name,
				space,
				job.Errors(),
			)
		}

		if job.State == constant.JobComplete {
			_, _, _, err := session.ClientV3.GetServiceInstances(ccv3.Query{
				Key:    ccv3.GUIDFilter,
				Values: []string{id},
			})
			if err != nil && !IsErrNotFound(err) {
				return true, err
			}
			return true, nil
		}
		// Last operation initial or inprogress or job not completed, continue polling
		return false, nil
	}, 5*time.Second, d.Timeout(schema.TimeoutDelete))

	return diag.FromErr(err)
}

func resourceServiceInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	session := meta.(*managers.Session)

	log.Printf("!!!! Importing state : %+v", d)
	GUID := d.Id()
	serviceInstances, _, _, err := session.ClientV3.GetServiceInstances(ccv3.Query{
		Key:    ccv3.GUIDFilter,
		Values: []string{GUID},
	})

	log.Printf("!!!! Found service instance : %+v", serviceInstances)

	if err != nil {
		return nil, err
	}

	if len(serviceInstances) == 0 {
		return nil, fmt.Errorf("Service instance with guid: %s not found", GUID)
	}

	serviceInstance := serviceInstances[0]

	d.Set("name", serviceInstance.Name)
	d.Set("service_plan", serviceInstance.ServicePlanGUID)
	d.Set("space", serviceInstance.SpaceGUID)
	if serviceInstance.Tags.IsSet {
		tags := make([]interface{}, len(serviceInstance.Tags.Value))
		for i, v := range serviceInstance.Tags.Value {
			tags[i] = v
		}
		d.Set("tags", tags)
	} else {
		d.Set("tags", nil)
	}

	d.Set("replace_on_service_plan_change", false)
	d.Set("replace_on_params_change", false)

	return ImportReadContext(resourceServiceInstanceRead)(ctx, d, meta)
}

func isServiceInstanceUpdateRequired(d ResourceChanger) bool {
	return d.HasChange("name") || d.HasChange("service_plan") || d.HasChange("json_params") || d.HasChange("tags")
}
