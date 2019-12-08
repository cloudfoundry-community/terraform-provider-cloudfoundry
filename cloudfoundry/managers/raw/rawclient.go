package raw

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type rawConnection struct {
	httpClient *http.Client
}

func (c rawConnection) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	response, err := c.httpClient.Do(request.Request)
	if err != nil {
		return err
	}
	passedResponse.HTTPResponse = response
	return nil
}

// RawClientConfig - configuration for RawClient
type RawClientConfig struct {
	DialTimeout       time.Duration
	SkipSSLValidation bool
	ApiEndpoint       string
}

// Raw http client has uaa client authentication to make raw request with golang native api.
type RawClient struct {
	connection  cloudcontroller.Connection
	apiEndpoint string
	wrappers    []ccv3.ConnectionWrapper
}

// NewRawClient -
func NewRawClient(config RawClientConfig, wrappers ...ccv3.ConnectionWrapper) *RawClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.SkipSSLValidation,
			},
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				KeepAlive: 30 * time.Second,
				Timeout:   config.DialTimeout,
			}).DialContext,
		},
	}
	var connection cloudcontroller.Connection = &rawConnection{httpClient}
	for _, wrapper := range wrappers {
		connection = wrapper.Wrap(connection)
	}
	return &RawClient{
		connection:  connection,
		apiEndpoint: strings.TrimSuffix(config.ApiEndpoint, "/"),
		wrappers:    wrappers,
	}
}

// Do - Do the request with given http client and wrappers
func (c RawClient) Do(req *http.Request) (*http.Response, error) {
	resp := &cloudcontroller.Response{}
	err := c.connection.Make(&cloudcontroller.Request{
		Request: req,
	}, resp)
	return resp.HTTPResponse, err
}

// NewRequest - Create a new request with setting api endpoint to the path
func (c RawClient) NewRequest(method, path string, body io.ReadCloser) (*http.Request, error) {
	return http.NewRequest(method, c.apiEndpoint+path, body)
}
