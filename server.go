// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/

package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// content holds static web server content.
//
//go:embed app/index.html
//go:embed app/http_error/404.html
var content embed.FS

type Config struct {
	host string
	port string
}

func addRoutes(
	mux *http.ServeMux,
) {
	mux.HandleFunc("/", handleHTML)
	// mux.HandleFunc("/api/v1/stats", handleServerStats)
}

func handleHTML(
	w http.ResponseWriter,
	r *http.Request,
) {
	var data []byte
	switch r.URL.Path {
	// case "<URL/ROUTE>":
	// 	data = content.ReadFile("<FilePath>"")
	case "/":
		data, _ = content.ReadFile("app/index.html")
	default:
		data, _ = content.ReadFile("app/http_error/404.html")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func handleServerStats(
	w http.ResponseWriter,
	r *http.Request,
) {
	var data []byte

	cpuCores := "Number of CPU's: " + strconv.Itoa(runtime.NumCPU()) + "\n"
	goRoutines := "Number of Go Routine's: " + strconv.Itoa(runtime.NumGoroutine()) + "\n"
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	mallocs := "Memory Allocated: " + strconv.Itoa(int(mem.Mallocs)) + "B\n"
	heapAlloc := "Heap Allocated: " + strconv.Itoa(int(mem.HeapAlloc)) + "B\n"
	heapInUse := "Heap In Use: " + strconv.Itoa(int(mem.HeapInuse)) + "B\n"

	data = []byte(cpuCores + goRoutines + mallocs + heapAlloc + heapInUse)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func NewServer(
	config Config,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux)
	var handler http.Handler = mux
	// Add middleware as follows
	// handler = someMiddleware(handler)
	return handler
}

func run(
	ctx context.Context,
	config Config,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	srv := NewServer(
		// logger,
		config,
	)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.host, config.port),
		Handler: srv,
	}
	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	config := Config{
		host: "0.0.0.0",
		port: "3000",
	}
	if err := run(ctx, config); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
