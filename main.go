package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type RetryType int
type AttemptsType int

const (
	Attempts AttemptsType = iota
	Retry
)

type BackendServer_config struct {
	Host string `json:"host"`
	Url  string `json:"url"`
}

type Config struct {
	Backend_config []BackendServer_config `json:"backend"`
}

type Backend struct {
	URL          *url.URL
	Alive        bool
	Mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) IsAlive() bool {
	b.Mux.RLock()
	defer b.Mux.RUnlock()
	return b.Alive
}

func (b *Backend) SetAlive(alive bool) {
	b.Mux.Lock()
	defer b.Mux.Unlock()
	b.Alive = alive
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

var serverPool ServerPool

// get the next server index in the server pool
func (s *ServerPool) NextServerIndex() int {
	nxtIndex := atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends))
	if(s.current >= math.MaxUint64){
		atomic.StoreUint64(&s.current, 0)
	}
	return int(nxtIndex)
}

func (s *ServerPool) GetNextPeer() *Backend {
	//next peer index.. we dont know if the peer is alive or not.. we need to iterate through the server pool to find the next aliver server
	log.Println("Getting next peer")
	nxtIndex := s.NextServerIndex()
	lenOfBackendArr := len(s.backends)
	lenghtNeedToTraverse := lenOfBackendArr + nxtIndex //start from the next index and traverse the entire server pool [cycle]

	for i := nxtIndex; i < lenghtNeedToTraverse; i++ {
		index := i % lenOfBackendArr
		log.Printf("Checking server at index %d, Alive: %v\n", index, s.backends[index].IsAlive())
		if s.backends[index].IsAlive() { //check if the server is alive
			atomic.StoreUint64(&s.current, uint64(index))
			return s.backends[index] //return the alive server
		}
	}

	return nil
}

// this functions returns the retry count from the context
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

// this function returns the attempts from the context
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func lb(w http.ResponseWriter, r *http.Request) {
	peer := serverPool.GetNextPeer()
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		http.Error(w, "Service not available, max attempts reached", http.StatusServiceUnavailable)
		return
	}
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		log.Println("This is request host", r.Host)
		log.Println(serverPool.current)
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

func testHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Test handler received request:", r.URL.Path)
	w.Write([]byte("Hello from test handler"))
}

// main function
func main() {
	backendservers := []string{}

	jsonFile, _ := os.Open("LoadDistrix.config.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config Config

	json.Unmarshal(byteValue, &config)
	for _, node := range config.Backend_config {
		backendservers = append(backendservers, node.Url)
	}

	if len(backendservers) == 0 {
		log.Println("No backend servers found")
		return
	}
	for _, backend := range backendservers {
		log.Println("Load balancing to the backend server: ", backend)
		be, err := url.Parse(backend)
		log.Println(be)
		if err != nil {
			log.Println("Error parsing URL")
		}
		proxy := httputil.NewSingleHostReverseProxy(be) //this is one of other backend servers need to pust in server pool
		proxy.Director = func(r *http.Request) {
			r.Header.Set("User-Agent", "Your-User-Agent")
			r.Header.Set("Accept", "application/json")
			r.Header.Set("X-Custom-Header", "CustomValue")
			// Adjust URL and Host
			r.URL.Scheme = be.Scheme
			r.URL.Host = be.Host
			r.Host = be.Host
		}

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			log.Printf("[%s] Request Canceled: %v\n", be.Host, r.Context().Err() == context.Canceled)
			log.Printf("[%s] %s\n", be.Host, e.Error())
		
			retries := GetRetryFromContext(r) // by default the retry count is 0
			log.Println("This is the retry count", retries, "of the server", serverPool.current)
		
			if retries < 3 {
				time.Sleep(10 * time.Millisecond)
				ctx := context.WithValue(r.Context(), Retry, retries+1) // increment the retry count and set it in context
		
				// Recreate the request body
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Printf("Error reading body: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				r.Body = ioutil.NopCloser(bytes.NewReader(body))
		
				proxy.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		
			// if the retry count is more than 3 then mark the server as down
			log.Printf("[%s] Marking server as down\n", be.Host, serverPool.current)
			serverPool.MarkDownTheServer(be, false)
			attempts := GetAttemptsFromContext(r)
			ctx := context.WithValue(r.Context(), Attempts, attempts+1)
		
			// Recreate the request body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading body: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			r.Body = ioutil.NopCloser(bytes.NewReader(body))
		
			lb(w, r.WithContext(ctx)) // this function will find the next alive server and redirect the request
		}

		serverPool.backends = append(serverPool.backends, &Backend{
			URL:          be,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}
	// http.HandleFunc("/", testHandler)
	server := &http.Server{
		Addr: ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
		Handler:http.HandlerFunc(lb),
	}
	log.Println("Server is starting on port 8000")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}
