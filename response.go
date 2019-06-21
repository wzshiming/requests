package requests

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// Response is an object represents executed request and its values.
type Response struct {
	rawResponse *http.Response
	body        []byte
	location    *url.URL
	method      string
	sendAt      time.Time
	recvAt      time.Time
}

func newResponse(resp *http.Response) (*Response, error) {
	r := &Response{
		rawResponse: resp,
	}
	err := r.process()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Response) init(sendAt time.Time, method string, u *url.URL) {
	if r.location != nil {
		r.location = u
	}
	r.method = method
	r.sendAt = sendAt
	r.recvAt = time.Now()
}

// Location return request url.
func (r *Response) Location() *url.URL {
	return r.location
}

// WriteFile is writes the response body to file.
func (r *Response) WriteFile(file string) error {
	return ioutil.WriteFile(file, r.body, 0666)
}

// Body returns HTTP response as []byte array for the executed request.
func (r *Response) Body() []byte {
	return r.body
}

// ContentType returns HTTP response content type
func (r *Response) ContentType() string {
	return r.rawResponse.Header.Get(HeaderContentType)
}

// Status returns the HTTP status string for the executed request.
func (r *Response) Status() string {
	if r.rawResponse == nil {
		return "from cache"
	}
	return r.rawResponse.Status
}

// StatusCode returns the HTTP status code for the executed request.
func (r *Response) StatusCode() int {
	return r.rawResponse.StatusCode
}

// Header returns the response headers
func (r *Response) Header() http.Header {
	return r.rawResponse.Header
}

// Cookies to access all the response cookies
func (r *Response) Cookies() []*http.Cookie {
	if r.rawResponse == nil {
		return nil
	}
	return r.rawResponse.Cookies()
}

// Time returns the time of HTTP response time that from request we sent and received a request.
func (r *Response) Time() time.Duration {
	return r.recvAt.Sub(r.sendAt)
}

// RecvAt returns when response got recv from server for the request.
func (r *Response) RecvAt() time.Time {
	return r.recvAt
}

// SendAt returns when response got send from server for the request.
func (r *Response) SendAt() time.Time {
	return r.sendAt
}

// Size returns the HTTP response size in bytes.
func (r *Response) Size() int {
	return len(r.body)
}

// RawBody returns the HTTP raw response body.
func (r *Response) RawBody() io.Reader {
	return bytes.NewReader(r.body)
}

// String returns the HTTP response basic information
func (r *Response) String() string {
	return fmt.Sprintf("%s %s %d %d %s", r.method, r.location.String(), r.StatusCode(), r.Size(), r.Time())
}

// Message returns the HTTP response all information
func (r *Response) Message() string {
	return r.message(true)
}

// MessageHead returns the HTTP response header information
func (r *Response) MessageHead() string {
	return r.message(false)
}

func (r *Response) message(body bool) string {
	b, err := httputil.DumpResponse(r.RawResponse(), body)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (r *Response) RawResponse() *http.Response {
	resp := r.rawResponse
	if resp == nil {
		return nil
	}
	resp.Body = ioutil.NopCloser(r.RawBody())
	return resp
}

func (r *Response) process() (err error) {
	resp := r.rawResponse
	if u, err := resp.Location(); err == nil {
		r.location = u
	}
	body := TryCharset(resp.Body, r.ContentType())
	r.body, _ = ioutil.ReadAll(body)
	if err := resp.Body.Close(); err != nil {
		return err
	}
	resp.Body = nil
	return nil
}

func (r *Response) MarshalText() ([]byte, error) {
	return MarshalResponse(r.RawResponse())
}

func (r *Response) UnarshalText(data []byte) error {
	resp, err := UnmarshalResponse(data)
	if err != nil {
		return err
	}
	r.rawResponse = resp
	return r.process()
}
