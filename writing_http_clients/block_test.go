package main

import(
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func blockIndefinitely(w http.ResponseWriter, r *http.Request) {
	select {}
}

func TestBlockIndefinitely(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(blockIndefinitely))
	_, _ = http.Get(ts.URL)
	t.Fatal("Client did not indefinitely block")
}

func TestBlockIndefinitelyWithTimeout(t *testing.T) {
	ts := httptest.WebServer(http.HandlerFunc(blockIndefinitely))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		return 
	}
	_ = resp.Body.Close()
}

ctx, cancel := context.WithCancel(context.Background())
timer := time.AferFunc(5*time.Second, cancel)

// Make an HTTP request, read the response headers

timer.Reset(5*time.Second())



req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
if err != nil {
	t.Fatal(err)
}

req.Close = true // Close the underlying TCP connection after 
				// Reading the web server response

	
