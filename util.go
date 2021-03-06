package requests

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// Common HTTP methods.
//
// Unless otherwise noted, these are defined in RFC 7231 section 4.3.
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

	charsetPrefix   = "; charset="
	charsetUTF8     = charsetPrefix + "utf-8"
	MimeJSON        = "application/json" + charsetUTF8
	MimeXML         = "application/xml" + charsetUTF8
	MimeTextPlain   = "text/plain" + charsetUTF8
	MimeOctetStream = "application/octet-stream" + charsetUTF8
	MimeURLEncoded  = "application/x-www-form-urlencoded" + charsetUTF8
	MimeFormData    = "multipart/form-data" + charsetUTF8

	HeaderUserAgent       = "User-Agent"
	HeaderAccept          = "Accept"
	HeaderContentType     = "Content-Type"
	HeaderContentLength   = "Content-Length"
	HeaderContentEncoding = "Content-Encoding"
	HeaderAuthorization   = "Authorization"
)

// Default
var (
	DefaultPrefix         = "REQUESTS"
	DefaultVersion        = "1.0"
	DefaultUserAgentValue = "Mozilla/5.0 (compatible; " + DefaultPrefix + "/" + DefaultVersion + "; +https://github.com/wzshiming/requests)"
)

// paramPair represent custom data part for header path query form
type paramPair struct {
	Param string
	Value string
}

type paramPairs []*paramPair

func (t *paramPairs) Clone() paramPairs {
	n := make(paramPairs, len(*t))
	copy(n, *t)
	return n
}

func (t *paramPairs) add(i int, n *paramPair) {
	*t = append(*t, nil)
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

func toHeader(header http.Header, p paramPairs, tr transform.Transformer) (http.Header, error) {
	for _, v := range p {
		val := v.Value
		if tr != nil {
			var err error
			val, _, err = transform.String(tr, val)
			if err != nil {
				val = v.Value
			}
		}
		header[v.Param] = append(header[v.Param], val)
	}
	return header, nil
}

func toQuery(p paramPairs, tr transform.Transformer) (string, error) {
	param := url.Values{}
	for _, v := range p {
		val := v.Value
		if tr != nil {
			vv, err := url.QueryUnescape(val)
			if err == nil {
				vv, _, err = transform.String(tr, vv)
				if err != nil {
					val = vv
				}
			}
		}
		param[v.Param] = append(param[v.Param], val)
	}
	return param.Encode(), nil
}

var toPathCompile = regexp.MustCompile(`\{[^}]*\}`)

func toPath(path string, p paramPairs, tr transform.Transformer) (string, error) {
	path = toPathCompile.ReplaceAllStringFunc(path, func(s string) string {
		k := s[1 : len(s)-1]
		// Because the number is small, it's faster to use the loop directly
		for _, v := range p {
			if v.Param == k {
				val := v.Value
				if tr != nil {
					var err error
					val, _, err = transform.String(tr, val)
					if err != nil {
						val = v.Value
					}
				}
				return val
			}
		}
		return s
	})
	return path, nil
}

func toForm(p paramPairs, tr transform.Transformer) (io.Reader, string, error) {
	vs := url.Values{}
	for _, v := range p {
		val := v.Value
		if tr != nil {
			var err error
			val, _, err = transform.String(tr, val)
			if err != nil {
				val = v.Value
			}
		}
		vs.Add(v.Param, val)
	}
	return bytes.NewBufferString(vs.Encode()), MimeURLEncoded, nil
}

func toMulti(p paramPairs, m multiFiles, tr transform.Transformer) (io.Reader, string, error) {
	buf := bytes.NewBuffer(nil)
	mw := multipart.NewWriter(buf)

	for _, v := range p {
		val := v.Value
		if tr != nil {
			var err error
			val, _, err = transform.String(tr, val)
			if err != nil {
				val = v.Value
			}
		}
		err := mw.WriteField(v.Param, val)
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

func readCookies(line string) (cookies []*http.Cookie) {
	parts := strings.Split(strings.TrimSpace(line), ";")
	if len(parts) == 1 && parts[0] == "" {
		return
	}
	// Per-line attributes
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.TrimSpace(parts[i])
		if len(parts[i]) == 0 {
			continue
		}
		name, val := parts[i], ""
		if j := strings.Index(name, "="); j >= 0 {
			name, val = name[:j], name[j+1:]
		}

		// Strip the quotes, if present.
		if len(val) > 1 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}

		cookies = append(cookies, &http.Cookie{Name: name, Value: val})
	}

	return cookies
}

// Cookies raw to Cookies.
func Cookies(raw interface{}) []*http.Cookie {
	switch t := raw.(type) {
	case []*http.Cookie:
		return t
	case *http.Cookie:
		return []*http.Cookie{t}
	case http.Cookie:
		return []*http.Cookie{&t}
	case string:
		return readCookies(t)
	}
	return nil
}

// URL raw to URL structure.
func URL(raw interface{}) *url.URL {
	switch t := raw.(type) {
	case *url.URL:
		return t
	case url.URL:
		return &t
	case string:
		r, _ := url.Parse(t)
		return r
	}
	return nil
}

// TryCharset try charset
func TryCharset(r io.Reader, contentType string) (io.Reader, string, error) {
	mediatype, params, err := mime.ParseMediaType(contentType)
	if err == nil {
		if cs, ok := params["charset"]; ok {
			if e, _ := charset.Lookup(cs); e != nil && e != encoding.Nop {
				params["charset"] = "uft-8"
				return transform.NewReader(r, e.NewDecoder()), mime.FormatMediaType(mediatype, params), nil
			}
		} else if mediatype == "text/html" {
			t, n, err := TryHTMLCharset(r)
			if err != nil {
				return nil, "", err
			}
			if n == "" {
				n = contentType
			}
			return t, n, nil
		}
	}
	return r, contentType, nil
}

// TryHTMLCharset try html charset
func TryHTMLCharset(r io.Reader) (io.Reader, string, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	reader := bytes.NewReader(buf)
	root, err := html.Parse(reader)
	reader.Seek(0, 0)
	var read io.Reader = reader
	if err != nil {
		return nil, "", err
	}

	if root == nil {
		return read, "", nil
	}

	node := root.FirstChild

	for ; node != nil; node = node.NextSibling {
		if node.Type == html.ElementNode && node.Data == "html" {
			node = node.FirstChild
			break

		}
	}

	if node == nil {
		return read, "", nil
	}

	for ; node != nil; node = node.NextSibling {
		if node.Type == html.ElementNode && node.Data == "head" {
			node = node.FirstChild
			break
		}
	}

	if node == nil {
		return read, "", nil
	}

	for ; node != nil; node = node.NextSibling {
		if node.Data == "meta" {
			m := map[string]string{}
			for _, attr := range node.Attr {
				m[strings.ToLower(attr.Key)] = attr.Val
			}
			switch m["http-equiv"] {
			case "content-type", "Content-Type":
				contentType := m["content"]
				if contentType != "" {
					mediatype, params, err := mime.ParseMediaType(contentType)
					if err == nil {
						if cs, ok := params["charset"]; ok {
							if e, _ := charset.Lookup(cs); e != nil && e != encoding.Nop {
								params["charset"] = "utf-8"
								return transform.NewReader(read, e.NewDecoder()), mime.FormatMediaType(mediatype, params), nil
							}
						}
					}
					return read, contentType, nil
				}
			}
		}
	}
	return read, "", nil
}
