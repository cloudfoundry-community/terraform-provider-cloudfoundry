package cloudcontroller

import "net/http"

// Response represents a Cloud Controller response object.
type Response struct {
	// DecodeJSONResponseInto represents the resource entity type that is
	// expected in the response JSON.
	DecodeJSONResponseInto interface{}

	// RawResponse represents the response body.
	RawResponse []byte

	// Warnings represents warnings parsed from the custom warnings headers of a
	// Cloud Controller response.
	Warnings []string

	// HTTPResponse represents the HTTP response object.
	HTTPResponse *http.Response

	// ResourceLocationURL represents the Location header value
	ResourceLocationURL string
}

func (r *Response) reset() {
	r.RawResponse = []byte{}
	r.Warnings = []string{}
	r.HTTPResponse = nil
}
