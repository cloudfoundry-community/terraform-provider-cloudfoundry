package utils_test

import (
	"github.com/benlaplanche/cf-basic-auth-route-service/routeserver/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {

	Describe(".Strip Special characters and reverse the string", func() {

		It("without special characters or numbers", func() {
			result := utils.StripAndReverse("https://myapp.pcf.io")
			Expect(result).To(Equal("ppaym"))
		})

		It("with a special character", func() {
			result := utils.StripAndReverse("http://my-app.pcf.io")
			Expect(result).To(Equal("ppaym"))
		})

		It("with a special character and number", func() {
			result := utils.StripAndReverse("https://my-app-1.pcf.io")
			Expect(result).To(Equal("1ppaym"))
		})

		It("with an ip address", func() {
			result := utils.StripAndReverse("http://127.0.0.1")
			Expect(result).To(Equal("721"))
		})

		It("with an ip address and port", func() {
			result := utils.StripAndReverse("https://127.0.0.1:8080")
			Expect(result).To(Equal("721"))
		})

	})
})
