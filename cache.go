package requests

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
)

type Cache interface {
	Hash(*Request) string
	Load(name string) (*Response, bool)
	Save(name string, resp *Response)
	Del(name string)
}

func FileCacheDir(s string) fileCacheDir {
	os.MkdirAll(s, 0755)
	return fileCacheDir(s)
}

func MemoryCache() memoryCacheDir {
	return memoryCacheDir{}
}

func Hash(r *Request) string {
	msg, err := r.Unique()
	if err != nil {
		return ""
	}
	data := md5.Sum(msg)
	name := hex.EncodeToString(data[:])
	return name
}

type CacheModel struct {
	Location    *url.URL
	StatusCode  int
	Header      http.Header
	Body        []byte
	ContentType string
}

type memoryCacheDir struct {
	m sync.Map
}

func (f memoryCacheDir) Hash(r *Request) string {
	return Hash(r)
}

func (f memoryCacheDir) Load(name string) (*Response, bool) {
	d, ok := f.m.Load(name)
	if !ok {
		return nil, false
	}
	data, ok := d.(*Response)
	if !ok {
		return nil, false
	}
	return data, ok
}

func (f memoryCacheDir) Save(name string, resp *Response) {
	f.m.Store(name, resp)
	return
}

func (f memoryCacheDir) Del(name string) {
	f.m.Delete(name)
	return
}

type fileCacheDir string

func (f fileCacheDir) Hash(r *Request) string {
	return Hash(r)
}

func (f fileCacheDir) Load(name string) (*Response, bool) {
	data, err := ioutil.ReadFile(path.Join(string(f), name))
	if err != nil {
		return nil, false
	}
	m := CacheModel{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, false
	}
	resp := Response{
		statusCode:  m.StatusCode,
		header:      m.Header,
		location:    m.Location,
		contentType: m.ContentType,
		body:        m.Body,
	}
	return &resp, true
}

func (f fileCacheDir) Save(name string, resp *Response) {
	m := CacheModel{
		StatusCode:  resp.StatusCode(),
		Header:      resp.Header(),
		Location:    resp.location,
		Body:        resp.Body(),
		ContentType: resp.ContentType(),
	}
	data, _ := json.Marshal(m)
	ioutil.WriteFile(path.Join(string(f), name), data, 0666)
	return
}

func (f fileCacheDir) Del(name string) {
	os.Remove(path.Join(string(f), name))
	return
}
