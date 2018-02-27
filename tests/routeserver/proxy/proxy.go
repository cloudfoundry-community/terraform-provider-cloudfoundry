package proxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/benlaplanche/cf-basic-auth-route-service/routeserver/utils"
)

type BasicAuthTransport struct {
	transport http.RoundTripper
}

func NewBasicAuthTransport(skipSSLValidation bool) http.RoundTripper {
	return &BasicAuthTransport{
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLValidation},
		},
	}
}

func (b *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	err := checkHeaders(req)
	if err != nil {
		fmt.Printf("Invalid headers. %+v\n", req.Header)
		return nil, err
	}

	url := req.Header.Get("X-CF-Forwarded-Url")

	expectedUsername := "admin"
	expectedPassword := utils.StripAndReverse(url)
	fmt.Printf("Expected password == %s", expectedPassword)

	if !checkAuthorization(expectedUsername, expectedPassword, req) {
		response := &http.Response{
			StatusCode: http.StatusForbidden,
		}

		err := errors.New(fmt.Sprintf("Unauthorized access attempt to %s", url))
		return response, err
	}

	return nil, nil
}

func checkHeaders(r *http.Request) error {
	if r.Header.Get("X-CF-Forwarded-Url") == "" {
		return missingHeaderError("X-CF-Forwarded-Url")
	}

	if r.Header.Get("X-CF-Proxy-Metadata") == "" {
		return missingHeaderError("X-CF-Proxy-Metadata")
	}

	if r.Header.Get("X-CF-Proxy-Signature") == "" {
		return missingHeaderError("X-CF-Proxy-Signature")
	}

	return nil
}

func missingHeaderError(header string) error {
	return errors.New(fmt.Sprintf("Missing expected header: %s", header))
}

func checkAuthorization(expectedUsername string, expectedPassword string, r *http.Request) bool {
	providedUsername, providedPassword, isOk := r.BasicAuth()
	fmt.Printf("Provided username == %s", providedUsername)
	fmt.Printf("Provided password == %s", providedPassword)
	fmt.Printf("Isok? == %s", isOk)
	return isOk && providedUsername == expectedUsername && providedPassword == expectedPassword
}
