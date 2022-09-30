package cloudfoundry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
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

	log.Printf("SI : %+v", userProvidedServiceInstance)
	userProvidedSI, _, err := session.ClientV3.CreateUserProvidedServiceInstance(userProvidedServiceInstance)
	log.Printf("Created SI : %+v", userProvidedSI)
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
	if _, ok := d.GetOk("syslogDrainURL"); ok {
		d.Set("syslogDrainURL", userProvidedServiceInstance.SyslogDrainURL)
		syslogSet = true
	}
	if _, ok := d.GetOk("syslog_drain_url"); ok {
		d.Set("syslog_drain_url", userProvidedServiceInstance.SyslogDrainURL)
		syslogSet = true
	}

	if !syslogSet && userProvidedServiceInstance.SyslogDrainURL.String() != "" {
		d.Set("syslog_drain_url", userProvidedServiceInstance.SyslogDrainURL)
	}

	routeServiceSet := false
	if _, ok := d.GetOk("routeServiceURL"); ok {
		d.Set("routeServiceURL", userProvidedServiceInstance.RouteServiceURL)
		routeServiceSet = true
	}
	if _, ok := d.GetOk("route_service_url"); ok {
		d.Set("route_service_url", userProvidedServiceInstance.RouteServiceURL)
		routeServiceSet = true
	}
	if !routeServiceSet && userProvidedServiceInstance.RouteServiceURL.String() != "" {
		d.Set("route_service_url", userProvidedServiceInstance.RouteServiceURL)
	}

	if _, hasJSON := d.GetOk("credentials_json"); hasJSON {
		bytes, _ := json.Marshal(userProvidedServiceInstance.Credentials)
		d.Set("credentials_json", string(bytes))
	} else {
		d.Set("credentials", userProvidedServiceInstance.Credentials)
	}
	d.Set("tags", userProvidedServiceInstance.Tags)
	return nil
}

func resourceUserProvidedServiceV3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	_, _, err := session.ClientV2.UpdateUserProvidedServiceInstance(ccv2.UserProvidedServiceInstance{
		GUID:            d.Id(),
		Name:            name,
		SpaceGuid:       space,
		Tags:            tags,
		RouteServiceUrl: routeServiceURL,
		SyslogDrainUrl:  syslogDrainURL,
		Credentials:     credentials,
	})
	return diag.FromErr(err)
}

func resourceUserProvidedServiceV3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	id := d.Id()

	jobURL, _, err := session.ClientV3.DeleteServiceInstance(id)

	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	space := d.Get("space").(string)
	poll_timeout_in_minutes := d.Get("timeout_in_minutes").(int)

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
			if err != nil {
				return true, err
			}
		}

		// Last operation initial or inprogress or job not completed, continue polling
		return false, nil
	}, 5*time.Second, time.Duration(poll_timeout_in_minutes)*time.Minute)

	return nil
}
