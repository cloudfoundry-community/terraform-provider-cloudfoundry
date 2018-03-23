package cloudfoundry

import (
	"fmt"

	"github.com/satori/go.uuid"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func dataSourceServiceInstance() *schema.Resource {

	return &schema.Resource{

		Read: dataSourceServiceInstanceRead,

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
			"service_plan_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceServiceInstanceRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	sm := session.ServiceManager()

	var (
		name_or_id      string
		space           string
		guid            string
		serviceInstance cfapi.CCServiceInstance
	)

	name_or_id = d.Get("name_or_id").(string)
	space = d.Get("space").(string)

	isUUID := uuid.FromStringOrNil(name_or_id)
	if &isUUID == nil || uuid.Equal(isUUID, uuid.Nil) {
		guid, serviceInstance, err = sm.FindServiceInstance(name_or_id, space)
	} else {
		guid = name_or_id
		serviceInstance, err = sm.ReadServiceInstance(name_or_id)
	}
	if err != nil {
		return err
	}

	d.SetId(guid)
	d.Set("name", serviceInstance.Name)
	d.Set("service_plan_id", serviceInstance.ServicePlanGUID)
	d.Set("tags", serviceInstance.Tags)

	return
}
