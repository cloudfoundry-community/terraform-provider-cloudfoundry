package cloudfoundry

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/structure"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceUserProvidedService() *schema.Resource {

	return &schema.Resource{

		Create: resourceUserProvidedServiceCreate,
		Read:   resourceUserProvidedServiceRead,
		Update: resourceUserProvidedServiceUpdate,
		Delete: resourceUserProvidedServiceDelete,

		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"syslog_drain_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"syslogDrainURL": &schema.Schema{
				Type:       schema.TypeString,
				Optional:   true,
				Default:    "",
				Deprecated: "Use syslog_drain_url, Terraform complain about field name may only contain lowercase alphanumeric characters & underscores",
			},
			"route_service_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"routeServiceURL": &schema.Schema{
				Type:       schema.TypeString,
				Optional:   true,
				Default:    "",
				Deprecated: "Use route_service_url, Terraform complain about field name may only contain lowercase alphanumeric characters & underscores",
			},
			"credentials": &schema.Schema{
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"credentials_json"},
			},
			"credentials_json": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{"credentials"},
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
		},
	}
}

func resourceUserProvidedServiceCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		id          string
		credentials map[string]interface{}
	)

	name := d.Get("name").(string)
	space := d.Get("space").(string)
	syslogDrainURL := d.Get("syslog_drain_url").(string)
	routeServiceURL := d.Get("route_service_url").(string)

	// should be removed when syslogDrainURL and routeServiceURL will be removed
	if syslogDrainURL == "" {
		syslogDrainURL = d.Get("syslogDrainURL").(string)
	}
	if routeServiceURL == "" {
		routeServiceURL = d.Get("routeServiceURL").(string)
	}

	credentials = make(map[string]interface{})
	if credsJSON, hasJSON := d.GetOk("credentials_json"); hasJSON {
		if err = json.Unmarshal([]byte(credsJSON.(string)), &credentials); err != nil {
			return err
		}
	} else {
		for k, v := range d.Get("credentials").(map[string]interface{}) {
			credentials[k] = v.(string)
		}
	}

	sm := session.ServiceManager()

	if id, err = sm.CreateUserProvidedService(name, space, credentials, syslogDrainURL, routeServiceURL); err != nil {
		return err
	}
	session.Log.DebugMessage("New User Provided Service : %# v", id)

	d.SetId(id)

	return nil
}

func resourceUserProvidedServiceRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("Reading User Provided Service : %s", d.Id())

	sm := session.ServiceManager()
	var ups cfapi.CCUserProvidedService

	if ups, err = sm.ReadUserProvidedService(d.Id()); err != nil {
		return err
	}

	d.Set("name", ups.Name)
	d.Set("space", ups.SpaceGUID)

	// should be changed when syslogDrainURL and routeServiceURL will be removed, this will be:
	// d.Set("syslog_drain_url", ups.SyslogDrainURL)
	// d.Set("route_service_url", ups.RouteServiceURL)
	if _, ok := d.GetOk("syslogDrainURL"); ok {
		d.Set("syslogDrainURL", ups.SyslogDrainURL)
	} else {
		d.Set("syslog_drain_url", ups.SyslogDrainURL)
	}
	if _, ok := d.GetOk("routeServiceURL"); ok {
		d.Set("routeServiceURL", ups.RouteServiceURL)
	} else {
		d.Set("route_service_url", ups.RouteServiceURL)
	}

	if _, hasJSON := d.GetOk("credentials_json"); hasJSON {
		bytes, _ := json.Marshal(ups.Credentials)
		d.Set("credentials_json", string(bytes))
	} else {
		d.Set("credentials", ups.Credentials)
	}

	session.Log.DebugMessage("Read User Provided Service : %# v", ups)

	return nil
}

func resourceUserProvidedServiceUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	sm := session.ServiceManager()

	session.Log.DebugMessage("Updating User Provided service %s ", d.Id())

	var (
		credentials map[string]interface{}
	)

	id := d.Id()
	name := d.Get("name").(string)
	syslogDrainURL := d.Get("syslog_drain_url").(string)
	routeServiceURL := d.Get("route_service_url").(string)

	//should be removed when syslogDrainURL and routeServiceURL will be removed
	if syslogDrainURL == "" {
		syslogDrainURL = d.Get("syslogDrainURL").(string)
	}
	if routeServiceURL == "" {
		routeServiceURL = d.Get("routeServiceURL").(string)
	}

	credentials = make(map[string]interface{})
	if credsJSON, hasJSON := d.GetOk("credentials_json"); hasJSON {
		if err = json.Unmarshal([]byte(credsJSON.(string)), &credentials); err != nil {
			return err
		}
	} else {
		for k, v := range d.Get("credentials").(map[string]interface{}) {
			credentials[k] = v.(string)
		}
	}

	if _, err = sm.UpdateUserProvidedService(id, name, credentials, syslogDrainURL, routeServiceURL); err != nil {
		return err
	}

	return nil
}

func resourceUserProvidedServiceDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	session.Log.DebugMessage("begin resourceServiceInstanceDelete")

	sm := session.ServiceManager()

	if err = sm.DeleteServiceInstance(d.Id()); err != nil {
		return err
	}

	session.Log.DebugMessage("Deleted Service Instance : %s", d.Id())

	return nil
}
