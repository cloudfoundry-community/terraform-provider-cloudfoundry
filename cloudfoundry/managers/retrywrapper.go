package managers

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/router"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
)

// RetryRequest is a wrapper that retries failed requests if they contain a 5XX
// status code.
// copy of wrapper retry request in cli but remove the necessary
// of have a readseeker body (annoying for sending in fullstream)
type RetryRequest struct {
	maxRetries int
	connection cloudcontroller.Connection
}

// NewRetryRequest returns a pointer to a RetryRequest wrapper.
func NewRetryRequest(maxRetries int) *RetryRequest {
	return &RetryRequest{
		maxRetries: maxRetries,
	}
}

// Make retries the request if it comes back with a 5XX status code.
func (retry *RetryRequest) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	var err error
	for i := 0; i < retry.maxRetries+1; i++ {
		err = retry.connection.Make(request, passedResponse)
		if err == nil {
			return nil
		}

		if skipRetry(request.Method, passedResponse.HTTPResponse) {
			break
		}

		if request.Body == nil {
			continue
		}

		// detect if body is ioutil.NopCloser(&bytes.Buffer)
		// if so we reset the content the buffer to be able to redo request with same body
		if reflect.TypeOf(request.Body) == reflect.TypeOf(ioutil.NopCloser) {
			reader := reflect.ValueOf(request.Body).FieldByName("Reader")
			if buf, ok := reader.Interface().(*bytes.Buffer); ok {
				data := buf.Bytes()
				buf.Reset()
				buf.Write(data)
			}
			continue
		}
		// detect if body is implementing interface ReadSeeker
		// if so we go to the beginning of the content to be able to redo request with same body
		if reader, ok := request.Body.(io.ReadSeeker); ok {
			_, resetErr := reader.Seek(0, 0)
			if resetErr != nil {
				if _, ok := resetErr.(ccerror.PipeSeekError); ok {
					return ccerror.PipeSeekError{Err: err}
				}
				return resetErr
			}
			continue
		}
		// if we reach this part, we are not able to know what is inside request body (and be able to resend the same content).
		// This probably cause of full stream send which can be necessary by user.
		// so we return directly the current error
		return err
	}
	return err
}

// Wrap sets the connection in the RetryRequest and returns itself.
func (retry *RetryRequest) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	retry.connection = innerconnection
	return retry
}

func skipRetry(httpMethod string, response *http.Response) bool {
	return httpMethod == http.MethodPost ||
		response != nil &&
			response.StatusCode != http.StatusInternalServerError &&
			response.StatusCode != http.StatusBadGateway &&
			response.StatusCode != http.StatusServiceUnavailable &&
			response.StatusCode != http.StatusGatewayTimeout
}

type retryRequestRouter struct {
	maxRetries int
	connection router.Connection
}

func newRetryRequestRouter(maxRetries int) *retryRequestRouter {
	return &retryRequestRouter{
		maxRetries: maxRetries,
	}
}

func (retry *retryRequestRouter) Make(request *router.Request, passedResponse *router.Response) error {
	var err error
	for i := 0; i < retry.maxRetries+1; i++ {
		err = retry.connection.Make(request, passedResponse)
		if err == nil {
			return nil
		}

		if skipRetry(request.Method, passedResponse.HTTPResponse) && passedResponse.HTTPResponse.StatusCode != http.StatusNotFound {
			break
		}
		if request.Body == nil {
			continue
		}
		// detect if body is ioutil.NopCloser(&bytes.Buffer)
		// if so we reset the content the buffer to be able to redo request with same body
		if reflect.TypeOf(request.Body) == reflect.TypeOf(ioutil.NopCloser) {
			reader := reflect.ValueOf(request.Body).FieldByName("Reader")
			if buf, ok := reader.Interface().(*bytes.Buffer); ok {
				data := buf.Bytes()
				buf.Reset()
				buf.Write(data)
			}
			continue
		}
		// detect if body is implementing interface ReadSeeker
		// if so we go to the beginning of the content to be able to redo request with same body
		if reader, ok := request.Body.(io.ReadSeeker); ok {
			_, resetErr := reader.Seek(0, 0)
			if resetErr != nil {
				if _, ok := resetErr.(ccerror.PipeSeekError); ok {
					return ccerror.PipeSeekError{Err: err}
				}
				return resetErr
			}
			continue
		}
		// if we reach this part, we are not able to know what is inside request body (and be able to resend the same content).
		// This probably cause of full stream send which can be necessary by user.
		// so we return directly the current error
		return err
	}
	return err
}

// Wrap sets the connection in the RetryRequest and returns itself.
func (retry *retryRequestRouter) Wrap(innerconnection router.Connection) router.Connection {
	retry.connection = innerconnection
	return retry
}
