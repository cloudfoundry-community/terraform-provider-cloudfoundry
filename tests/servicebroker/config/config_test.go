package config_test

import (
	"path"
	"path/filepath"

	"github.com/benlaplanche/cf-basic-auth-route-service/servicebroker/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Loading the broker Config file ", func() {
	Describe("Parse the broker config file", func() {

		var (
			configPath     string
			brokerConfig   config.Config
			parseConfigErr error
		)

		BeforeEach(func() {
			configPath = "test_config.yml"
		})

		JustBeforeEach(func() {
			path, err := filepath.Abs(path.Join("assets", configPath))
			Expect(err).ToNot(HaveOccurred())
			brokerConfig, parseConfigErr = config.ParseConfig(path)
		})

		Context("when the config is valid", func() {
			It("does not error", func() {
				Expect(parseConfigErr).ToNot(HaveOccurred())
			})

			It("returns the correct route service url", func() {
				Expect(brokerConfig.BrokerConfiguration.RouteServiceURL).To(Equal("https://my-route-service.com"))
			})

			It("returns the correct broker user name", func() {
				Expect(brokerConfig.BrokerConfiguration.BrokerUserName).To(Equal("admin"))
			})

			It("returns the correct broker password", func() {
				Expect(brokerConfig.BrokerConfiguration.BrokerPassword).To(Equal("letmein"))
			})
		})
	})
})
