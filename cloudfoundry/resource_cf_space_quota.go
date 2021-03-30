package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSpaceQuota() *schema.Resource {

	return &schema.Resource{

		Create: resourceSpaceQuotaCreate,
		Read:   resourceSpaceQuotaRead,
		Update: resourceSpaceQuotaUpdate,
		Delete: resourceSpaceQuotaDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceSpaceQuotaRead),
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

func resourceSpaceQuotaCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	qm := session.ClientV2

	quota, _, err := qm.CreateQuota(constant.SpaceQuota, readSpaceQuotaResource(d))
	if err != nil {
		return err
	}
	d.SetId(quota.GUID)
	return nil
}

func resourceSpaceQuotaRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	qm := session.ClientV2
	quota, _, err := qm.GetQuota(constant.SpaceQuota, d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", quota.Name)
	d.Set("allow_paid_service_plans", quota.NonBasicServicesAllowed)
	d.Set("total_services", quota.TotalServices)
	d.Set("total_routes", quota.TotalRoutes)
	d.Set("total_route_ports", quota.TotalReservedRoutePorts.Value)
	d.Set("total_memory", NullByteSizeToInt(quota.MemoryLimit))
	d.Set("total_service_keys", quota.TotalServiceKeys.Value)
	d.Set("instance_memory", NullByteSizeToInt(quota.InstanceMemoryLimit))
	d.Set("total_app_instances", quota.AppInstanceLimit.Value)
	d.Set("org", quota.OrganizationGUID)
	d.Set("total_app_tasks", quota.AppTaskLimit.Value)

	return nil
}

func resourceSpaceQuotaUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	qm := session.ClientV2
	quota := readSpaceQuotaResource(d)
	quota.GUID = d.Id()
	_, _, err := qm.UpdateQuota(constant.SpaceQuota, quota)
	return err
}

func resourceSpaceQuotaDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	qm := session.ClientV2
	_, err := qm.DeleteQuota(constant.SpaceQuota, d.Id())
	return err
}

func readSpaceQuotaResource(d *schema.ResourceData) ccv2.Quota {
	quota := ccv2.Quota{
		Name:                    d.Get("name").(string),
		NonBasicServicesAllowed: d.Get("allow_paid_service_plans").(bool),
		TotalServices:           d.Get("total_services").(int),
		TotalServiceKeys:        IntToNullInt(d.Get("total_service_keys").(int)),
		TotalRoutes:             d.Get("total_routes").(int),
		TotalReservedRoutePorts: IntToNullInt(d.Get("total_route_ports").(int)),
		MemoryLimit:             IntToNullByteSize(d.Get("total_memory").(int)),
		InstanceMemoryLimit:     IntToNullByteSize(d.Get("instance_memory").(int)),
		AppInstanceLimit:        IntToNullInt(d.Get("total_app_instances").(int)),
		AppTaskLimit:            IntToNullInt(d.Get("total_app_tasks").(int)),
		OrganizationGUID:        d.Get("org").(string),
	}
	if v, ok := d.GetOk("org"); ok {
		quota.OrganizationGUID = v.(string)
	}
	return quota
}
