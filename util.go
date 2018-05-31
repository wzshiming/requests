package requests

import (
	"encoding/base64"
	"io"
	"net/textproto"
)

// In is located in
type In uint8

const (
	_ In = iota
	Header
	Path
	Query
	Body
	Form
)

var (
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

	HeaderUserAgent       = textproto.CanonicalMIMEHeaderKey("User-Agent")
	HeaderAccept          = textproto.CanonicalMIMEHeaderKey("Accept")
	HeaderContentType     = textproto.CanonicalMIMEHeaderKey("Content-Type")
	HeaderContentLength   = textproto.CanonicalMIMEHeaderKey("Content-Length")
	HeaderContentEncoding = textproto.CanonicalMIMEHeaderKey("Content-Encoding")
	HeaderAuthorization   = textproto.CanonicalMIMEHeaderKey("Authorization")
)

var (
	DefaultPrefix         = "REQUESTS"
	DefaultUserAgentValue = DefaultPrefix + " - https://github.com/wzshiming/requests"
)

type paramPair struct {
	Param string
	Value string
}

// multipartField represent custom data part for multipart request
type multiFile struct {
	Param       string
	FileName    string
	ContentType string
	io.Reader
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
