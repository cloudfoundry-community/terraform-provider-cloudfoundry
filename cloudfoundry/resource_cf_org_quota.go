package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOrgQuota() *schema.Resource {

	return &schema.Resource{

		Create: resourceOrgQuotaCreate,
		Read:   resourceOrgQuotaRead,
		Update: resourceOrgQuotaUpdate,
		Delete: resourceOrgQuotaDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceOrgQuotaRead),
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
			"total_service_keys": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
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
			"total_private_domains": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
			"total_memory": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
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
			"total_app_tasks": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},
		},
	}
}

func resourceOrgQuotaCreate(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*managers.Session)
	qm := session.ClientV2

	var quota ccv2.Quota
	if quota, _, err = qm.CreateQuota(constant.OrgQuota, readOrgQuotaResource(d)); err != nil {
		return err
	}
	d.SetId(quota.GUID)
	return nil
}

func resourceOrgQuotaRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	qm := session.ClientV2

	quota, _, err := qm.GetQuota(constant.OrgQuota, d.Id())
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
	d.Set("total_service_keys", quota.TotalServiceKeys.Value)
	d.Set("total_routes", quota.TotalRoutes)
	d.Set("total_route_ports", quota.TotalReservedRoutePorts.Value)
	d.Set("total_private_domains", quota.TotalPrivateDomains.Value)
	d.Set("total_memory", NullByteSizeToInt(quota.MemoryLimit))
	d.Set("instance_memory", NullByteSizeToInt(quota.InstanceMemoryLimit))
	d.Set("total_app_instances", quota.AppInstanceLimit.Value)
	d.Set("total_app_tasks", quota.AppTaskLimit.Value)
	return nil
}

func resourceOrgQuotaUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	qm := session.ClientV2

	quota := readOrgQuotaResource(d)
	quota.GUID = d.Id()
	_, _, err := qm.UpdateQuota(constant.OrgQuota, quota)
	return err
}

func resourceOrgQuotaDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	qm := session.ClientV2
	id := d.Id()
	// remove orgs associated to this quota by setting default quota on it
	// For context: org quota can't be removed if there is still an org associated on it
	orgs, _, err := qm.GetOrganizations()
	if err != nil {
		return err
	}
	for _, org := range orgs {
		if org.QuotaDefinitionGUID != id {
			continue
		}
		_, _, err := qm.UpdateOrganization(org.GUID, org.Name, session.DefaultQuotaGuid())
		if err != nil {
			return err
		}
	}
	_, err = qm.DeleteQuota(constant.OrgQuota, id)
	return err
}

func readOrgQuotaResource(d *schema.ResourceData) ccv2.Quota {
	quota := ccv2.Quota{
		Name:                    d.Get("name").(string),
		NonBasicServicesAllowed: d.Get("allow_paid_service_plans").(bool),
		TotalServices:           d.Get("total_services").(int),
		TotalServiceKeys:        IntToNullInt(d.Get("total_service_keys").(int)),
		TotalRoutes:             d.Get("total_routes").(int),
		TotalReservedRoutePorts: IntToNullInt(d.Get("total_route_ports").(int)),
		TotalPrivateDomains:     IntToNullInt(d.Get("total_private_domains").(int)),
		MemoryLimit:             IntToNullByteSize(d.Get("total_memory").(int)),
		InstanceMemoryLimit:     IntToNullByteSize(d.Get("instance_memory").(int)),
		AppInstanceLimit:        IntToNullInt(d.Get("total_app_instances").(int)),
		AppTaskLimit:            IntToNullInt(d.Get("total_app_tasks").(int)),
	}
	return quota
}
