package proxy_test

import (
	. "github.com/benlaplanche/cf-basic-auth-route-service/routeserver/proxy"
	utils "github.com/benlaplanche/cf-basic-auth-route-service/routeserver/utils"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Proxy", func() {
	var (
		transport        http.RoundTripper
		req              *http.Request
		helloworldServer *ghttp.Server
	)

	BeforeEach(func() {
		helloworldServer = ghttp.NewServer()
		helloworldServer.AppendHandlers(ghttp.RespondWith(200, []byte("Hello World! I'm protected with basic authentication")))

		req, _ = http.NewRequest("GET", helloworldServer.URL(), nil)
		transport = NewBasicAuthTransport(true)

		req.Header.Add("X-CF-Forwarded-Url", string(helloworldServer.URL()))
		req.Header.Add("X-CF-Proxy-Metadata", "metadata-goes-here")
		req.Header.Add("X-CF-Proxy-Signature", "signature-goes-here")
	})

	Context("The added gorouter headers are missing", func() {
		It("returns an error response when there is no forwarded url", func() {
			req.Header.Del("X-CF-Forwarded-Url")

			res, err := transport.RoundTrip(req)
			Expect(res).To(BeNil())
			Expect(err).ToNot(BeNil())
		})

		It("returns an error response when there is no proxy metadata", func() {
			req.Header.Del("X-CF-Proxy-Metadata")

			res, err := transport.RoundTrip(req)
			Expect(res).To(BeNil())
			Expect(err).ToNot(BeNil())
		})

		It("returns an error response when there is no proxy signature", func() {
			req.Header.Del("X-CF-Proxy-Signature")

			res, err := transport.RoundTrip(req)
			Expect(res).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})

	Context("The added gorouter headers are present", func() {

		Context("Without valid basic authentication username and password", func() {
			It("returns an error", func() {
				req.SetBasicAuth("admin", "invalid-password")

				res, err := transport.RoundTrip(req)
				Expect(err).ToNot(BeNil())
				Expect(res.StatusCode).To(Equal(403))
				Expect(helloworldServer.ReceivedRequests()).To(HaveLen(0))
			})
		})

		Context("With valid basic authentication username and password", func() {

			JustBeforeEach(func() {
				password := utils.StripAndReverse(helloworldServer.URL())
				req.SetBasicAuth("admin", password)
			})

			It("returns the correct HTTP Status code", func() {
				res, err := transport.RoundTrip(req)

				Expect(err).To(BeNil())
				Expect(res.StatusCode).To(Equal(200))
			})

			It("returns the expected http body test", func() {
				res, err := transport.RoundTrip(req)

				Expect(err).To(BeNil())
				Expect(res.Body).To(Equal("Hello World! I'm protected with basic authentication"))
			})
		})

	})
})
