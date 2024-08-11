package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type RetryType int;
type AttemptsType int;

const (
	Attempts AttemptsType = iota
	Retry
)

type Backend struct {
	URL   *url.URL
	Alive bool
	Mux sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) IsAlive() bool {
	b.Mux.RLock()
	defer b.Mux.RUnlock()
	return b.Alive
}

func (b * Backend) SetAlive(alive bool) {
	b.Mux.RLock()
	defer b.Mux.RUnlock()
	b.Alive = alive
}

type ServerPool struct {
	backends []*Backend
	current uint64
}

var serverPool ServerPool

//get the next server index in the server pool
func (s *ServerPool) NextServerIndex() int {
	nxtIndex := atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends))
	return int(nxtIndex)
}

func (s *ServerPool) GetNextPeer() *Backend {
	//next peer index.. we dont know if the peer is alive or not.. we need to iterate through the server pool to find the next aliver server
	nxtIndex := s.NextServerIndex()
	lenOfBackendArr := len(s.backends)
	lenghtNeedToTraverse := lenOfBackendArr + nxtIndex //start from the next index and traverse the entire server pool [cycle]

	for i := nxtIndex; i < lenghtNeedToTraverse; i++ {
		index := i % lenOfBackendArr
		if s.backends[index].IsAlive() { //check if the server is alive
			atomic.StoreUint64(&s.current, uint64(index))
			return s.backends[index] //return the alive server
		}
	}

	return nil
}

//this functions returns the retry count from the context
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

//this function returns the attempts from the context
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func lb(w http.ResponseWriter, r *http.Request) {
	peer := serverPool.GetNextPeer()
	attempts := GetAttemptsFromContext(r)
	fmt.Println(serverPool.current)
	if(attempts >3) {
		http.Error(w, "Service not available, max attempts reached", http.StatusServiceUnavailable)
		return
	}
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func (s *ServerPool) MarkDownTheServer(backendUrl *url.URL, serverStatus bool) {
	for _, backend := range s.backends {
		if backend.URL.String() == backendUrl.String() {
			backend.SetAlive(serverStatus)
			return
		}
	}
}

//main function
func main() {
	url, err := url.Parse("http://localhost:8080")
	if err != nil {
		fmt.Println("Error parsing URL")
	}
	proxy := httputil.NewSingleHostReverseProxy(url) //this is one of other backend servers need to pust in server pool
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("[%s] %s\n", url.Host, e.Error())
		retries := GetRetryFromContext(r) //by default the retry count is 0
		if retries < 3{
			time.Sleep(10 * time.Millisecond)
			ctx := context.WithValue(r.Context(), RetryType(Retry), retries+1) //increment the retry count and set it in context
			proxy.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		//if the retry count is more than 3 then mark the server as down
		serverPool.MarkDownTheServer(url, false)
		attempts := GetAttemptsFromContext(r)
		ctx := context.WithValue(r.Context(), Attempts, attempts+1)
		lb(w, r.WithContext(ctx)) //this function will find the next alive server and redirect the request
	}

	serverPool.backends = append(serverPool.backends, &Backend{
		URL: url,
		Alive: true,
		ReverseProxy: proxy,
	})

	server := http.Server{ // this is main server
		Addr:    fmt.Sprintf(":%d", 8081),
		Handler: http.HandlerFunc(lb),
	  }
	  server.ListenAndServe()	  
}
