package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"strings"
	"time"
)

// Request type is used to compose and send individual request from client
type Request struct {
	baseURL         *url.URL
	method          string
	headerParam     paramPairs
	queryParam      paramPairs
	pathParam       paramPairs
	formParam       paramPairs
	multiFiles      multiFiles
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
	if n.baseURL != nil {
		bu := *n.baseURL
		n.baseURL = &bu
	}
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

	if r.baseURL.RawQuery != "" {
		qs, _ := url.ParseQuery(r.baseURL.RawQuery)
		for k, v := range qs {
			for _, v := range v {
				r.AddQuery(k, v)
			}
		}
		r.baseURL.RawQuery = ""
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

// SetTimeout method sets the timeout for current Request.
func (r *Request) SetTimeout(timeout time.Duration) *Request {
	return r.SetDeadline(time.Now().Add(timeout))
}

// SetDeadline method sets the deadline for current Request.
func (r *Request) SetDeadline(d time.Time) *Request {
	if r.ctx == nil {
		r.ctx = context.TODO()
	}
	r.ctx, _ = context.WithDeadline(r.ctx, d)
	return r
}

func (r *Request) withContext() {
	if r.ctx != nil {
		r.rawRequest = r.rawRequest.WithContext(r.ctx)
	}
}

func (r *Request) isCancelled() bool {
	return r.ctx != nil && r.ctx.Err() != nil
}

// SetHeader method is to sets a single header field and its value in the current request.
func (r *Request) SetHeader(param, value string) *Request {
	param = textproto.CanonicalMIMEHeaderKey(param)
	r.headerParam.AddReplace(param, value)
	return r
}

// AddHeader method is to adds a single header field and its value in the current request.
func (r *Request) AddHeader(param, value string) *Request {
	param = textproto.CanonicalMIMEHeaderKey(param)
	r.headerParam.Add(param, value)
	return r
}

// AddHeaderIfNot method is to adds a single header field and its value in the current request if not.
func (r *Request) AddHeaderIfNot(param, value string) *Request {
	param = textproto.CanonicalMIMEHeaderKey(param)
	r.headerParam.AddNoRepeat(param, value)
	return r
}

// SetPath method sets single path parameter and its value in the current request.
func (r *Request) SetPath(param, value string) *Request {
	r.pathParam.AddReplace(param, value)
	return r
}

// AddPathIfNot method is to adds a single path parameter and its value in the current request if not.
func (r *Request) AddPathIfNot(param, value string) *Request {
	r.pathParam.AddNoRepeat(param, value)
	return r
}

// SetQuery method sets single query parameter and its value in the current request.
func (r *Request) SetQuery(param, value string) *Request {
	r.queryParam.AddReplace(param, value)
	return r
}

// AddQuery method is to adds a single query field and its value in the current request.
func (r *Request) AddQuery(param, value string) *Request {
	r.queryParam.Add(param, value)
	return r
}

// AddQueryIfNot method is to adds a single query field and its value in the current request if not.
func (r *Request) AddQueryIfNot(param, value string) *Request {
	r.queryParam.AddNoRepeat(param, value)
	return r
}

// SetForm method appends multiple form parameters with multi-value
func (r *Request) SetForm(param, value string) *Request {
	r.formParam.AddReplace(param, value)
	return r
}

// AddForm method is to adds a single from field and its value in the current request.
func (r *Request) AddForm(param, value string) *Request {
	r.formParam.Add(param, value)
	return r
}

// AddFormIfNot method is to adds a single from field and its value in the current request if not.
func (r *Request) AddFormIfNot(param, value string) *Request {
	r.formParam.AddNoRepeat(param, value)
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
		return r
	}
	r.body = bytes.NewReader(data)
	r.AddHeaderIfNot(HeaderContentType, MimeJSON)
	return r
}

// SetXML method sets the data encoded by XML to the request body.
func (r *Request) SetXML(i interface{}) *Request {
	data, err := xml.Marshal(i)
	if err != nil {
		r.client.printError(err)
		return r
	}
	r.body = bytes.NewReader(data)
	r.AddHeaderIfNot(HeaderContentType, MimeXML)
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
	r.method = strings.ToUpper(method)
	return r
}

// Head method does HEAD HTTP request.
func (r *Request) Head(url string) (*Response, error) {
	return r.Clone().SetMethod(MethodHead).SetURL(url).do()
}

// Get method does GET HTTP request.
func (r *Request) Get(url string) (*Response, error) {
	return r.Clone().SetMethod(MethodGet).SetURL(url).do()
}

// Post method does POST HTTP request.
func (r *Request) Post(url string) (*Response, error) {
	return r.Clone().SetMethod(MethodPost).SetURL(url).do()
}

// Put method does PUT HTTP request.
func (r *Request) Put(url string) (*Response, error) {
	return r.Clone().SetMethod(MethodPut).SetURL(url).do()
}

// Delete method does DELETE HTTP request.
func (r *Request) Delete(url string) (*Response, error) {
	return r.Clone().SetMethod(MethodDelete).SetURL(url).do()
}

// Options method does OPTIONS HTTP request.
func (r *Request) Options(url string) (*Response, error) {
	return r.Clone().SetMethod(MethodOptions).SetURL(url).do()
}

// Trace method does TRACE HTTP request.
func (r *Request) Trace(url string) (*Response, error) {
	return r.Clone().SetMethod(MethodTrace).SetURL(url).do()
}

// Patch method does PATCH HTTP request.
func (r *Request) Patch(url string) (*Response, error) {
	return r.Clone().SetMethod(MethodPatch).SetURL(url).do()
}

// Do method performs the HTTP request
func (r *Request) Do() (*Response, error) {
	return r.Clone().do()
}

func (r *Request) do() (*Response, error) {
	_, err := r.fill()
	if err != nil {
		return nil, err
	}
	r.withContext()
	return r.client.do(r)
}

func (r *Request) fill() (*http.Request, error) {
	if r.rawRequest != nil {
		return r.rawRequest, nil
	}

	// fill path
	if len(r.pathParam) != 0 {
		path, err := toPath(r.baseURL.Path, r.pathParam)
		if err != nil {
			return nil, err
		}
		r.baseURL.Path = path
	}

	// fill query
	if len(r.queryParam) != 0 {
		rq, err := toQuery(r.baseURL.RawQuery, r.queryParam)
		if err != nil {
			return nil, err
		}
		r.baseURL.RawQuery = rq
	}

	if r.body == nil {
		if len(r.multiFiles) != 0 { // fill multpair
			body, contentType, err := toMulti(r.formParam, r.multiFiles)
			if err != nil {
				return nil, err
			}
			r.AddHeaderIfNot(HeaderContentType, contentType)
			r.body = body
		} else { // fill form
			body, err := toForm(r.formParam)
			if err != nil {
				return nil, err
			}
			r.AddHeaderIfNot(HeaderContentType, MimeURLEncoded)
			r.body = body
		}
	}

	req, err := http.NewRequest(r.method, r.baseURL.String(), r.body)
	if err != nil {
		return nil, err
	}

	// fill header
	r.AddHeaderIfNot(HeaderUserAgent, DefaultUserAgentValue)
	header, err := toHeader(req.Header, r.headerParam)
	if err != nil {
		return nil, err
	}
	req.Header = header
	r.rawRequest = req
	return req, nil
}

func (r *Request) messageBody() []byte {
	body, _ := ioutil.ReadAll(r.rawRequest.Body)
	r.rawRequest.Body.Close()
	r.rawRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	return body
}

func (r *Request) String() string {
	return fmt.Sprintf("%s %s", r.method, r.baseURL.String())
}

func (r *Request) Message() string {
	return r.message(true)
}

func (r *Request) MessageHead() string {
	return r.message(false)
}

func (r *Request) message(body bool) string {
	req, err := r.Clone().fill()
	if err != nil {
		return err.Error()
	}

	for _, v := range r.client.cli.Jar.Cookies(req.URL) {
		req.AddCookie(v)
	}

	b, err := httputil.DumpRequest(req, false)
	if err != nil {
		return err.Error()
	}

	if body {
		b = append(b, r.messageBody()...)
	}
	return string(b)
}
