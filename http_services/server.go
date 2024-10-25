package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/awoodbeck/gnp/ch09/handlers"
	"github.com/awoodbeck/gnp/ch09/middleware"
)

var (
	// Define command-line flags for server address, certificate, key, and file directory.
	addr  = flag.String("listen", "127.0.0.1:8080", "listen address")
	cert  = flag.String("cert", "", "certificate")
	pkey  = flag.String("key", "", "private key")
	files = flag.String("files", "./files", "static file directory")
)

func main() {
	// Parse command-line flags.
	flag.Parse()

	// Run the server with specified configuration.
	err := run(*addr, *files, *cert, *pkey)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server gracefully shutdown")
}

func run(addr, files, cert, pkey string) error {
	// Create a new ServeMux to route HTTP requests.
	mux := http.NewServeMux()

	// 1. Handle static files at the /static/ route.
	mux.Handle("/static/",
		http.StripPrefix("/static/",
			middleware.RestrictPrefix(
				".", // Restrict file serving to the specified directory.
				http.FileServer(http.Dir(files)), // Serve static files from `files` directory.
			),
		),
	)

	// 2. Handle the root route ("/") with a GET method.
	mux.Handle("/",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					// 3. Check if HTTP/2 Server Push is supported.
					if pusher, ok := w.(http.Pusher); ok {
						targets := []string{
							"/files/style.css", // 4. Static files to push.
							"/files/hiking.svg",
							"/files/index.html",
						}
						for _, target := range targets {
							// Attempt to push each static file to the client.
							if err := pusher.Push(target, nil); err != nil {
								log.Printf("%s push failed: %v", target, err)
							}
						}
					}
					// Serve the main HTML file for the root route.
					http.ServeFile(w, r, filepath.Join(files, "index.html"))
				},
			),
		},
	)

	// 5. Handle an additional route (/2) with a different HTML file.
	mux.Handle("/2",
		handlers.Methods{
			http.MethodGet: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, filepath.Join(files, "index2.html"))
				},
			),
		},
	)

	// 6. Set up HTTP server configuration.
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
	}

	// Channel to signal server shutdown.
	done := make(chan struct{})
	go func() {
		// Set up OS signal listener for graceful shutdown.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		// Wait for interrupt signal to trigger shutdown.
		for {
			if <-c == os.Interrupt {
				if err := srv.Shutdown(context.Background()); err != nil {
					log.Printf("Shutdown: %v", err)
				}
				close(done)
				return
			}
		}
	}()

	log.Printf("Serving files in %q over %s\n", files, srv.Addr)

	var err error
	// Check if TLS certificate and key are provided.
	if cert != "" && pkey != "" {
		log.Println("TLS enabled")
		err = srv.ListenAndServeTLS(cert, pkey)
	} else {
		// Start HTTP server if no TLS.
		err = srv.ListenAndServe()
	}

	if err == http.ErrServerClosed {
		err = nil // Server closed gracefully.
	}

	// Wait for shutdown signal before exiting.
	<-done

	return err
}
