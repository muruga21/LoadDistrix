package main

import (
	"bytes"
	"context"
	"io"
	"loadbalancer/lib"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

type RetryType int
type AttemptsType int

const (
	Attempts AttemptsType = iota
	Retry
)

var serverPool lib.ServerPool

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
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

// main function with arguement for configfile
func main() {
	// get file name from argument
	arg := os.Args
	if len(arg) != 2 {
		log.Fatal("usage go run main.go <config-file>'")
	}
	
	// declare slice for backend server
	backendservers := []string{}

	// read the config file and get the host and url.
	var config lib.Config
	config, err := lib.ReadConfig(arg[1])
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range config.BackendConfig {
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
			if r.Body != nil {
				bodyBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
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
			log.Println("This is the retry count", retries, "of the server", serverPool.Current)

			if retries < 3 {
				time.Sleep(10 * time.Millisecond)
				ctx := context.WithValue(r.Context(), Retry, retries+1) // increment the retry count and set it in context
				log.Println("check")

				proxy.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// if the retry count is more than 3 then mark the server as down
			log.Printf("[%s] Marking server as down\n", be.Host)
			serverPool.MarkDownTheServer(be, false)
			// attempts := GetAttemptsFromContext(r)
			// ctx := context.WithValue(r.Context(), Attempts, attempts+1)

			lb(w, r) // this function will find the next alive server and redirect the request
		}

		serverPool.Backends = append(serverPool.Backends, &lib.ServerNode{
			URL:          be,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}
	// http.HandleFunc("/", testHandler)
	server := &http.Server{
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      http.HandlerFunc(lb),
	}
	log.Println("Server is starting on port 8000")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}
