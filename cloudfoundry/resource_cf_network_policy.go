package cloudfoundry

import (
	"fmt"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/hashcode"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceNetworkPolicy() *schema.Resource {

	return &schema.Resource{

		Create: resourceNetworkPolicyCreate,
		Read:   resourceNetworkPolicyRead,
		Update: resourceNetworkPolicyUpdate,
		Delete: resourceNetworkPolicyDelete,

		Schema: map[string]*schema.Schema{
			"policy": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Set: func(v interface{}) int {
					elem := v.(map[string]interface{})
					str := fmt.Sprintf("%s-%s-%s-%s",
						elem["source_app"],
						elem["destination_app"],
						elem["protocol"],
						elem["port"],
					)
					return hashcode.String(str)
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_app": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"destination_app": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"port": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(i interface{}, s string) ([]string, []error) {
								_, _, err := portRangeParse(i.(string))
								if err != nil {
									return []string{}, []error{err}
								}
								return []string{}, []error{}
							},
						},
						"protocol": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp"}, false),
							Default:      "tcp",
						},
					},
				},
			},
		},
	}
}

func resourceNetworkPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	guid, err := uuid.GenerateUUID()
	if err != nil {
		return err
	}
	d.SetId(guid)
	policiesTf := getListOfStructs(d.Get("policy"))
	return session.NetClient.CreatePolicies(resourceNetworkPoliciesToPolicies(policiesTf))
}

func resourceNetworkPolicyRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	policiesTf := getListOfStructs(d.Get("policy"))

	idsMap := make(map[string]bool)
	for _, p := range policiesTf {
		idsMap[p["source_app"].(string)] = true
		idsMap[p["destination_app"].(string)] = true
	}
	ids := make([]string, 0)
	for k := range idsMap {
		ids = append(ids, k)
	}
	policies, err := session.NetClient.ListPolicies(ids...)
	if err != nil {
		return err
	}

	finalPolicies := intersectSlices(policiesTf, policies, func(source, item interface{}) bool {
		policyTf := source.(map[string]interface{})
		policy := item.(cfnetv1.Policy)
		start, end, err := portRangeParse(policyTf["port"].(string))
		if err != nil {
			// this is already validated in validate func
			// so if we have something wrong we are deeply unlucky
			panic(err)
		}
		if start != policy.Destination.Ports.Start || end != policy.Destination.Ports.End {
			return false
		}
		return policyTf["source_app"].(string) == policy.Source.ID &&
			policyTf["destination_app"].(string) == policy.Destination.ID &&
			policyTf["protocol"].(string) == string(policy.Destination.Protocol)
	})
	d.Set("policy", finalPolicies)
	return nil
}

func resourceNetworkPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	old, now := d.GetChange("policy")
	remove, add := getListMapChanges(old, now, func(source, item map[string]interface{}) bool {
		return source["source_app"] == item["source_app"] &&
			source["destination_app"] == item["destination_app"] &&
			source["protocol"] == item["protocol"] &&
			source["port"] == item["port"]
	})
	if len(remove) > 0 {
		err := session.NetClient.RemovePolicies(resourceNetworkPoliciesToPolicies(remove))
		if err != nil {
			return err
		}
	}
	if len(add) > 0 {
		err := session.NetClient.CreatePolicies(resourceNetworkPoliciesToPolicies(add))
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceNetworkPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	policiesTf := getListOfStructs(d.Get("policy"))
	return session.NetClient.RemovePolicies(resourceNetworkPoliciesToPolicies(policiesTf))
}

func resourceNetworkPoliciesToPolicies(policiesTf []map[string]interface{}) []cfnetv1.Policy {
	policies := make([]cfnetv1.Policy, len(policiesTf))
	for i, policyTf := range policiesTf {
		ports := cfnetv1.Ports{}
		start, end, err := portRangeParse(policyTf["port"].(string))
		if err != nil {
			// this is already validated in validate func
			// so if we have something wrong we are deeply unlucky
			panic(err)
		}
		ports.Start = start
		ports.End = end
		policies[i] = cfnetv1.Policy{
			Source: cfnetv1.PolicySource{
				ID: policyTf["source_app"].(string),
			},
			Destination: cfnetv1.PolicyDestination{
				Protocol: cfnetv1.PolicyProtocol(policyTf["protocol"].(string)),
				ID:       policyTf["destination_app"].(string),
				Ports:    ports,
			},
		}
	}
	return policies
}
