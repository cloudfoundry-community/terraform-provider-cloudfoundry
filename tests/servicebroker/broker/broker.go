package broker

import (
	"github.com/benlaplanche/cf-basic-auth-route-service/servicebroker/config"
	"github.com/pivotal-cf/brokerapi"
)

type BasicAuthBroker struct {
	Config config.Config
}

func (basicAuthBroker *BasicAuthBroker) Services() []brokerapi.Service {
	return []brokerapi.Service{
		brokerapi.Service{
			ID:            "6a97b5b8-1d1f-44bc-98ae-01d8d1047555",
			Name:          "p-basic-auth",
			Description:   "Protect applications with basic authentication in the routing path",
			Bindable:      true,
			Tags:          []string{"route-service", "basic-auth"},
			PlanUpdatable: false,
			Requires:      []brokerapi.RequiredPermission{brokerapi.PermissionRouteForwarding},
			Plans: []brokerapi.ServicePlan{
				brokerapi.ServicePlan{
					ID:          "7becb74f-ce9d-4f52-87a2-50cc1b2b4b8f",
					Name:        "reverse-name",
					Description: "The password is the url before the dot (.) in reverse, username is admin",
					Metadata: &brokerapi.ServicePlanMetadata{
						DisplayName: "Reverse Name",
						Bullets:     []string{"Routing service", "Provides basic authentication", "Password is the application URL before the dot (.) in reverse", "Username is admin"},
					},
				},
			},
			Metadata: &brokerapi.ServiceMetadata{
				DisplayName:         "Basic Auth",
				SupportUrl:          "https://github.com/benlaplanche/cf-basic-auth-route-service/",
				DocumentationUrl:    "https://github.com/benlaplanche/cf-basic-auth-route-service/",
				ProviderDisplayName: "Ben Laplanche",
				LongDescription:     "Protect access to your application with this basic auth routing service",
			},
		},
	}
}

func (basicAuthBroker *BasicAuthBroker) Provision(instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	return brokerapi.ProvisionedServiceSpec{}, nil
}

func (basicAuthBroker *BasicAuthBroker) Deprovision(instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {
	return false, nil
}

func (basicAuthBroker *BasicAuthBroker) Bind(instanceID string, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	return brokerapi.Binding{
			Credentials:     "",
			RouteServiceURL: basicAuthBroker.Config.BrokerConfiguration.RouteServiceURL},
		nil
}

func (basicAuthBroker *BasicAuthBroker) Unbind(instanceID string, bindingID string, details brokerapi.UnbindDetails) error {
	return nil
}

func (basicAuthBroker *BasicAuthBroker) LastOperation(instanceID string) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}

func (basicAuthBroker *BasicAuthBroker) Update(instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.IsAsync, error) {
	return false, nil
}
