package cloudfoundry

import (
	"context"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	uuid "github.com/satori/go.uuid"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceApp() *schema.Resource {

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
			"buildpacks": &schema.Schema{
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"buildpack"},
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

func dataSourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	session := meta.(*managers.Session)
	if session == nil {
		return diag.Errorf("client is nil")
	}

	var (
		nameOrId string
		space    string
		app      resources.Application
		err      error
	)

	nameOrId = d.Get("name_or_id").(string)
	space = d.Get("space").(string)

	isUUID := uuid.FromStringOrNil(nameOrId)

	var query []ccv3.Query

	if uuid.Equal(isUUID, uuid.Nil) {
		query = []ccv3.Query{
			{
				Key:    ccv3.NameFilter,
				Values: []string{nameOrId},
			},
			{
				Key:    ccv3.SpaceGUIDFilter,
				Values: []string{space},
			},
		}
	} else {
		query = []ccv3.Query{
			{
				Key:    ccv3.GUIDFilter,
				Values: []string{nameOrId},
			},
			{
				Key:    ccv3.SpaceGUIDFilter,
				Values: []string{space},
			},
		}
	}

	apps, _, err := session.ClientV3.GetApplications(query...)

	if err != nil {
		return diag.FromErr(err)
	}

	if len(apps) == 0 {
		return diag.FromErr(NotFound)
	}
	app = apps[0]

	d.SetId(app.GUID)
	d.Set("name", app.Name)
	d.Set("space", app.SpaceGUID)

	proc, _, err := session.ClientV3.GetApplicationProcessByType(d.Id(), constant.ProcessTypeWeb)

	if err != nil {
		return diag.FromErr(err)
	}

	if proc.Instances.IsSet {
		d.Set("instances", proc.Instances.Value)
	}

	if proc.MemoryInMB.IsSet {
		d.Set("memory", proc.MemoryInMB.Value)
	}
	if proc.DiskInMB.IsSet {
		d.Set("disk_quota", proc.DiskInMB.Value)
	}

	d.Set("stack", app.StackName)
	if bpkg := app.LifecycleBuildpacks; len(bpkg) > 0 {
		d.Set("buildpacks", bpkg)
		d.Set("buildpack", bpkg[0])
	}

	if proc.Command.IsSet {
		d.Set("command", proc.Command.Value)
	}

	enableSSH, _, err := session.ClientV3.GetAppFeature(d.Id(), "ssh")
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("enable_ssh", enableSSH.Enabled)

	env, err := session.BitsManager.GetAppEnvironmentVariables(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("environment", env)

	d.Set("state", app.State)

	if proc.HealthCheckEndpoint != "" {
		d.Set("health_check_http_endpoint", proc.HealthCheckEndpoint)
	}
	if proc.HealthCheckTimeout != 0 {
		d.Set("health_check_timeout", proc.HealthCheckTimeout)
	}
	if proc.HealthCheckType != "" {
		d.Set("health_check_type", proc.HealthCheckType)
	}

	err = metadataRead(appMetadata, d, meta, true)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
