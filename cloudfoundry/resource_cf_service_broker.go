package cloudfoundry

import (
	"fmt"
	"encoding/json"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/hashcode"
)

func resourceServiceBroker() *schema.Resource {

	return &schema.Resource{

		Create: resourceServiceBrokerCreate,
		Read:   resourceServiceBrokerRead,
		Update: resourceServiceBrokerUpdate,
		Delete: resourceServiceBrokerDelete,

		Importer: &schema.ResourceImporter{
			State: ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"url": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"visibilities": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Set:      hashVisibility,
				Elem:     &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"public": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{ Type: schema.TypeString },
							Set:      resourceStringHash,
						},
						"private": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{ Type: schema.TypeString },
							Set:      resourceStringHash,
						},
					},
				},
			},
			"service_plans": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func resourceServiceBrokerCreate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id, name, url, username, password, space := getSchemaAttributes(d)

	sm := session.ServiceManager()
	if id, err = sm.CreateServiceBroker(name, url, username, password, space); err != nil {
		return err
	}

	var services []cfapi.CCService
	if services, err = sm.ReadServiceInfo(id); err != nil {
		return
	}
	setServicePlans(d, services)
	if err = updateServicePlanVisibilities(d, sm, services); err != nil {
		return
	}

	session.Log.DebugMessage("Service detail for service broker: %s:\n%# v\n", name, d.Get("service_plans"))

	d.SetId(id)
	return nil
}

func resourceServiceBrokerRead(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	var (
		serviceBroker cfapi.CCServiceBroker
		services      []cfapi.CCService
	)

	sm := session.ServiceManager()
	if serviceBroker, err = sm.ReadServiceBroker(d.Id()); err != nil {
		d.SetId("")
		return
	}
	if services, err = sm.ReadServiceInfo(d.Id()); err != nil {
		d.SetId("")
		return
	}

	d.Set("name", serviceBroker.Name)
	d.Set("url", serviceBroker.BrokerURL)
	d.Set("username", serviceBroker.AuthUserName)
	d.Set("space", serviceBroker.SpaceGUID)
	setServicePlans(d, services)
	setServicePlanVisibilities(d, services)

	return
}

func resourceServiceBrokerUpdate(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	id, name, url, username, password, space := getSchemaAttributes(d)

	sm := session.ServiceManager()
	if _, err = sm.UpdateServiceBroker(id, name, url, username, password, space); err != nil {
		d.SetId("")
		return err
	}

	var services []cfapi.CCService
	if services, err = sm.ReadServiceInfo(d.Id()); err != nil {
		d.SetId("")
		return
	}

	if err = updateServicePlanVisibilities(d, sm, services); err != nil {
		return
	}
	setServicePlans(d, services)

	session.Log.DebugMessage("Service detail for service broker: %s:\n%# v\n", name, d.Get("service_plans"))
	return
}

func resourceServiceBrokerDelete(d *schema.ResourceData, meta interface{}) (err error) {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	sm := session.ServiceManager()
	err = sm.DeleteServiceBroker(d.Id())
	return
}

func getSchemaAttributes(d *schema.ResourceData) (id, name, url, username, password, space string) {

	id = d.Id()
	name = d.Get("name").(string)
	url = d.Get("url").(string)
	if v, ok := d.GetOk("username"); ok {
		username = v.(string)
	}
	if v, ok := d.GetOk("password"); ok {
		password = v.(string)
	}
	if v, ok := d.GetOk("space"); ok {
		space = v.(string)
	}
	return
}

func setServicePlans(d *schema.ResourceData, services []cfapi.CCService) {
	servicePlans := make(map[string]interface{})
	for _, s := range services {
		for _, p := range s.ServicePlans {
			servicePlans[s.Label+"/"+p.Name] = p.ID
		}
	}
	d.Set("service_plans", servicePlans)
}

func updateServicePlanVisibilities(
	d        *schema.ResourceData,
	sm       *cfapi.ServiceManager,
	services []cfapi.CCService) (err error) {

	for _, data := range d.Get("visibilities").(*schema.Set).List() {
		dataService  := data.(map[string]interface{})
		serviceName  := dataService["service"].(string)
		publicPlans  := dataService["public"].(*schema.Set).List()
		privatePlans := dataService["private"].(*schema.Set).List()

		// ensure public plans
		for _, pp := range publicPlans {
			ppName := pp.(string)
			plan := findServicePlan(serviceName, ppName, services)
			if (plan != nil) && (plan.Public != true) {
				if err = sm.UpdatePlanVisibility(plan.ID, true); err != nil {
					return
				}
			}
		}

		// ensure private plans
		for _, pp := range privatePlans {
			ppName := pp.(string)
			plan := findServicePlan(serviceName, ppName, services)
			if (plan != nil) && (plan.Public != false) {
				if err = sm.UpdatePlanVisibility(plan.ID, false); err != nil {
					return
				}
			}
		}

	}
	return
}

func hashVisibilityObj(serviceName string, public, private []string) int {
	bytes, _ := json.Marshal(struct{
		Name    string `json:"name"`
		Public  []string `json:"public"`
		Private []string `json:"private"`
	}{
		Name: serviceName,
		Public: public,
		Private: private,
	})
	return hashcode.String(string(bytes))
}

func hashVisibility(visibility interface{}) int {
	v := visibility.(map[string]interface{})
	name,    _ := v["service"].(string)
	public,  _ := v["public"]
	private, _ := v["private"]
	var pub, priv []string

	for _, p := range public.(*schema.Set).List() {
		pub = append(pub, p.(string))
	}

	for _, p := range private.(*schema.Set).List() {
		priv = append(priv, p.(string))
	}
	return hashVisibilityObj(name, pub, priv)
}

func setServicePlanVisibilities(d *schema.ResourceData, services []cfapi.CCService) {
	var rvisibilities []interface{}

	for _, data := range d.Get("visibilities").(*schema.Set).List() {
		dataService  := data.(map[string]interface{})
		serviceName  := dataService["service"].(string)
		publicPlans  := dataService["public"].(*schema.Set).List()
		privatePlans := dataService["private"].(*schema.Set).List()

		rservice := make(map[string]interface{})
		var rpublic, rprivate []interface{}

		// read public plans
		for _, pp := range publicPlans {
			ppName := pp.(string)
			plan := findServicePlan(serviceName, ppName, services)
			if (plan != nil) && (plan.Public == true) {
				rpublic = append(rpublic, ppName)
			}
		}

		// read private plans
		for _, pp := range privatePlans {
			ppName := pp.(string)
			plan := findServicePlan(serviceName, ppName, services)
			if (plan != nil) && (plan.Public == false) {
				rprivate = append(rprivate, ppName)
			}
		}

		rservice["service"] = serviceName
		rservice["public"]   = schema.NewSet(resourceStringHash, rpublic)
		rservice["private"]  = schema.NewSet(resourceStringHash, rprivate)
		rvisibilities = append(rvisibilities, rservice)
	}

	d.Set("visibilities", schema.NewSet(hashVisibility, rvisibilities))
}


func findServicePlan(serviceName, planName string, services []cfapi.CCService) (*cfapi.CCServicePlan) {
	for _, s := range services {
		if serviceName == s.Label {
			for _, p := range s.ServicePlans {
				if planName == p.Name {
					return &p
				}
			}
		}
	}
	return nil
}
