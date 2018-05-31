package requests

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// Response is an object represents executed request and its values.
type Response struct {
	request     *Request
	rawResponse *http.Response
	body        []byte
	recvAt      time.Time
}

// Body method returns HTTP response as []byte array for the executed request.
func (r *Response) Body() []byte {
	return r.body
}

// Status method returns the HTTP status string for the executed request.
func (r *Response) Status() string {
	return r.rawResponse.Status
}

// StatusCode method returns the HTTP status code for the executed request.
func (r *Response) StatusCode() int {
	return r.rawResponse.StatusCode
}

// Header method returns the response headers
func (r *Response) Header() http.Header {
	return r.rawResponse.Header
}

// Cookies method to access all the response cookies
func (r *Response) Cookies() []*http.Cookie {
	return r.rawResponse.Cookies()
}

// Time method returns the time of HTTP response time that from request we sent and received a request.
func (r *Response) Time() time.Duration {
	return r.recvAt.Sub(r.request.sendAt)
}

// ReceivedAt method returns when response got recevied from server for the request.
func (r *Response) ReceivedAt() time.Time {
	return r.recvAt
}

// Size method returns the HTTP response size in bytes.
func (r *Response) Size() int {
	return len(r.body)
}

// RawBody method exposes the HTTP raw response body.
func (r *Response) RawBody() io.Reader {
	return bytes.NewReader(r.body)
}
