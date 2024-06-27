// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"server/pkg/router"
)

type Config struct {
	host string
	port string
}

func NewServer(
	config Config,
) http.Handler {
	mux := http.NewServeMux()
	router.AddRoutes(mux)
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
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
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
