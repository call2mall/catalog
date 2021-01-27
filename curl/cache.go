package curl

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

type Cache struct {
	mx      *sync.Mutex
	dirPath string
}

func NewCache(dirPath string) (cache *Cache) {
	return &Cache{
		mx:      &sync.Mutex{},
		dirPath: dirPath,
	}
}

func (cache *Cache) Save(rawUrl string, bs []byte) (err error) {
	if len(bs) == 0 {
		return
	}

	path := cache.genPath(rawUrl)

	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(path, bs, 0666)

	return
}

func (cache *Cache) Load(rawUrl string) (bs []byte, err error) {
	path := cache.genPath(rawUrl)

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		bs, err = []byte{}, nil

		return
	} else if err != nil {
		return
	}

	bs, err = ioutil.ReadFile(path)

	return
}

func (cache *Cache) RemoveCacheForDomain(domain string) (err error) {
	cache.mx.Lock()
	path := filepath.Join(cache.dirPath, domain)
	cache.mx.Unlock()

	return os.RemoveAll(path)
}

func (cache Cache) genPath(rawUrl string) (path string) {
	data, err := url.Parse(rawUrl)
	if err != nil {
		return
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(rawUrl)))

	section := hash[0:2]

	cache.mx.Lock()
	path = filepath.Join(cache.dirPath, data.Host, section, hash)
	cache.mx.Unlock()

	return
}
