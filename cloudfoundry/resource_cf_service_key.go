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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
			Delete: schema.DefaultTimeout(60 * time.Second),
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
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
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

	binding := resources.ServiceCredentialBinding{
		ServiceInstanceGUID: serviceInstance,
		Name:                name,
		Parameters:          types.NewOptionalObject(params),
		Type:                resources.ServiceCredentialBindingType("key"),
	}

	jobURL, _, err := session.ClientV3.CreateServiceCredentialBinding(binding)
	if err != nil {
		return diag.FromErr(err)
	}

	var serviceKey resources.ServiceCredentialBinding

	// Poll the state of the async job
	err = common.PollingWithTimeout(func() (bool, error) {
		job, _, err := session.ClientV3.GetJob(jobURL)
		if err != nil {
			return true, err
		}

		// Stop polling and return error if job failed
		if job.State == constant.JobFailed {
			return true, fmt.Errorf(
				"ServiceKey creation failed for key %s, reason: %+v",
				serviceInstance,
				job.Errors(),
			)
		}

		// Check binding state if job completed
		if job.State == constant.JobComplete {
			createdKeys, _, err := session.ClientV3.GetServiceCredentialBindings(
				ccv3.Query{
					Key:    ccv3.QueryKey("service_instance_guids"),
					Values: []string{binding.ServiceInstanceGUID},
				}, ccv3.Query{
					Key:    ccv3.NameFilter,
					Values: []string{name},
				},
			)
			if err != nil {
				return true, err
			}

			if len(createdKeys) == 0 {
				return false, nil
			}

			if createdKeys[0].LastOperation.State == resources.OperationSucceeded {
				serviceKey = createdKeys[0]
				return true, nil
			}

			if createdKeys[0].LastOperation.State == resources.OperationFailed {
				return true, fmt.Errorf(
					"Service key creation failed for key %s, reason: %+v",
					serviceInstance,
					job.Errors(),
				)
			}
		}

		// Last operation initial or inprogress or job not completed, continue polling
		return false, nil
	}, 5*time.Second, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	credentials, _, err := session.ClientV3.GetServiceCredentialBindingDetails(serviceKey.GUID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("credentials", normalizeMap(credentials.Credentials, make(map[string]interface{}), "", "_"))
	d.SetId(serviceKey.GUID)
	return nil
}

func resourceServiceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	var serviceKeys []resources.ServiceCredentialBinding
	var err error
	if d.Id() != "" {
		serviceKeys, _, err = session.ClientV3.GetServiceCredentialBindings(
			ccv3.Query{
				Key:    ccv3.GUIDFilter,
				Values: []string{d.Id()},
			},
		)
	} else {
		serviceKeys, _, err = session.ClientV3.GetServiceCredentialBindings(
			ccv3.Query{
				Key:    ccv3.QueryKey("service_instance_guids"),
				Values: []string{d.Get("service_instance").(string)},
			}, ccv3.Query{
				Key:    ccv3.NameFilter,
				Values: []string{d.Get("name").(string)},
			},
		)
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if len(serviceKeys) == 0 {
		d.SetId("")
		return nil
	}
	if len(serviceKeys) != 1 {
		return diag.FromErr(fmt.Errorf("Some thing went wrong"))
	}
	d.Set("name", serviceKeys[0].Name)
	d.Set("service_instance", serviceKeys[0].ServiceInstanceGUID)
	d.SetId(serviceKeys[0].GUID)
	serviceKeyDetails, _, err := session.ClientV3.GetServiceCredentialBindingDetails(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("credentials", normalizeMap(serviceKeyDetails.Credentials, make(map[string]interface{}), "", "_"))
	return nil
}

func resourceServiceKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Note : When deleting credential bindings originated from user provided service instances,
	// the delete operation does not require interactions with service brokers,
	// therefore the API will respond synchronously to the delete request.
	session := meta.(*managers.Session)
	jobURL, _, err := session.ClientV3.DeleteServiceCredentialBinding(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if jobURL == "" {
		log.Printf("[INFO] Deleted service credential binding %s for User-Provided service instance, finishing without polling", d.Id())
		return diag.FromErr(err)
	}

	// Polling when deleting service credential binding for a managed service instance
	err = common.PollingWithTimeout(func() (bool, error) {
		job, _, err := session.ClientV3.GetJob(jobURL)
		if err != nil {
			return true, err
		}

		// Stop polling and return error if job failed
		if job.State == constant.JobFailed {
			return true, fmt.Errorf(
				"Service key deletion failed, reason: %+v",
				job.Errors(),
			)
		}

		// Check binding state if job completed
		if job.State == constant.JobComplete {
			return true, nil
		}
		// Last operation initial or inprogress or job not completed, continue polling
		return false, nil
	}, 5*time.Second, d.Timeout(schema.TimeoutDelete))
	return diag.FromErr(err)
}
