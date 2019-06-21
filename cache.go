package requests

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
)

var ErrNotExist = errors.New("not exist")

type Cache interface {
	Hash(*Request) (string, error)
	Load(name string) (*Response, error)
	Save(name string, resp *Response) error
	Del(name string) error
}

func FileCacheDir(s string) fileCacheDir {
	return fileCacheDir(s)
}

func MemoryCache() memoryCacheDir {
	return memoryCacheDir{}
}

type memoryCacheDir struct {
	m sync.Map
}

func (f memoryCacheDir) Hash(r *Request) (string, error) {
	msg, err := r.RawRequest()
	if err != nil {
		return "", err
	}
	name := msg.URL.String()
	return name, nil
}

func (f memoryCacheDir) Load(name string) (*Response, error) {
	d, ok := f.m.Load(name)
	if !ok {
		return nil, ErrNotExist
	}
	data, ok := d.(*Response)
	if !ok {
		return nil, ErrNotExist
	}
	return data, nil
}

func (f memoryCacheDir) Save(name string, resp *Response) error {
	f.m.Store(name, resp)
	return nil
}

func (f memoryCacheDir) Del(name string) error {
	f.m.Delete(name)
	return nil
}

type fileCacheDir string

func (f fileCacheDir) Hash(r *Request) (string, error) {
	msg, err := r.RawRequest()
	if err != nil {
		return "", err
	}
	name := msg.URL.String()
	sum := md5.Sum([]byte(name))
	h := hex.EncodeToString(sum[:])
	return path.Join(msg.URL.Scheme, msg.URL.Host, msg.URL.Path, h), nil
}

func (f fileCacheDir) Load(name string) (*Response, error) {
	data, err := ioutil.ReadFile(path.Join(string(f), name))
	if err != nil {
		return nil, ErrNotExist
	}

	resp := &Response{}
	err = resp.UnarshalText(data)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (f fileCacheDir) Save(name string, resp *Response) error {
	data, err := resp.MarshalText()
	if err != nil {
		return err
	}
	p := filepath.Join(string(f), name)
	dir, _ := filepath.Split(p)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p, data, 0644)
}

func (f fileCacheDir) Del(name string) error {
	return os.Remove(path.Join(string(f), name))
}
