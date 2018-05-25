package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	uuid "github.com/satori/go.uuid"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func dataSourceApp() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceAppRead,

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
			"ports": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Set:      resourceIntegerSet,
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
		},
	}
}

func dataSourceAppRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		name_or_id string
		space      string
		app        cfapi.CCApp
	)

	name_or_id = d.Get("name_or_id").(string)
	space = d.Get("space").(string)

	am := session.AppManager()

	isUUID := uuid.FromStringOrNil(name_or_id)
	if &isUUID == nil || uuid.Equal(isUUID, uuid.Nil) {
		app, err = am.FindApp(name_or_id, space)
	} else {
		app, err = am.ReadApp(name_or_id)
	}
	if err != nil {
		return err
	}

	d.SetId(app.ID)
	d.Set("name", app.Name)
	d.Set("space", app.SpaceGUID)
	d.Set("instances", app.Instances)
	d.Set("memory", app.Memory)
	d.Set("disk_quota", app.DiskQuota)
	d.Set("stack", app.StackGUID)
	d.Set("buildpack", app.Buildpack)
	d.Set("command", app.Command)
	d.Set("enable_ssh", app.EnableSSH)
	d.Set("environment", app.Environment)
	d.Set("state", app.State)

	if app.HealthCheckHTTPEndpoint != nil {
		d.Set("health_check_http_endpoint", app.HealthCheckHTTPEndpoint)
	}
	if app.HealthCheckType != nil {
		d.Set("health_check_type", app.HealthCheckType)
	}
	if app.HealthCheckTimeout != nil {
		d.Set("health_check_timeout", app.HealthCheckTimeout)
	}

	ports := []interface{}{}
	for _, p := range *app.Ports {
		ports = append(ports, p)
	}
	d.Set("ports", schema.NewSet(resourceIntegerSet, ports))

	return
}
