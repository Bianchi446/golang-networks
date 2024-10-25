package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/awoodbeck/gnp/ch09/handlers"
)

// TestSimpleHTTPServer sets up and tests a simple HTTP server.
func TestSimpleHTTPServer(t *testing.T) {
	// Define a new HTTP server with timeouts and a default handler.
	srv := &http.Server{
		Addr:              "127.0.0.1:8443", // Server address.
		Handler:           http.TimeoutHandler(handlers.DefaultHandler(), 2*time.Minute, ""),
		IdleTimeout:       5 * time.Minute,
		ReadHeaderTimeout: time.Minute,
	}

	// Listen on the specified address.
	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		t.Fatal(err) // Terminate test if the server fails to start.
	}

	// Start the server in a new goroutine.
	go func() {
		// Start the server with TLS.
		err := srv.ServeTLS(l, "cert.pem", "key.pem")
		if err != http.ErrServerClosed {
			t.Error(err) // Log error if server stops unexpectedly.
		}
	}()

	// Define test cases with different HTTP methods and expected responses.
	testCases := []struct {
		method   string    // HTTP method (GET, POST, etc.).
		body     io.Reader // Request body.
		code     int       // Expected HTTP status code.
		response string    // Expected response body.
	}{
		{http.MethodGet, nil, http.StatusOK, "Hello, friend!"},
		{http.MethodPost, bytes.NewBufferString("<World>"), http.StatusOK, "Hello, &lt;World&gt;!"},
		{http.MethodHead, nil, http.StatusMethodNotAllowed, " "},
	}

	client := new(http.Client)
	path := fmt.Sprintf("https://%s/", srv.Addr) // Server address with protocol.

	// Iterate over test cases.
	for i, c := range testCases {
		// Create a new HTTP request.
		r, err := http.NewRequest(c.method, path, c.body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		// Send the request.
		resp, err := client.Do(r)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}

		// Check if the status code matches the expected code.
		if resp.StatusCode != c.code {
			t.Errorf("%d: unexpected status code: %q", i, resp.Status)
		}

		// Read and verify the response body.
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("%d: %v", i, err)
			continue
		}
		_ = resp.Body.Close()

		// Compare response with expected output.
		if c.response != string(b) {
			t.Errorf("%d: expected %q; actual %q", i, c.response, b)
		}
	}

	// Shut down the server.
	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
}
