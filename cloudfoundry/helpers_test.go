package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"net/http"
)

type HelpersTest struct {
	session *managers.Session
}

func (h HelpersTest) ForceDeleteServiceBroker(serviceBrokerID string) (err error) {

	services, _, err := h.session.ClientV2.GetServices(ccv2.Filter{
		Values:   []string{serviceBrokerID},
		Operator: constant.EqualOperator,
		Type:     constant.ServiceBrokerGUIDFilter,
	})
	if err != nil {
		return err
	}

	for _, s := range services {
		plans, _, err := h.session.ClientV2.GetServicePlans(ccv2.Filter{
			Values:   []string{s.GUID},
			Operator: constant.EqualOperator,
			Type:     constant.ServiceGUIDFilter,
		})
		if err != nil {
			return err
		}
		for _, sp := range plans {
			instances, _, err := h.session.ClientV2.GetServiceInstances(ccv2.Filter{
				Values:   []string{sp.GUID},
				Operator: constant.EqualOperator,
				Type:     constant.ServicePlanGUIDFilter,
			})
			if err != nil {
				return err
			}
			for _, i := range instances {
				req, err := h.session.RawClient.NewRequest(
					http.MethodDelete,
					fmt.Sprintf("/v2/service_instances/%s?purge=true", i.GUID),
					nil,
				)
				if err != nil {
					panic(err)
				}
				resp, err := h.session.RawClient.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode >= 400 || resp.StatusCode < 200 {
					return fmt.Errorf(resp.Status)
				}
			}
		}
	}

	_, err = h.session.ClientV2.DeleteServiceBroker(serviceBrokerID)
	return err
}
