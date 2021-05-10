package proxy

import (
	"fmt"
	"testing"
)

func TestNewProxies(t *testing.T) {
	p := NewProxies([]string{
		"http://195.123.240.170:12000",
		"http://195.123.240.170:12001",
	})

	addr, ok := p.Next()
	if !ok {
		t.Fatal("Can't get next proxy")
	}

	if addr != "http://195.123.240.170:12000" {
		t.Fatal("Catch unexpected proxy address")
	}

	addr, ok = p.Next()
	if !ok {
		t.Fatal("Can't get next proxy")
	}

	if addr != "http://195.123.240.170:12001" {
		t.Fatal("Catch unexpected proxy address")
	}

	addr, ok = p.Next()
	if !ok {
		t.Fatal("Can't get next proxy")
	}

	if addr != "http://195.123.240.170:12000" {
		t.Fatal("Catch unexpected proxy address")
	}
}

func TestCheckProxy(t *testing.T) {
	proxies := NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.131.158:29842",
		"http://emiles01:xVypbJnv@51.83.17.222:29842",
		"http://emiles01:xVypbJnv@51.83.17.50:29842",
		"http://emiles01:xVypbJnv@51.91.196.148:29842",
		"http://emiles01:xVypbJnv@51.91.196.182:29842",
		"http://emiles01:xVypbJnv@51.91.196.59:29842",
	})

	var (
		addr string
		ok   bool
	)
	for i := 0; i < 10; i++ {
		addr, ok = proxies.Next()
		if !ok {
			t.Fatal("Proxy.Next return unexpected flag")
		}

		if !CheckProxy(addr) {
			t.Fatal("CheckProxy return unexpected result")
		}
	}
}

func TestLoadProxiesFromFile(t *testing.T) {
	list, err := LoadProxiesFromFile("../proxies.csv")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(list)
}
