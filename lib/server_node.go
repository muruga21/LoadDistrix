package lib

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type ServerNode struct {
	URL          *url.URL
	Alive        bool
	Weight       int
	Mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *ServerNode) IsAlive() bool {
	b.Mux.RLock()
	defer b.Mux.RUnlock()
	return b.Alive
}

func (b *ServerNode) SetAlive(alive bool) {
	b.Mux.Lock()
	defer b.Mux.Unlock()
	b.Alive = alive
}
