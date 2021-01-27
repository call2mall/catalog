package proxy

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Proxies struct {
	mx    *sync.Mutex
	list  *sync.Map
	total int
	ix    int
}

func NewProxies(list []string) (p *Proxies) {
	p = &Proxies{
		mx:   &sync.Mutex{},
		list: &sync.Map{},
	}

	p.Init(list)

	return
}

func (p *Proxies) Init(list []string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	for ix, proxy := range list {
		p.list.Store(ix, proxy)
	}

	p.total = len(list)
	p.ix = 0
}

func (p *Proxies) Next() (addr string, ok bool) {
	p.mx.Lock()
	defer p.mx.Unlock()

	if p.total == 0 {
		return
	}

	val, ok := p.list.Load(p.ix)
	if !ok {
		val, ok = p.list.Load(0)
		if !ok {
			return
		}

		p.ix = -1
	}

	addr = val.(string)

	p.ix++
	if p.ix >= p.total {
		p.ix = 0
	}

	ok = true

	return
}

type answer struct {
	Status string `json:"status"`
}

func CheckProxy(proxyAddr string) (ok bool) {
	proxyUrl, err := url.Parse(proxyAddr)
	if err != nil {
		return
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	rawUrl := "http://ip-api.com/json"

	var req *http.Request
	req, err = http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		return
	}

	var res *http.Response
	res, err = client.Do(req)
	if err != nil {
		return
	}

	var bs []byte
	bs, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	var answer answer
	err = json.Unmarshal(bs, &answer)
	if err != nil {
		return
	}

	ok = answer.Status == "success"

	return
}

func LoadProxiesFromFile(path string) (list []string, err error) {
	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		return
	}

	var bs []byte
	bs, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}

	data := string(bs)
	for _, line := range strings.Split(data, "\n") {
		if len(line) > 0 {
			list = append(list, line)
		}
	}

	return
}
