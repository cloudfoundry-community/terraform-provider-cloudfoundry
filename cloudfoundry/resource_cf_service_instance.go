package cloudfoundry

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/cf/terminal"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceServiceInstance() *schema.Resource {

	return &schema.Resource{

		Create: resourceServiceInstanceCreate,
		Read:   resourceServiceInstanceRead,
		Update: resourceServiceInstanceUpdate,
		Delete: resourceServiceInstanceDelete,

		Importer: &schema.ResourceImporter{
			State: resourceServiceInstanceImport,
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
			},
			"json_params": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceServiceInstanceCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		id     string
		tags   []string
		params map[string]interface{}
	)
	name := d.Get("name").(string)
	servicePlan := d.Get("service_plan").(string)
	space := d.Get("space").(string)
	jsonParameters := d.Get("json_params").(string)

	for _, v := range d.Get("tags").([]interface{}) {
		tags = append(tags, v.(string))
	}

	if len(jsonParameters) > 0 {
		if err = json.Unmarshal([]byte(jsonParameters), &params); err != nil {
			return err
		}
	}

	sm := session.ServiceManager()

	if id, err = sm.CreateServiceInstance(name, servicePlan, space, params, tags); err != nil {
		return err
	}
	stateConf := &resource.StateChangeConf{
		Pending:      resourceServiceInstancePendingStates,
		Target:       resourceServiceInstanceSucceesStates,
		Refresh:      resourceServiceInstanceStateFunc(id, "create", meta),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		PollInterval: 30 * time.Second,
		Delay:        5 * time.Second,
	}

	// Wait, catching any errors
	if _, err = stateConf.WaitForState(); err != nil {
		return err
	}

	session.Log.DebugMessage("New Service Instance : %# v", id)

	d.SetId(id)

	return nil
}

func resourceServiceInstanceRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("Reading Service Instance : %s", terminal.EntityNameColor(d.Id()))

	sm := session.ServiceManager()
	var serviceInstance cfapi.CCServiceInstance

	serviceInstance, err = sm.ReadServiceInstance(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "status code: 404") {
			d.SetId("")
			err = nil
		}
		return err
	}

	d.Set("name", serviceInstance.Name)
	d.Set("service_plan", serviceInstance.ServicePlanGUID)
	d.Set("space", serviceInstance.SpaceGUID)

	if serviceInstance.Tags != nil {
		tags := make([]interface{}, len(serviceInstance.Tags))
		for i, v := range serviceInstance.Tags {
			tags[i] = v
		}
		d.Set("tags", tags)
	} else {
		d.Set("tags", nil)
	}

	session.Log.DebugMessage("Read Service Instance : %# v", serviceInstance)

	return nil
}

func resourceServiceInstanceUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	sm := session.ServiceManager()

	session.Log.DebugMessage("begin resourceServiceInstanceUpdate")

	var (
		id, name string
		tags     []string
		params   map[string]interface{}
	)

	id = d.Id()
	name = d.Get("name").(string)
	servicePlan := d.Get("service_plan").(string)
	jsonParameters := d.Get("json_params").(string)

	if len(jsonParameters) > 0 {
		if err = json.Unmarshal([]byte(jsonParameters), &params); err != nil {
			return err
		}
	}

	for _, v := range d.Get("tags").([]interface{}) {
		tags = append(tags, v.(string))
	}

	if _, err = sm.UpdateServiceInstance(id, name, servicePlan, params, tags); err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:      resourceServiceInstancePendingStates,
		Target:       resourceServiceInstanceSucceesStates,
		Refresh:      resourceServiceInstanceStateFunc(id, "update", meta),
		Timeout:      d.Timeout(schema.TimeoutUpdate),
		PollInterval: 30 * time.Second,
		Delay:        5 * time.Second,
	}
	// Wait, catching any errors
	if _, err = stateConf.WaitForState(); err != nil {
		return err
	}

	return nil
}

func resourceServiceInstanceDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	id := d.Id()

	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("begin resourceServiceInstanceDelete")

	sm := session.ServiceManager()

	if err = sm.DeleteServiceInstance(id); err != nil {
		return err
	}
	stateConf := &resource.StateChangeConf{
		Pending:      resourceServiceInstancePendingStates,
		Target:       resourceServiceInstanceSucceesStates,
		Refresh:      resourceServiceInstanceStateFunc(id, "delete", meta),
		Timeout:      d.Timeout(schema.TimeoutDelete),
		PollInterval: 30 * time.Second,
		Delay:        5 * time.Second,
	}
	// Wait, catching any errors
	if _, err = stateConf.WaitForState(); err != nil {
		return err
	}

	session.Log.DebugMessage("Deleted Service Instance : %s", terminal.EntityNameColor(d.Id()))

	return nil
}

func resourceServiceInstanceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	session := meta.(*cfapi.Session)

	if session == nil {
		return nil, fmt.Errorf("client is nil")
	}

	sm := session.ServiceManager()

	serviceinstance, err := sm.ReadServiceInstance(d.Id())

	if err != nil {
		return nil, err
	}

	d.Set("name", serviceinstance.Name)
	d.Set("service_plan", serviceinstance.ServicePlanGUID)
	d.Set("space", serviceinstance.SpaceGUID)
	d.Set("tags", serviceinstance.Tags)

	// json_param can't be retrieved from CF, please inject manually if necessary
	d.Set("json_param", "")

	return ImportStatePassthrough(d, meta)
}

func resourceServiceInstanceStateFunc(serviceInstanceID string, operationType string, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		session := meta.(*cfapi.Session)
		sm := session.ServiceManager()
		var err error
		var serviceInstance cfapi.CCServiceInstance
		if serviceInstance, err = sm.ReadServiceInstance(serviceInstanceID); err != nil {
			// if the service instance is gone the error message should contain error code 60004 ("ServiceInstanceNotFound")
			// which is the correct behavour if the service instance has been deleted
			// e.g. CLI output: cf_service_instance.redis: Server error, status code: 404, error code: 60004, message: The service instance could not be found: babababa-d977-4e9c-9bd0-4903d146d822
			if strings.Contains(err.Error(), "error code: 60004") && operationType == "delete" {
				return serviceInstance, "succeeded", nil
			} else {
				session.Log.DebugMessage("Error on retrieving the serviceInstance %s", serviceInstanceID)
				return nil, "", err
			}
			return nil, "", err
		}

		if serviceInstance.LastOperation["type"] == operationType {
			state := serviceInstance.LastOperation["state"]
			switch state {
			case "succeeded":
				return serviceInstance, "succeeded", nil
			case "failed":
				session.Log.DebugMessage("service instance with guid=%s async provisioning has failed", serviceInstanceID)
				return nil, "", err
			}
		}

		return serviceInstance, "in progress", nil
	}
}

var resourceServiceInstancePendingStates = []string{
	"in progress",
}

var resourceServiceInstanceSucceesStates = []string{
	"succeeded",
}
