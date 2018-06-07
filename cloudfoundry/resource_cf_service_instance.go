package cloudfoundry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"golang.org/x/sync/semaphore"

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
			State: ImportStatePassthrough,
		},

		SchemaVersion: 1,
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
			"service_plan_concurrency": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Allows for the concurrency of changes to service instances, sharing a particular service_plan, to be restricted.",
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

	if sem := limitConcurrency(d); sem != nil {
		defer (*sem).Release(1)
	}

	if id, err = sm.CreateServiceInstance(name, servicePlan, space, params, tags); err != nil {
		return err
	}
	session.Log.DebugMessage("New Service Instance : %# v", id)

	// TODO deal with asynchronous responses

	d.SetId(id)

	return nil
}

func resourceServiceInstanceRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("Reading Service Instance : %s", d.Id())

	sm := session.ServiceManager()
	var serviceInstance cfapi.CCServiceInstance

	serviceInstance, err = sm.ReadServiceInstance(d.Id())
	if err != nil {
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

	if sem := limitConcurrency(d); sem != nil {
		defer (*sem).Release(1)
	}

	if _, err = sm.UpdateServiceInstance(id, name, servicePlan, params, tags); err != nil {
		return err
	}

	return nil
}

func resourceServiceInstanceDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("begin resourceServiceInstanceDelete")

	sm := session.ServiceManager()

	if sem := limitConcurrency(d); sem != nil {
		defer (*sem).Release(1)
	}

	err = sm.DeleteServiceInstance(d.Id())
	if err != nil {
		return err
	}

	session.Log.DebugMessage("Deleted Service Instance : %s", d.Id())

	return nil
}

// #######################
// # Concurrency Limiter #
// #######################
// Updates to some types of services in Cloud Foundry (generally badly behaved service brokers)
// cannot be done in parallel or need to be done with limited concurrency.  This is a hack around
// the lack of a terraform provided method to limit the level of concurrency around a particular
// type of resource.  The idea here is that for all of the cf_service_instance resources
// which share a service_plan ID and set the service_plan_concurrency to a value greater than
// zero, then this code will cause all creates/updates/deletes of those service plan instances
// to be throttled to the defined concurrency limit.
//
// Limitations
// - The concurrency defined by the first resource to use a given service_plan ID wins
// - cf_service_instance resources of the same service plan which do not define service_plan_concurrency
//   will not take part in the limitation on concurrency

var concurrencySemaphore = make(map[string]*semaphore.Weighted)
var concurrencySemaphoreMutex = &sync.Mutex{}

func limitConcurrency(d *schema.ResourceData) *semaphore.Weighted {
	if d.Get("service_plan_concurrency").(int) <= 0 {
		// if no limit, then just skip
		return nil
	}

	concurrencySemaphoreMutex.Lock()
	if _, ok := concurrencySemaphore[d.Get("service_plan").(string)]; !ok {
		concurrencySemaphore[d.Get("service_plan").(string)] = semaphore.NewWeighted(int64(d.Get("service_plan_concurrency").(int)))
	}
	sem := concurrencySemaphore[d.Get("service_plan").(string)]
	concurrencySemaphoreMutex.Unlock()

	sem.Acquire(context.TODO(), 1)
	return sem
}
