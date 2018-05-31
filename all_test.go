package requests

import (
	"net/http"
	"net/url"
	"testing"

	ffmt "gopkg.in/ffmt.v1"
)

func TestA(t *testing.T) {
	mock, err := NewMock(func(s string) {
		t.Log(s)
	})
	if err != nil {
		t.Log(err)
	}
	mock.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ffmt.P(r)
	})
	host := mock.URL()
	u, err := url.Parse(host)
	if err != nil {
		t.Error(err)
	}
	cli := NewClient().
		WithLogger().
		WithSystemCertPool().
		WithCookieJar().
		NewRequest().
		SetBaseURL(u).
		SetHeader(HeaderAccept, "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8").
		SetHeader("Accept-Language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,zh-HK;q=0.7").
		SetHeader(HeaderUserAgent, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36")

	resp, err := cli.Get("/")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(resp.Body()))
}
