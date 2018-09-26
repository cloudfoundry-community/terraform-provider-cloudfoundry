package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceSpaceQuota() *schema.Resource {

	return &schema.Resource{

		Create: resourceSpaceQuotaCreate,
		Read:   resourceSpaceQuotaRead,
		Update: resourceSpaceQuotaUpdate,
		Delete: resourceSpaceQuotaDelete,

		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"allow_paid_service_plans": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"total_services": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"total_routes": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"total_route_ports": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"total_memory": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"total_service_keys": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"instance_memory": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"total_app_instances": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"total_app_tasks": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  5,
			},
		},
	}
}

func resourceSpaceQuotaCreate(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	qm := session.QuotaManager()

	var id string
	if id, err = qm.CreateQuota(cfapi.SpaceQuota, readSpaceQuotaResource(d)); err != nil {
		return err
	}
	d.SetId(id)
	return nil
}

func resourceSpaceQuotaRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	qm := session.QuotaManager()

	var quota cfapi.CCQuota
	if quota, err = qm.ReadQuota(cfapi.SpaceQuota, d.Id()); err != nil {
		return err
	}

	d.Set("name", quota.Name)
	d.Set("allow_paid_service_plans", quota.NonBasicServicesAllowed)
	d.Set("total_services", quota.TotalServices)
	d.Set("total_routes", quota.TotalRoutes)
	d.Set("total_route_ports", quota.TotalReserveredPorts)
	d.Set("total_memory", quota.MemoryLimit)
	d.Set("total_service_keys", quota.TotalServiceKeys)
	d.Set("instance_memory", quota.InstanceMemoryLimit)
	d.Set("total_app_instances", quota.AppInstanceLimit)
	d.Set("org", quota.OrgGUID)
	d.Set("total_app_tasks", quota.AppTaskLimit)

	return nil
}

func resourceSpaceQuotaUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	qm := session.QuotaManager()
	quota := readSpaceQuotaResource(d)
	quota.ID = d.Id()
	return qm.UpdateQuota(cfapi.SpaceQuota, quota)
}

func resourceSpaceQuotaDelete(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	qm := session.QuotaManager()
	return qm.DeleteQuota(cfapi.SpaceQuota, d.Id())
}

func readSpaceQuotaResource(d *schema.ResourceData) cfapi.CCQuota {
	quota := cfapi.CCQuota{
		Name:                    d.Get("name").(string),
		AppInstanceLimit:        d.Get("total_app_instances").(int),
		AppTaskLimit:            d.Get("total_app_tasks").(int),
		InstanceMemoryLimit:     int64(d.Get("instance_memory").(int)),
		MemoryLimit:             int64(d.Get("total_memory").(int)),
		NonBasicServicesAllowed: d.Get("allow_paid_service_plans").(bool),
		TotalServices:           d.Get("total_services").(int),
		TotalServiceKeys:        d.Get("total_service_keys").(int),
		TotalRoutes:             d.Get("total_routes").(int),
		TotalReserveredPorts:    d.Get("total_route_ports").(int),
		OrgGUID:                 d.Get("org").(string),
	}
	if v, ok := d.GetOk("org"); ok {
		quota.OrgGUID = v.(string)
	}
	return quota
}
