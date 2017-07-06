package cloudfoundry

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAsg() *schema.Resource {

	return &schema.Resource{

		Create: resourceAsgCreate,
		Read:   resourceAsgRead,
		Update: resourceAsgUpdate,
		Delete: resourceAsgDelete,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"rule": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateAsgProtocol,
						},
						"destination": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"ports": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"code": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"log": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func validateAsgProtocol(v interface{}, k string) (ws []string, errs []error) {
	value := v.(string)
	if value != "tcp" && value != "icmp" && value != "udp" && value != "all" {
		errs = append(errs, fmt.Errorf("%q must be one of 'tcp', 'icmp', 'udp' or 'all'", k))
	}
	return
}

func resourceAsgCreate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.ASGManager()

	rules, err := readASGRulesFromConfig(d)
	if err != nil {
		return err
	}
	id, err := am.CreateASG(d.Get("name").(string), rules)
	if err != nil {
		return err
	}
	d.SetId(id)

	return nil
}

func resourceAsgRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.ASGManager()
	asg, err := am.GetASG(d.Id())
	if err != nil {
		return err
	}

	session.Log.DebugMessage("Read ASG from CC: %# v", asg)

	d.Set("name", asg.Name)

	tfRules := []interface{}{}
	for _, r := range asg.Rules {
		tfRule := make(map[string]interface{})
		tfRule["protocol"] = r.Protocol
		tfRule["destination"] = r.Destination
		if len(r.Ports) > 0 {
			tfRule["ports"] = r.Ports
		}
		if r.Protocol == "icmp" {
			tfRule["type"] = r.Type
			tfRule["code"] = r.Code
		}
		tfRule["log"] = r.Log
		tfRule["description"] = r.Description
		tfRules = append(tfRules, tfRule)
	}
	d.Set("rule", tfRules)

	return nil
}

func resourceAsgUpdate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	am := session.ASGManager()

	rules, err := readASGRulesFromConfig(d)
	if err != nil {
		return err
	}
	err = am.UpdateASG(d.Id(), d.Get("name").(string), rules)
	if err != nil {
		return err
	}
	return nil
}

func resourceAsgDelete(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*cfapi.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}
	return session.ASGManager().Delete(d.Id())
}

func readASGRulesFromConfig(d *schema.ResourceData) (rules []cfapi.CCASGRule, err error) {

	rules = []cfapi.CCASGRule{}
	for _, r := range d.Get("rule").([]interface{}) {

		tfRule := r.(map[string]interface{})
		asgRule := cfapi.CCASGRule{
			Protocol:    tfRule["protocol"].(string),
			Destination: tfRule["destination"].(string),
		}
		if v, ok := tfRule["ports"]; ok {
			asgRule.Ports = v.(string)
		}
		if v, ok := tfRule["type"]; ok {
			asgRule.Type = v.(int)
		}
		if v, ok := tfRule["code"]; ok {
			asgRule.Code = v.(int)
		}
		if v, ok := tfRule["log"]; ok {
			asgRule.Log = v.(bool)
		}
		if v, ok := tfRule["description"]; ok {
			asgRule.Description = v.(string)
		}

		if asgRule.Protocol != "icmp" && (asgRule.Type > 0 || asgRule.Code > 0) {
			err = fmt.Errorf(
				"'type' or 'code' arguments are valid only for 'icmp' protocol and not for '%s' protocol",
				asgRule.Protocol)
			return
		}

		rules = append(rules, asgRule)
	}
	return
}
