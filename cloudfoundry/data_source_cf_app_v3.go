package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	resources "code.cloudfoundry.org/cli/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	uuid "github.com/satori/go.uuid"
)

func dataSourceAppV3() *schema.Resource {

	return &schema.Resource{

		ReadContext: dataSourceAppRead,

		Schema: map[string]*schema.Schema{

			"name_or_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"instances": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"disk_quota": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"stack": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"buildpack": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"command": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_ssh": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment": &schema.Schema{
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
			},
			"health_check_http_endpoint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"health_check_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"health_check_timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
		},
	}
}

func dataSourceAppV3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	var (
		nameOrID string
		space    string
		app      resources.Application
		err      error
	)

	nameOrID = d.Get("name_or_id").(string)
	space = d.Get("space").(string)

	nameQuery := ccv3.Query{
		Key:    ccv3.NameFilter,
		Values: []string{nameOrID},
	}
	spaceQuery := ccv3.Query{
		Key:    ccv3.SpaceGUIDFilter,
		Values: []string{space},
	}

	isUUID := uuid.FromStringOrNil(nameOrID)
	if uuid.Equal(isUUID, uuid.Nil) {

		apps, _, err := session.ClientV3.GetApplications(nameQuery, spaceQuery)
		if err != nil {
			return diag.FromErr(err)
		}
		if len(apps) == 0 {
			return diag.FromErr(NotFound)
		}
		app = apps[0]
	} else {
		apps, _, err := session.ClientV3.GetApplications(nameQuery, spaceQuery)
		if err != nil {
			return diag.FromErr(err)
		}
		app = apps[0]
	}

	guid := app.GUID

	d.SetId(app.GUID)
	d.Set("name", app.Name)
	d.Set("space", app.SpaceGUID)
	d.Set("buildpack", app.LifecycleBuildpacks[0])
	// In v3 the following information are kept in separate endpoints.
	// Process
	appProcesses, _, err := session.ClientV3.GetApplicationProcesses(guid)
	if err != nil {
		return diag.FromErr(err)
	}
	appProcess := appProcesses[0]
	d.Set("instances", appProcess.Instances.Value)
	d.Set("memory", appProcess.MemoryInMB.Value)
	d.Set("disk_quota", appProcess.DiskInMB.Value)
	d.Set("stack", app.StackName)
	d.Set("command", appProcess.Command.Value)

	// Check for sshEnabled feature
	sshEnabled, _, err := session.ClientV3.GetAppFeature(guid, "ssh")
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("enable_ssh", sshEnabled.Enabled)

	// Environment
	appEnvironment, _, err := session.ClientV3.GetApplicationEnvironment(guid)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("environment", appEnvironment.EnvironmentVariables)
	d.Set("state", app.State)
	d.Set("health_check_http_endpoint", appProcess.HealthCheckEndpoint)
	d.Set("health_check_type", appProcess.HealthCheckType)
	d.Set("health_check_timeout", appProcess.HealthCheckTimeout)

	err = metadataRead(appMetadata, d, meta, true)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
