package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"time"
)

// Request type is used to compose and send individual request from client
type Request struct {
	baseURL         *url.URL
	method          string
	headerParam     []*paramPair
	queryParam      []*paramPair
	pathParam       []*paramPair
	formParam       []*paramPair
	multiFiles      []*multiFile
	body            io.Reader
	sendAt          time.Time
	rawRequest      *http.Request
	client          *Client
	ctx             context.Context
	discardResponse bool
}

func newRequest(c *Client) *Request {
	return &Request{
		client: c,
		method: MethodGet,
	}
}

// Clone method clone the request
func (r *Request) Clone() *Request {
	n := &Request{}
	*n = *r
	return n
}

// SetBaseURL method is to set URL in the client instance.
func (r *Request) SetBaseURL(u *url.URL) *Request {
	if u == nil {
		r.baseURL = nil
		return r
	}
	r.baseURL = u
	if user := r.baseURL.User; user != nil {
		pwd, _ := user.Password()
		r.SetBasicAuth(user.Username(), pwd)
		r.baseURL.User = nil
	}
	return r
}

// SetURL method is to set URL in the client instance.
func (r *Request) SetURL(rawurl string) *Request {
	var nu *url.URL
	var err error
	if r.baseURL == nil {
		nu, err = url.Parse(rawurl)
	} else {
		nu, err = r.baseURL.Parse(rawurl)
	}
	if err != nil {
		r.client.printError(err)
	}
	r.SetBaseURL(nu)
	return r
}

// SetContext method sets the context.Context for current Request.
func (r *Request) SetContext(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

func (r *Request) withContext() {
	if r.ctx != nil {
		r.rawRequest = r.rawRequest.WithContext(r.ctx)
	}
}

func (r *Request) isCancelled() bool {
	if r.ctx != nil {
		if r.ctx.Err() != nil {
			return true
		}
	}
	return false
}

// SetHeader method is to set a single header field and its value in the current request.
func (r *Request) SetHeader(param, value string) *Request {
	param = textproto.CanonicalMIMEHeaderKey(param)
	r.headerParam = append(r.headerParam, &paramPair{
		Param: param,
		Value: value,
	})
	return r
}

// SetPath method sets single path parameter and its value in the current request.
func (r *Request) SetPath(param, value string) *Request {
	r.pathParam = append(r.pathParam, &paramPair{
		Param: param,
		Value: value,
	})
	return r
}

// SetQuery method sets single parameter and its value in the current request.
func (r *Request) SetQuery(param, value string) *Request {
	r.queryParam = append(r.queryParam, &paramPair{
		Param: param,
		Value: value,
	})
	return r
}

// SetForm method appends multiple form parameters with multi-value
func (r *Request) SetForm(param, value string) *Request {
	r.formParam = append(r.formParam, &paramPair{
		Param: param,
		Value: value,
	})
	return r
}

// SetFile method is to set custom data using io.Reader for multipart upload.
func (r *Request) SetFile(param, fileName, contentType string, reader io.Reader) *Request {
	r.multiFiles = append(r.multiFiles, &multiFile{
		Param:       param,
		FileName:    fileName,
		ContentType: contentType,
		Reader:      reader,
	})
	return r
}

// SetJSON method sets the data encoded by JSON to the request body.
func (r *Request) SetJSON(i interface{}) *Request {
	data, err := json.Marshal(i)
	if err != nil {
		r.client.printError(err)
	}
	r.body = bytes.NewReader(data)
	r.SetContentType(MimeJSON)
	return r
}

// SetXML method sets the data encoded by XML to the request body.
func (r *Request) SetXML(i interface{}) *Request {
	data, err := xml.Marshal(i)
	if err != nil {
		r.client.printError(err)
	}
	r.body = bytes.NewReader(data)
	r.SetContentType(MimeXML)
	return r
}

// SetBody method sets the request body for the request.
func (r *Request) SetBody(body io.Reader) *Request {
	r.body = body
	return r
}

// SetContentType method sets the content type header in the HTTP request.
func (r *Request) SetContentType(contentType string) *Request {
	r.SetHeader(HeaderContentType, contentType)
	return r
}

// SetBasicAuth method sets the basic authentication header in the HTTP request.
func (r *Request) SetBasicAuth(username, password string) *Request {
	r.SetHeader(HeaderAuthorization, "Basic "+basicAuth(username, password))
	return r
}

// SetAuthToken method sets bearer auth token header in the HTTP request.
func (r *Request) SetAuthToken(token string) *Request {
	r.SetHeader(HeaderAuthorization, "Bearer "+token)
	return r
}

// SetUserAgent method sets user agent header in the HTTP request.
func (r *Request) SetUserAgent(ua string) *Request {
	r.SetHeader(HeaderUserAgent, ua)
	return r
}

// SetDiscardResponse method unread the response body.
func (r *Request) SetDiscardResponse(discard bool) *Request {
	r.discardResponse = discard
	return r
}

// SetMethod method sets method in the HTTP request.
func (r *Request) SetMethod(method string) *Request {
	r.method = method
	return r
}

// Head method does HEAD HTTP request.
func (r *Request) Head(url string) (*Response, error) {
	return r.SetMethod(MethodHead).SetURL(url).Do()
}

// Get method does GET HTTP request.
func (r *Request) Get(url string) (*Response, error) {
	return r.SetMethod(MethodGet).SetURL(url).Do()
}

// Post method does POST HTTP request.
func (r *Request) Post(url string) (*Response, error) {
	return r.SetMethod(MethodPost).SetURL(url).Do()
}

// Put method does PUT HTTP request.
func (r *Request) Put(url string) (*Response, error) {
	return r.SetMethod(MethodPut).SetURL(url).Do()
}

// Delete method does DELETE HTTP request.
func (r *Request) Delete(url string) (*Response, error) {
	return r.SetMethod(MethodDelete).SetURL(url).Do()
}

// Options method does OPTIONS HTTP request.
func (r *Request) Options(url string) (*Response, error) {
	return r.SetMethod(MethodOptions).SetURL(url).Do()
}

// Trace method does TRACE HTTP request.
func (r *Request) Trace(url string) (*Response, error) {
	return r.SetMethod(MethodTrace).SetURL(url).Do()
}

// Patch method does PATCH HTTP request.
func (r *Request) Patch(url string) (*Response, error) {
	return r.SetMethod(MethodPatch).SetURL(url).Do()
}

// Do method performs the HTTP request
func (r *Request) Do() (*Response, error) {
	r = r.Clone()

	// fill path
	if len(r.pathParam) != 0 {
		err := toPath(r.baseURL, r.pathParam)
		if err != nil {
			return nil, err
		}
	}

	// fill query
	if len(r.queryParam) != 0 {
		err := toQuery(r.baseURL, r.queryParam)
		if err != nil {
			return nil, err
		}
	}

	if r.body == nil {
		if len(r.multiFiles) != 0 { // fill multpair
			body, contentType, err := toMulti(r.formParam, r.multiFiles)
			if err != nil {
				return nil, err
			}
			r.SetContentType(contentType)
			r.body = body
		} else { // fill form
			body, err := toForm(r.formParam)
			if err != nil {
				return nil, err
			}
			r.SetContentType(MimeURLEncoded)
			r.body = body
		}
	}

	req, err := http.NewRequest(r.method, r.baseURL.String(), r.body)
	if err != nil {
		return nil, err
	}

	// fill header
	if len(r.headerParam) != 0 {
		err := toHeader(req, r.headerParam)
		if err != nil {
			return nil, err
		}
	}

	r.rawRequest = req
	return r.client.do(r)
}
