package cloudfoundry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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
			// Some instances takes more time for creation
			// This a custom timeout flag to give service more time for creation in minutes
			"timeout_in_minutes": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  5,
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
				"params", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
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

	// Some instances takes more time for creation
	// This a custom timeout_in_minutes flag to give service more time for creation in minutes
	poll_timeout_in_minutes := d.Get("timeout_in_minutes").(int)
	for _, v := range d.Get("tags").([]interface{}) {
		tags = append(tags, v.(string))
	}
	tags_format := types.OptionalStringSlice{
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
	params_format := types.OptionalObject{
		IsSet: true,
		Value: params,
	}
	log.Printf("params_format : %+v", params_format)
	serviceInstance := resources.ServiceInstance{}
	serviceInstance.Type = ManagedServiceInstance
	serviceInstance.Name = name
	serviceInstance.SpaceGUID = space
	serviceInstance.ServicePlanGUID = servicePlan
	serviceInstance.Tags = tags_format
	serviceInstance.Parameters = params_format

	log.Printf("SI : %+v", serviceInstance)
	jobURL, _, err := session.ClientV3.CreateServiceInstance(serviceInstance)
	log.Printf("Job URL : %+v", jobURL)
	if err != nil {
		return diag.FromErr(err)
	}

	// Poll the state of the async job
	err = common.PollingWithTimeout(func() (bool, error) {
		job, _, err := session.ClientV3.GetJob(jobURL)
		log.Printf("Job URL Status output: %+v", job)
		if err != nil {
			return true, err
		}

		// Stop polling and return error if job failed
		if job.State == constant.JobFailed {
			log.Printf("Failed")
			return true, fmt.Errorf(
				"Service Instance %s failed %s, reason: async job failed",
				name,
				space,
			)
		}
		if job.State == constant.JobComplete {
			si, _, _, err := session.ClientV3.GetServiceInstanceByNameAndSpace(name, space)
			log.Printf("Job completed for Service Instance Creation")
			log.Printf("Service Instance Object : %+v", si)
			if err != nil {
				return true, err
			}

			log.Printf("Service Instance GUID : %+v", si.GUID)
			d.SetId(si.GUID)
			return true, err
		}
		// Last operation initial or inprogress or job not completed, continue polling
		return false, nil
	}, 5*time.Second, time.Duration(poll_timeout_in_minutes)*time.Minute)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	name := d.Get("name").(string)
	space := d.Get("space").(string)

	serviceInstance, _, _, err := session.ClientV3.GetServiceInstanceByNameAndSpace(name, space)
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

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

		d.Set("jsonParameters", params)
	} else {
		d.Set("jsonParameters", nil)
	}

	return nil
}

func resourceServiceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
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

	id = d.Id()
	name = d.Get("name").(string)
	servicePlan := d.Get("service_plan").(string)
	jsonParameters := d.Get("json_params").(string)
	space := d.Get("space").(string)

	// Some instances takes more time for creation
	// This a custom timeout_in_minutes flag to give service more time for creation in minutes
	poll_timeout_in_minutes := d.Get("timeout_in_minutes").(int)
	if len(jsonParameters) > 0 {
		err := json.Unmarshal([]byte(jsonParameters), &params)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	tags = make([]string, 0)
	log.Printf("Tags : %+v", tags)

	for _, v := range d.Get("tags").([]interface{}) {
		tags = append(tags, v.(string))
	}

	tags_format := types.OptionalStringSlice{
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
	params_format := types.OptionalObject{
		IsSet: true,
		Value: params,
	}
	log.Printf("Tags Format : %+v", tags_format)
	log.Printf("Executing Update Instance")
	jobURL, _, err := session.ClientV3.UpdateServiceInstance(id, resources.ServiceInstance{
		Name:            name,
		ServicePlanGUID: servicePlan,
		Parameters:      params_format,
		Tags:            tags_format,
	})
	log.Printf("Service Instance Object Job URL : %+v", jobURL)
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
				"Instance %s failed %s, reason: async job failed",
				name,
				space,
			)
		}
		/*
			query := ccv3.Query{
				Key:    ccv3.GUIDFilter,
				Values: []string{d.Id()},
			}*/
		// Check the state if job completed
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
	}, 5*time.Second, time.Duration(poll_timeout_in_minutes)*time.Minute)
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
	serviceInstanceDeleteTimeout := d.Get("timeout_in_minutes").(int)

	// Poll the state of the async job
	err = common.PollingWithTimeout(func() (bool, error) {
		job, _, err := session.ClientV3.GetJob(jobURL)
		if err != nil {
			return true, err
		}

		// Stop polling and return error if job failed
		if job.State == constant.JobFailed {
			return true, fmt.Errorf(
				"Instance %s failed %s, reason: async job failed",
				name,
				space,
			)
		}
		/*
			query := ccv3.Query{
				Key:    ccv3.GUIDFilter,
				Values: []string{d.Id()},
			}*/
		// Check the state if job completed
		if job.State == constant.JobComplete {
			_, _, _, err := session.ClientV3.GetServiceInstanceByNameAndSpace(name, space)
			if err != nil && !IsErrNotFound(err) {
				return true, err
			}
			return true, nil
		}

		// Last operation initial or inprogress or job not completed, continue polling
		return false, nil
	}, 5*time.Second, time.Duration(serviceInstanceDeleteTimeout)*time.Minute)

	return diag.FromErr(err)
}

func resourceServiceInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	space := d.Get("space").(string)
	serviceInstance, _, _, err := session.ClientV3.GetServiceInstanceByNameAndSpace(name, space)

	if err != nil {
		return nil, err
	}

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
