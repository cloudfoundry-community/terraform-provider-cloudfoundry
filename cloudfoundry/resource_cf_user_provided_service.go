package cloudfoundry

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceUserProvidedService() *schema.Resource {

	return &schema.Resource{

		Create: resourceUserProvidedServiceCreate,
		Read:   resourceUserProvidedServiceRead,
		Update: resourceUserProvidedServiceUpdate,
		Delete: resourceUserProvidedServiceDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceUserProvidedServiceRead),
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
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
				Sensitive:     true,
				ConflictsWith: []string{"credentials_json"},
			},
			"credentials_json": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{"credentials"},
				Sensitive:        true,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				ValidateFunc:     validation.StringIsJSON,
			},
			"tags": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func resourceUserProvidedServiceCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

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

	credentials := make(map[string]interface{})
	if credsJSON, hasJSON := d.GetOk("credentials_json"); hasJSON {
		err := json.Unmarshal([]byte(credsJSON.(string)), &credentials)
		if err != nil {
			return err
		}
	} else {
		for k, v := range d.Get("credentials").(map[string]interface{}) {
			credentials[k] = v.(string)
		}
	}

	tagsSchema := d.Get("tags").(*schema.Set)
	tags := make([]string, 0)
	for _, tag := range tagsSchema.List() {
		tags = append(tags, tag.(string))
	}

	usi, _, err := session.ClientV2.CreateUserProvidedServiceInstance(ccv2.UserProvidedServiceInstance{
		Name:            name,
		SpaceGuid:       space,
		Tags:            tags,
		RouteServiceUrl: routeServiceURL,
		SyslogDrainUrl:  syslogDrainURL,
		Credentials:     credentials,
	})
	if err != nil {
		return err
	}

	d.SetId(usi.GUID)

	return nil
}

func resourceUserProvidedServiceRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	ups, _, err := session.ClientV2.GetUserProvidedServiceInstance(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", ups.Name)
	d.Set("space", ups.SpaceGuid)

	syslogSet := false
	if _, ok := d.GetOk("syslogDrainURL"); ok {
		d.Set("syslogDrainURL", ups.SyslogDrainUrl)
		syslogSet = true
	}
	if _, ok := d.GetOk("syslog_drain_url"); ok {
		d.Set("syslog_drain_url", ups.SyslogDrainUrl)
		syslogSet = true
	}

	if !syslogSet && ups.SyslogDrainUrl != "" {
		d.Set("syslog_drain_url", ups.SyslogDrainUrl)
	}

	routeServiceSet := false
	if _, ok := d.GetOk("routeServiceURL"); ok {
		d.Set("routeServiceURL", ups.RouteServiceUrl)
		routeServiceSet = true
	}
	if _, ok := d.GetOk("route_service_url"); ok {
		d.Set("route_service_url", ups.RouteServiceUrl)
		routeServiceSet = true
	}
	if !routeServiceSet && ups.RouteServiceUrl != "" {
		d.Set("route_service_url", ups.RouteServiceUrl)
	}

	if _, hasJSON := d.GetOk("credentials_json"); hasJSON {
		bytes, _ := json.Marshal(ups.Credentials)
		d.Set("credentials_json", string(bytes))
	} else {
		d.Set("credentials", ups.Credentials)
	}
	d.Set("tags", ups.Tags)
	return nil
}

func resourceUserProvidedServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	syslogDrainURL := d.Get("syslog_drain_url").(string)
	routeServiceURL := d.Get("route_service_url").(string)
	space := d.Get("space").(string)
	// should be removed when syslogDrainURL and routeServiceURL will be removed
	if syslogDrainURL == "" {
		syslogDrainURL = d.Get("syslogDrainURL").(string)
	}
	if routeServiceURL == "" {
		routeServiceURL = d.Get("routeServiceURL").(string)
	}

	credentials := make(map[string]interface{})
	if credsJSON, hasJSON := d.GetOk("credentials_json"); hasJSON {
		err := json.Unmarshal([]byte(credsJSON.(string)), &credentials)
		if err != nil {
			return err
		}
	} else {
		for k, v := range d.Get("credentials").(map[string]interface{}) {
			credentials[k] = v.(string)
		}
	}
	tagsSchema := d.Get("tags").(*schema.Set)
	tags := make([]string, 0)
	for _, tag := range tagsSchema.List() {
		tags = append(tags, tag.(string))
	}
	_, _, err := session.ClientV2.UpdateUserProvidedServiceInstance(ccv2.UserProvidedServiceInstance{
		GUID:            d.Id(),
		Name:            name,
		SpaceGuid:       space,
		Tags:            tags,
		RouteServiceUrl: routeServiceURL,
		SyslogDrainUrl:  syslogDrainURL,
		Credentials:     credentials,
	})
	return err
}

func resourceUserProvidedServiceDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	_, err := session.ClientV2.DeleteUserProvidedServiceInstance(d.Id())
	return err
}
