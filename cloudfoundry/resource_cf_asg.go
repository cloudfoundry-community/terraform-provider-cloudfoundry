package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const protocolICMP = "icmp"
const protocolTCP = "tcp"
const protocolUDP = "udp"
const protocolALL = "all"

func resourceAsg() *schema.Resource {

	return &schema.Resource{

		Create: resourceAsgCreate,
		Read:   resourceAsgRead,
		Update: resourceAsgUpdate,
		Delete: resourceAsgDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceAsgRead),
		},

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
	if value != protocolTCP && value != protocolICMP && value != protocolUDP && value != protocolALL {
		errs = append(errs, fmt.Errorf("%q must be one of 'tcp', 'icmp', 'udp' or 'all'", k))
	}
	return ws, errs
}

func resourceAsgCreate(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)
	am := session.ClientV2
	rules, err := readASGRulesFromConfig(d)
	if err != nil {
		return err
	}
	asg, _, err := am.CreateSecurityGroup(ccv2.SecurityGroup{
		Name:  d.Get("name").(string),
		Rules: rules,
	})
	if err != nil {
		return err
	}
	d.SetId(asg.GUID)

	return nil
}

func resourceAsgRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)

	am := session.ClientV2
	asg, _, err := am.GetSecurityGroup(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", asg.Name)

	tfRules := []interface{}{}
	for _, r := range asg.Rules {
		tfRule := make(map[string]interface{})
		tfRule["protocol"] = r.Protocol
		tfRule["destination"] = r.Destination
		if len(r.Ports) > 0 {
			tfRule["ports"] = r.Ports
		}
		if r.Protocol == protocolICMP {
			tfRule["type"] = r.Type.Value
			tfRule["code"] = r.Code.Value
		}
		if !r.Log.IsSet {
			tfRule["log"] = false
		} else {
			tfRule["log"] = r.Log.Value
		}

		tfRule["description"] = r.Description
		tfRules = append(tfRules, tfRule)
	}
	d.Set("rule", tfRules)

	return nil
}

func resourceAsgUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	am := session.ClientV2

	rules, err := readASGRulesFromConfig(d)
	if err != nil {
		return err
	}
	_, _, err = am.UpdateSecurityGroup(ccv2.SecurityGroup{
		GUID:  d.Id(),
		Name:  d.Get("name").(string),
		Rules: rules,
	})
	return err
}

func resourceAsgDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	_, err := session.ClientV2.DeleteSecurityGroup(d.Id())
	return err
}

func readASGRulesFromConfig(d *schema.ResourceData) (rules []ccv2.SecurityGroupRule, err error) {

	rules = []ccv2.SecurityGroupRule{}
	for _, r := range d.Get("rule").([]interface{}) {
		tfRule := r.(map[string]interface{})
		protocol := strings.ToLower(tfRule["protocol"].(string))
		asgRule := ccv2.SecurityGroupRule{
			Protocol:    tfRule["protocol"].(string),
			Destination: tfRule["destination"].(string),
		}
		if v, ok := tfRule["ports"]; ok {
			asgRule.Ports = v.(string)
		}
		if v, ok := tfRule["type"]; ok && protocol == protocolICMP {
			asgRule.Type = IntToNullInt(v.(int))
		}
		if v, ok := tfRule["code"]; ok && protocol == protocolICMP {
			asgRule.Code = IntToNullInt(v.(int))
		}
		if v, ok := tfRule["log"]; ok {
			asgRule.Log = BoolToNullBool(v.(bool))
		}
		if v, ok := tfRule["description"]; ok {
			asgRule.Description = v.(string)
		}

		if asgRule.Protocol != protocolICMP && (asgRule.Type.IsSet || asgRule.Code.IsSet) {
			err = fmt.Errorf(
				"'type' or 'code' arguments are valid only for 'icmp' protocol and not for '%s' protocol",
				asgRule.Protocol)
			return
		}

		rules = append(rules, asgRule)
	}
	return rules, err
}
