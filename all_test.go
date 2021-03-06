package requests

import (
	"net/http"
	"testing"
	"time"
)

func TestParam(t *testing.T) {
	mock, err := NewMock(func(err error) {
		t.Error(err)
	})
	if err != nil {
		t.Error(err)
		return
	}
	mock.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
			return
		}
		if r.URL.Path != "/url" {
			t.Error("parameter error")
		}
		u, p, _ := r.BasicAuth()
		if u != "u" || p != "p" {
			t.Error("parameter error")
		}
		if r.FormValue("q") != "query" {
			t.Error("parameter error")
		}

		if r.FormValue("f") != "form" {
			t.Error("parameter error")
		}
	})
	cli := NewRequest().
		SetURLByStr(mock.URL()).
		SetForm("f", "form").
		SetQuery("q", "query").
		SetBasicAuth("u", "p").
		SetPath("p", "url")
	_, err = cli.Post("/{p}")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestContext(t *testing.T) {
	mock, err := NewMock(func(err error) {
		t.Error(err)
	})
	if err != nil {
		t.Error(err)
		return
	}
	_, err = NewRequest().
		SetTimeout(time.Microsecond).
		SetURLByStr(mock.URL()).
		Do()
	if err == nil {
		t.Error("No timely interruption")
		return
	}
}
