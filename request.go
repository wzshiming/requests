package requests

import (
	"context"
	"io"
	"net/http"
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

// Head method does HEAD HTTP request.
func (r *Request) Head(url string) (*Response, error) {
	return r.Do(MethodHead, url)
}

// Get method does GET HTTP request.
func (r *Request) Get(url string) (*Response, error) {
	return r.Do(MethodGet, url)
}

// Post method does POST HTTP request.
func (r *Request) Post(url string) (*Response, error) {
	return r.Do(MethodPost, url)
}

// Put method does PUT HTTP request.
func (r *Request) Put(url string) (*Response, error) {
	return r.Do(MethodPut, url)
}

// Delete method does DELETE HTTP request.
func (r *Request) Delete(url string) (*Response, error) {
	return r.Do(MethodDelete, url)
}

// Options method does OPTIONS HTTP request.
func (r *Request) Options(url string) (*Response, error) {
	return r.Do(MethodOptions, url)
}

// Trace method does TRACE HTTP request. It's defined in section 4.3.2 of RFC7231.
func (r *Request) Trace(url string) (*Response, error) {
	return r.Do(MethodTrace, url)
}

// Patch method does PATCH HTTP request. It's defined in section 2 of RFC5789.
func (r *Request) Patch(url string) (*Response, error) {
	return r.Do(MethodPatch, url)
}

// Do method performs the HTTP request
func (r *Request) Do(method, rawurl string) (*Response, error) {
	r = r.Clone()
	r.method = method
	r.SetURL(rawurl)

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

	req, err := http.NewRequest(method, r.baseURL.String(), r.body)
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
