package cloudfoundry

import (
	"context"
	"encoding/json"
	"log"

	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceUserProvidedServiceV3() *schema.Resource {

	return &schema.Resource{

		CreateContext: resourceUserProvidedServiceV3Create,
		ReadContext:   resourceUserProvidedServiceV3Read,
		UpdateContext: resourceUserProvidedServiceV3Update,
		DeleteContext: resourceUserProvidedServiceV3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: ImportReadContext(resourceUserProvidedServiceRead),
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

func resourceUserProvidedServiceV3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	// credentials := types.OptionalObject{}

	credentials := make(map[string]interface{})
	if credsJSON, hasJSON := d.GetOk("credentials_json"); hasJSON {
		err := json.Unmarshal([]byte(credsJSON.(string)), &credentials)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		for k, v := range d.Get("credentials").(map[string]interface{}) {
			credentials[k] = v.(string)
		}
	}
	credentialsFormat := types.OptionalObject{
		IsSet: len(credentials) != 0,
		Value: credentials,
	}

	tagsSchema := d.Get("tags").(*schema.Set)
	tags := make([]string, 0)
	for _, tag := range tagsSchema.List() {
		tags = append(tags, tag.(string))
	}

	tagsFormat := types.OptionalStringSlice{
		IsSet: tags != nil,
		Value: tags,
	}

	syslogDrainURLFormat := types.OptionalString{
		IsSet: syslogDrainURL != "",
		Value: syslogDrainURL,
	}

	routeServiceURLFormat := types.OptionalString{
		IsSet: routeServiceURL != "",
		Value: routeServiceURL,
	}

	userProvidedServiceInstance := resources.ServiceInstance{
		Type:            resources.UserProvidedServiceInstance,
		Name:            name,
		SpaceGUID:       space,
		Credentials:     credentialsFormat,
		Tags:            tagsFormat,
		SyslogDrainURL:  syslogDrainURLFormat,
		RouteServiceURL: routeServiceURLFormat,
	}

	// log.Printf("SI : %+v", userProvidedServiceInstance)
	userProvidedSI, _, err := session.ClientV3.CreateUserProvidedServiceInstance(userProvidedServiceInstance)
	// log.Printf("Created SI : %+v", userProvidedSI)
	if err != nil {
		return diag.FromErr(err)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(userProvidedSI.GUID)
	return nil
}

func resourceUserProvidedServiceV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	name := d.Get("name").(string)
	space := d.Get("space").(string)

	userProvidedServiceInstance, _, _, err := session.ClientV3.GetServiceInstanceByNameAndSpace(name, space)
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("name", userProvidedServiceInstance.Name)
	d.Set("space", userProvidedServiceInstance.SpaceGUID)

	syslogSet := false
	syslogDrainURL := userProvidedServiceInstance.SyslogDrainURL.String()
	if _, ok := d.GetOk("syslogDrainURL"); ok {
		d.Set("syslogDrainURL", syslogDrainURL)
		syslogSet = true
	}
	if _, ok := d.GetOk("syslog_drain_url"); ok {
		d.Set("syslog_drain_url", syslogDrainURL)
		syslogSet = true
	}

	if !syslogSet && syslogDrainURL != "" {
		d.Set("syslog_drain_url", syslogDrainURL)
	}

	routeServiceSet := false
	routeServiceURL := userProvidedServiceInstance.RouteServiceURL.String()
	if _, ok := d.GetOk("routeServiceURL"); ok {
		d.Set("routeServiceURL", routeServiceURL)
		routeServiceSet = true
	}
	if _, ok := d.GetOk("route_service_url"); ok {
		d.Set("route_service_url", routeServiceURL)
		routeServiceSet = true
	}
	if !routeServiceSet && routeServiceURL != "" {
		d.Set("route_service_url", routeServiceURL)
	}

	credentials, _, err := session.ClientV3.GetUserProvidedServiceInstanceCredentails(d.Id())

	if _, hasJSON := d.GetOk("credentials_json"); hasJSON {
		bytes, _ := json.Marshal(credentials)
		log.Printf("Creds : %s //// state: %s", string(bytes), d.Get("credentials_json"))
		d.Set("credentials_json", string(bytes))
	} else {
		d.Set("credentials", credentials)
	}
	d.Set("tags", userProvidedServiceInstance.Tags.Value)
	return nil
}

func resourceUserProvidedServiceV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
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
			return diag.FromErr(err)
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

	_, _, err := session.ClientV3.UpdateUserProvidedServiceInstance(d.Id(), resources.ServiceInstance{
		Name: name,
		Tags: types.OptionalStringSlice{
			IsSet: tags != nil,
			Value: tags,
		},
		RouteServiceURL: types.OptionalString{
			IsSet: routeServiceURL != "",
			Value: routeServiceURL,
		},
		SyslogDrainURL: types.OptionalString{
			IsSet: syslogDrainURL != "",
			Value: syslogDrainURL,
		},
		Credentials: types.OptionalObject{
			IsSet: len(credentials) > 0,
			Value: credentials,
		},
	})

	// log.Printf("updated service instance: %+v", updated)
	return diag.FromErr(err)
}

func resourceUserProvidedServiceV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	// No polling needed since no discussion with service broker
	_, _, err := session.ClientV3.DeleteServiceInstance(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
