package requests

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"runtime"
	"time"

	"golang.org/x/net/publicsuffix"
)

// NewClient the create a client
func NewClient() *Client {
	return &Client{
		cli: &http.Client{},
	}
}

// Client contains basic
type Client struct {
	cli   *http.Client
	log   *log.Logger
	proxy *url.URL
}

// NewRequest method creates a request instance.
func (c *Client) NewRequest() *Request {
	return &Request{
		client: c,
	}
}

// SetCookieJar method set cookie jar.
func (c *Client) SetCookieJar(jar *cookiejar.Jar) *Client {
	c.cli.Jar = jar
	return c
}

// WithCookieJar method with default cookie jar.
func (c *Client) WithCookieJar() *Client {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		c.printError(err)
	}
	return c.SetCookieJar(jar)
}

// SetLogger method sets given writer for logging.
func (c *Client) SetLogger(w io.Writer) *Client {
	c.log = log.New(w, DefaultPrefix, log.LstdFlags)
	return c
}

// SetLogger method with logger.
func (c *Client) WithLogger() *Client {
	return c.SetLogger(os.Stdout)
}

// SetTimeout method sets timeout for request raised from client.
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.cli.Timeout = timeout
	return c
}

// SetTLSClientConfig method sets TLSClientConfig.
func (c *Client) SetTLSClientConfig(config *tls.Config) *Client {
	transport, err := c.getTransport()
	if err != nil {
		c.printError(err)
		return c
	}
	transport.TLSClientConfig = config
	return c
}

// SetKeepAlives method sets the keep alives.
func (c *Client) SetKeepAlives(enable bool) *Client {
	transport, err := c.getTransport()
	if err != nil {
		c.printError(err)
		return c
	}
	transport.DisableKeepAlives = !enable
	return c
}

// SetProxyFunc method sets the Proxy function.
func (c *Client) SetProxyFunc(proxy func(*http.Request) (*url.URL, error)) *Client {
	transport, err := c.getTransport()
	if err != nil {
		c.printError(err)
		return c
	}
	transport.Proxy = proxy
	return c
}

// SetProxyURL method sets the Proxy URL.
func (c *Client) SetProxyURL(u *url.URL) *Client {
	return c.SetProxyFunc(http.ProxyURL(u))
}

// AddRootCert method helps to add one or more root certificates into requests client
func (c *Client) AddRootCert(cert *x509.Certificate) *Client {
	config, err := c.getTLSConfig()
	if err != nil {
		c.printError(err)
		return c
	}
	if config.RootCAs == nil {
		config.RootCAs = x509.NewCertPool()
	}
	config.RootCAs.AddCert(cert)
	return c
}

// WithSystemCertPool sets system cert poll
func (c *Client) WithSystemCertPool() *Client {
	config, err := c.getTLSConfig()
	if err != nil {
		c.printError(err)
		return c
	}

	if runtime.GOOS != "windows" {
		ca, err := x509.SystemCertPool()
		if err != nil {
			c.printError(err)
			return c
		}
		config.RootCAs = ca
	}
	return c
}

// SetSkipVerify sets skip ca verify
func (c *Client) SetSkipVerify(b bool) *Client {
	config, err := c.getTLSConfig()
	if err != nil {
		c.printError(err)
		return c
	}
	config.InsecureSkipVerify = b
	return c
}

// getTLSConfig returns a TLS config
func (c *Client) getTLSConfig() (*tls.Config, error) {
	transport, err := c.getTransport()
	if err != nil {
		return nil, err
	}
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	}
	return transport.TLSClientConfig, nil
}

// getTransport returns a transport
func (c *Client) getTransport() (*http.Transport, error) {
	if c.cli.Transport == nil {
		c.cli.Transport = &http.Transport{}
	}

	if transport, ok := c.cli.Transport.(*http.Transport); ok {
		return transport, nil
	}
	return nil, errors.New("not a *http.Transport")
}

// do executes and returns response
func (c *Client) do(req *Request) (*Response, error) {
	req.sendAt = time.Now()
	resp, err := c.cli.Do(req.rawRequest)
	if err != nil {
		return nil, err
	}

	var body []byte
	if resp.Body != nil {
		defer func() {
			resp.Body.Close()
		}()
		if !req.discardResponse {
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
		}
	}
	response := &Response{
		request:     req,
		rawResponse: resp,
		body:        body,
		recvAt:      time.Now(),
	}
	return response, nil
}

func (c *Client) printError(i interface{}) {
	if c.log != nil {
		c.log.Printf("ERROR %v", i)
	}
}
