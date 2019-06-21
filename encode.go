package requests

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httputil"
)

// MarshalRequest returns text of the request
func MarshalRequest(req *http.Request) ([]byte, error) {
	return httputil.DumpRequest(req, true)
}

// UnmarshalRequest reads and returns an HTTP request from data.
func UnmarshalRequest(data []byte) (req *http.Request, err error) {
	return http.ReadRequest(bufio.NewReader(bytes.NewBuffer(data)))
}

// MarshalRequest returns text of the request
func MarshalResponse(resp *http.Response) ([]byte, error) {
	return httputil.DumpResponse(resp, true)
}

// UnmarshalResponse reads and returns an HTTP response from data.
func UnmarshalResponse(data []byte) (resp *http.Response, err error) {
	return http.ReadResponse(bufio.NewReader(bytes.NewBuffer(data)), nil)
}
