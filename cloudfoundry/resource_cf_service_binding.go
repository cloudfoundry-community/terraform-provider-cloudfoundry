package cloudfoundry

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceServiceBinding() *schema.Resource {

	return &schema.Resource{

		Create: resourceServiceBindingCreate,
		Read:   resourceServiceBindingRead,
		Update: resourceServiceBindingUpdate,
		Delete: resourceServiceBindingDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{

			"service_instance": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"application": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			//"name": &schema.Schema{
			//	Type:     schema.TypeString,
			//	Optional: true,
			//	ForceNew: true,
			//},
			"restage_application": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"params": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"credentials": &schema.Schema{
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceServiceBindingCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.AppManager()

	var (
		bindingID   string
		credentials map[string]interface{}
	)

	appID := d.Get("application").(string)
	serviceInstanceID := d.Get("service_instance").(string)
	bindingParams := d.Get("params").(map[string]interface{})

	if bindingID, credentials, err = am.CreateServiceBinding(appID, serviceInstanceID, &bindingParams); err != nil {
		return err
	}

	d.SetId(bindingID)
	if err = d.Set("credentials", credentials); err != nil {
		return err
	}

	if d.Get("restage_application").(bool) {
		if err = am.RestageApp(appID, d.Timeout(schema.TimeoutCreate)); err != nil {
			return
		}
	}

	return nil
}

func resourceServiceBindingRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id := d.Id()
	am := session.AppManager()

	var serviceBinding cfapi.CCSeviceBinding
	if serviceBinding, err = am.ReadServiceBinding(id); err != nil {
		if strings.Contains(err.Error(), "status code: 404") {
			d.SetId("")
			err = nil
		}
	} else {
		//d.Set("name", serviceBinding.Name)
		d.Set("service_instance", serviceBinding.ServiceInstance)
		d.Set("application", serviceBinding.Application)
		d.Set("credentials", serviceBinding.Credentials)
	}

	return nil
}

func resourceServiceBindingUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	// currently a no-op
	return nil
}

func resourceServiceBindingDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.AppManager()

	if bindingID := d.Id(); bindingID != "" {
		if err := am.DeleteServiceBinding(bindingID); err != nil {
			return err
		}
	}
	return nil
}
