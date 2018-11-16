package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type proxyServer struct {
	matchers map[string]string
	mux      sync.Mutex
}

var proxy = proxyServer{
	matchers: make(map[string]string),
}

func getRedirect(r *http.Request) (string, string) {
	normalHost := r.Header.Get("X-Forward-Host")
	r.Header.Del("X-Forward-Host")
	proxy.mux.Lock()
	defer proxy.mux.Unlock()
	for path, host := range proxy.matchers {
		if strings.HasPrefix(r.URL.Path, path) {
			return host, strings.TrimPrefix(r.URL.Path, path)
		}
	}
	return normalHost, r.URL.Path
}

func addRedirect(path, host string) {
	url, err := url.Parse(host)
	if err != nil {
		fmt.Println(err)
		return
	}
	proxy.mux.Lock()
	defer proxy.mux.Unlock()
	proxy.matchers[path] = url.Host
}
