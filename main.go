package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL   *url.URL
	IsAlive bool
	Mux sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	backedends []*Backend
	current uint64
}

func (s *ServerPool) NextPeer() int {
	NextIndex := atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backedends))
	return int(NextIndex)
}

func main() {
	url, err := url.Parse("http://localhost:8080")
	if err != nil {
		fmt.Println("Error parsing URL")
	}
	revproxy := httputil.NewSingleHostReverseProxy(url)
	server := http.HandlerFunc(revproxy.ServeHTTP)
	fmt.Println(reflect.TypeOf(server))
}
