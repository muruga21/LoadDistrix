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
	"os/signal"
	"syscall"
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
		proxy := httputil.NewSingleHostReverseProxy(be)
		proxy.Director = func(r *http.Request) {
			if r.Body != nil {
				bodyBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
			r.Header.Set("User-Agent", "Your-User-Agent")
			r.Header.Set("Accept", "application/json")
			r.Header.Set("X-Custom-Header", "CustomValue")
			r.URL.Scheme = be.Scheme
			r.URL.Host = be.Host
			r.Host = be.Host
		}

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			log.Printf("[%s] Request Canceled: %v\n", be.Host, r.Context().Err() == context.Canceled)
			log.Printf("[%s] %s\n", be.Host, e.Error())

			retries := GetRetryFromContext(r)
			log.Println("This is the retry count", retries, "of the server", serverPool.Current)

			if retries < 3 {
				time.Sleep(10 * time.Millisecond)
				ctx := context.WithValue(r.Context(), Retry, retries+1)
				log.Println("check")

				proxy.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			log.Printf("[%s] Marking server as down\n", be.Host)
			serverPool.MarkDownTheServer(be, false)

			lb(w, r)
		}

		serverPool.Backends = append(serverPool.Backends, &lib.ServerNode{
			URL:          be,
			Alive:        true,
			ReverseProxy: proxy,
		})
	}

	server := &http.Server{
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      http.HandlerFunc(lb),
	}

	// Channel to listen for interrupt or termination signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		log.Println("Server is starting on port 8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for termination signal
	<-shutdown

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Shutting down gracefully...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
