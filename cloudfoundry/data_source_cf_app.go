package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/satori/go.uuid"
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
		app      ccv2.Application
		err      error
	)

	nameOrId = d.Get("name_or_id").(string)
	space = d.Get("space").(string)

	isUUID := uuid.FromStringOrNil(nameOrId)
	if uuid.Equal(isUUID, uuid.Nil) {
		apps, _, err := session.ClientV2.GetApplications(ccv2.FilterByName(nameOrId), ccv2.FilterBySpace(space))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(apps) == 0 {
			return diag.FromErr(NotFound)
		}
		app = apps[0]
	} else {
		app, _, err = session.ClientV2.GetApplication(nameOrId)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(app.GUID)
	d.Set("name", app.Name)
	d.Set("space", app.SpaceGUID)
	d.Set("instances", app.Instances.Value)
	d.Set("memory", app.Memory.Value)
	d.Set("disk_quota", app.DiskQuota.Value)
	d.Set("stack", app.StackGUID)
	d.Set("buildpack", app.Buildpack.Value)
	d.Set("command", app.Command.Value)
	d.Set("enable_ssh", app.EnableSSH.Value)
	d.Set("environment", app.EnvironmentVariables)
	d.Set("state", app.State)
	d.Set("health_check_http_endpoint", app.HealthCheckHTTPEndpoint)
	d.Set("health_check_type", app.HealthCheckType)
	d.Set("health_check_timeout", app.HealthCheckTimeout)

	err = metadataRead(appMetadata, d, meta, true)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
