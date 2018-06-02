package requests

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"sort"
)

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"

	CharsetUTF8     = "; charset=utf-8"
	MimeJSON        = "application/json" + CharsetUTF8
	MimeXML         = "application/xml" + CharsetUTF8
	MimeTextPlain   = "text/plain" + CharsetUTF8
	MimeOctetStream = "application/octet-stream" + CharsetUTF8
	MimeURLEncoded  = "application/x-www-form-urlencoded" + CharsetUTF8
	MimeFormData    = "multipart/form-data" + CharsetUTF8

	HeaderUserAgent       = "User-Agent"
	HeaderAccept          = "Accept"
	HeaderContentType     = "Content-Type"
	HeaderContentLength   = "Content-Length"
	HeaderContentEncoding = "Content-Encoding"
	HeaderAuthorization   = "Authorization"
)

var (
	DefaultPrefix         = "REQUESTS "
	DefaultUserAgentValue = DefaultPrefix + " - https://github.com/wzshiming/requests"
)

// paramPair represent custom data part for header path query form
type paramPair struct {
	Param string
	Value string
}

type paramPairs []*paramPair

func (t *paramPairs) add(i int, n *paramPair) {
	*t = append(*t, n)
	l := len(*t)
	copy((*t)[i+1:l], (*t)[i:l-1])
	(*t)[i] = n
}

func (t *paramPairs) Add(param, value string) {
	i := t.SearchIndex(param)
	t.add(i, &paramPair{
		Param: param,
		Value: value,
	})
}

func (t *paramPairs) AddReplace(param, value string) {
	i := t.SearchIndex(param)
	tt := t.Index(i - 1)
	if tt == nil || tt.Param != param {
		t.add(i, &paramPair{
			Param: param,
			Value: value,
		})
	} else {
		tt.Value = value
	}
	return
}

func (t *paramPairs) AddNoRepeat(param, value string) {
	i := t.SearchIndex(param)
	tt := t.Index(i - 1)
	if tt == nil || tt.Param != param {
		t.add(i, &paramPair{
			Param: param,
			Value: value,
		})
	}
	return
}

func (t *paramPairs) Search(name string) (*paramPair, bool) {
	i := t.SearchIndex(name)
	if i == 0 {
		return nil, false
	}
	tt := t.Index(i - 1)
	if tt == nil || tt.Param != name {
		return nil, false
	}
	return tt, true
}

func (t *paramPairs) SearchIndex(name string) int {
	i := sort.Search(t.Len(), func(i int) bool {
		d := t.Index(i)
		if d == nil {
			return false
		}
		return d.Param < name
	})
	return i
}

func (t *paramPairs) Index(i int) *paramPair {
	if i >= t.Len() || i < 0 {
		return nil
	}
	return (*t)[i]
}

func (t *paramPairs) Len() int {
	return len(*t)
}

// multiFile represent custom data part for multipart request
type multiFile struct {
	Param       string
	FileName    string
	ContentType string
	io.Reader
}

type multiFiles []*multiFile

func toHeader(req *http.Request, p paramPairs) error {
	for _, v := range p {
		req.Header.Add(v.Param, v.Value)
	}
	return nil
}

func toQuery(u *url.URL, p paramPairs) error {
	param := u.Query()
	for _, v := range p {
		param.Add(v.Param, v.Value)
	}
	u.RawQuery = param.Encode()
	return nil
}

var toPathCompile = regexp.MustCompile(`\{.*\}`)

func toPath(u *url.URL, p paramPairs) error {
	u.Path = toPathCompile.ReplaceAllStringFunc(u.Path, func(s string) string {
		k := s[1 : len(s)-1]
		// Because the number is small, it's faster to use the loop directly
		for _, v := range p {
			if v.Param == k {
				return v.Value
			}
		}
		return s
	})
	return nil
}

func toForm(p paramPairs) (io.Reader, error) {
	vs := url.Values{}
	for _, v := range p {
		vs.Add(v.Param, v.Value)
	}
	return bytes.NewBufferString(vs.Encode()), nil
}

func toMulti(p paramPairs, m multiFiles) (io.Reader, string, error) {
	buf := bytes.NewBuffer(nil)
	mw := multipart.NewWriter(buf)

	for _, v := range p {
		err := mw.WriteField(v.Param, v.Value)
		if err != nil {
			return nil, "", err
		}
	}

	for _, v := range m {
		w, err := mw.CreateFormFile(v.Param, v.FileName)
		if err != nil {
			return nil, "", err
		}
		_, err = io.Copy(w, v.Reader)
		if err != nil {
			return nil, "", err
		}
	}

	err := mw.Close()
	if err != nil {
		return nil, "", err
	}
	return buf, mw.FormDataContentType(), nil
}

// See 2 (end of page 4) http://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
