package requests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
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

func (r *Response) String() string {
	return fmt.Sprintf("%s %s %d %d %s", r.request.method, r.request.baseURL.String(), r.StatusCode(), r.Size(), r.Time())
}

func (r *Response) Message() string {
	return r.message(true)
}

func (r *Response) MessageHead() string {
	return r.message(false)
}

func (r *Response) message(body bool) string {
	b, err := httputil.DumpResponse(r.rawResponse, false)
	if err != nil {
		return err.Error()
	}
	if body {
		b = append(b, r.Body()...)
	}
	return string(b)
}
