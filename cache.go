package requests

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path"
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
	os.MkdirAll(s, 0755)
	return fileCacheDir(s)
}

func MemoryCache() memoryCacheDir {
	return memoryCacheDir{}
}

func Hash(r *Request) (string, error) {
	msg, err := r.Unique()
	if err != nil {
		return "", err
	}
	data := md5.Sum(msg)
	name := hex.EncodeToString(data[:])
	return name, nil
}

type memoryCacheDir struct {
	m sync.Map
}

func (f memoryCacheDir) Hash(r *Request) (string, error) {
	return Hash(r)
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
	return Hash(r)
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
	return ioutil.WriteFile(path.Join(string(f), name), data, 0666)
}

func (f fileCacheDir) Del(name string) error {
	return os.Remove(path.Join(string(f), name))
}
