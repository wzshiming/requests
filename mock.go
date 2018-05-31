package requests

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
)

// Mock represent a registry mock
type Mock struct {
	server   *httptest.Server
	hostport string
	handlers map[string]http.HandlerFunc
	mu       sync.Mutex
}

// HandleFunc register the specified handler for the registry mock
func (tr *Mock) HandleFunc(path string, h http.HandlerFunc) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.handlers[path] = h
}

// NewMock creates a registry mock
func NewMock(t func(string)) (*Mock, error) {
	testReg := &Mock{handlers: make(map[string]http.HandlerFunc)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()

		var matched bool
		var err error
		for re, function := range testReg.handlers {
			matched, err = regexp.MatchString(re, url)
			if err != nil {
				t("Error with handler regexp")
			}
			if matched {
				function(w, r)
				break
			}
		}

		if !matched {
			t("Unable to match " + url + " with regexp")
		}
	}))

	testReg.server = ts
	testReg.hostport = ts.URL
	return testReg, nil
}

// URL returns the url of the registry
func (tr *Mock) URL() string {
	return tr.hostport
}

// Close closes mock and releases resources
func (tr *Mock) Close() {
	tr.server.Close()
}
