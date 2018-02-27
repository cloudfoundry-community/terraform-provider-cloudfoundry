package broker_test

import (
	"github.com/benlaplanche/cf-basic-auth-route-service/servicebroker/broker"
	"github.com/benlaplanche/cf-basic-auth-route-service/servicebroker/config"
	"github.com/pivotal-cf/brokerapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Auth Service Broker", func() {
	var basicAuthBroker *broker.BasicAuthBroker
	var basicAuthService *brokerapi.Service
	var basicAuthServicePlan *brokerapi.ServicePlan

	BeforeEach(func() {
		basicAuthBroker = &broker.BasicAuthBroker{
			Config: config.Config{
				config.BrokerConfiguration{
					RouteServiceURL: "https://my-route-service.com",
					BrokerUserName:  "admin",
					BrokerPassword:  "letmein"}}}
		basicAuthService = &basicAuthBroker.Services()[0]
		basicAuthServicePlan = &basicAuthService.Plans[0]
	})

	Describe(".Services", func() {
		It("returns a single service", func() {
			services := basicAuthBroker.Services()
			Expect(len(services)).To(Equal(1))
		})

		It("returns the correct service id", func() {
			Expect(basicAuthService.ID).To(Equal("6a97b5b8-1d1f-44bc-98ae-01d8d1047555"))
		})

		It("returns the correct service name", func() {
			Expect(basicAuthService.Name).To(Equal("p-basic-auth"))
		})

		It("returns the correct description", func() {
			Expect(basicAuthService.Description).To(Equal("Protect applications with basic authentication in the routing path"))
		})

		It("returns the correct tags", func() {
			Expect(basicAuthService.Tags).To(Equal([]string{"route-service", "basic-auth"}))
		})

		It("returns the service as bindable", func() {
			Expect(basicAuthService.Bindable).To(BeTrue())
		})

		It("returns the service plan as not updateable", func() {
			Expect(basicAuthService.PlanUpdatable).To(BeFalse())
		})

		It("requires route forwarding", func() {
			Expect(basicAuthService.Requires).To(Equal([]brokerapi.RequiredPermission{brokerapi.PermissionRouteForwarding}))
		})

		Describe(".Plans", func() {
			It("returns a single plan", func() {
				plans := basicAuthService.Plans
				Expect(len(plans)).To(Equal(1))
			})

			It("returns the correct plan ID", func() {
				Expect(basicAuthServicePlan.ID).To(Equal("7becb74f-ce9d-4f52-87a2-50cc1b2b4b8f"))
			})

			It("returns the correct plan name", func() {
				Expect(basicAuthServicePlan.Name).To(Equal("reverse-name"))
			})

			It("returns the correct plan description", func() {
				Expect(basicAuthServicePlan.Description).To(Equal("The password is the url before the dot (.) in reverse, username is admin"))
			})

			It("returns the correct plan display name", func() {
				Expect(basicAuthServicePlan.Metadata.DisplayName).To(Equal("Reverse Name"))
			})

			It("returns the correct plan bullet points", func() {
				Expect(basicAuthServicePlan.Metadata.Bullets).To(Equal([]string{"Routing service", "Provides basic authentication", "Password is the application URL before the dot (.) in reverse", "Username is admin"}))
			})
		})

		Describe(".Metadata", func() {
			It("returns the correct service metadata display name", func() {
				Expect(basicAuthService.Metadata.DisplayName).To(Equal("Basic Auth"))
			})

			It("returns the correct service metadata support url", func() {
				Expect(basicAuthService.Metadata.SupportUrl).To(Equal("https://github.com/benlaplanche/cf-basic-auth-route-service/"))
			})

			It("returns the correct service metadata documentation url", func() {
				Expect(basicAuthService.Metadata.DocumentationUrl).To(Equal("https://github.com/benlaplanche/cf-basic-auth-route-service/"))
			})

			It("returns the correct service metadata provider display name", func() {
				Expect(basicAuthService.Metadata.ProviderDisplayName).To(Equal("Ben Laplanche"))
			})

			It("returns the correct service metadata long description", func() {
				Expect(basicAuthService.Metadata.LongDescription).To(Equal("Protect access to your application with this basic auth routing service"))
			})
		})
	})
	Describe(".Provision", func() {
		It("does not return an error", func() {
			_, err := basicAuthBroker.Provision("instance-id", brokerapi.ProvisionDetails{}, false)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe(".Bind", func() {
		It("does not return an error", func() {
			_, err := basicAuthBroker.Bind("instance-id", "binding-id", brokerapi.BindDetails{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns the correct route service url", func() {
			credentials, err := basicAuthBroker.Bind("instance-id", "binding-id", brokerapi.BindDetails{})
			Expect(err).ToNot(HaveOccurred())

			expectedResult := brokerapi.Binding{
				Credentials:     "",
				RouteServiceURL: "https://my-route-service.com",
			}

			Expect(credentials).To(Equal(expectedResult))
		})
	})

	Describe(".Unbind", func() {
		It("does not return an error", func() {
			err := basicAuthBroker.Unbind("instance-id", "binding-id", brokerapi.UnbindDetails{})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe(".Deprovision", func() {
		It("does not return an error", func() {
			_, err := basicAuthBroker.Deprovision("instance-id", brokerapi.DeprovisionDetails{}, false)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe(".LastOperation", func() {
		It("does not return an error", func() {
			_, err := basicAuthBroker.LastOperation("instance-id")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe(".Update", func() {
		It("does not return an error", func() {
			_, err := basicAuthBroker.Update("instance-id", brokerapi.UpdateDetails{}, false)
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
