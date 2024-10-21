package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sync := make(chan struct{})

	go func() {
		defer func() { sync <- struct{}{} }() // Corrected to use a proper deferred function syntax

		var d net.Dialer
		d.Control = func(_, addr string, _ syscall.RawConn) error {
			time.Sleep(time.Second) // Simulate a delay
			return nil
		}
		conn, err := d.DialContext(ctx, "tcp", "10.0.0.1:80")
		if err != nil {
			t.Log(err)
			return
		}
		defer conn.Close() // Close the connection if it was established
		t.Error("Connection did not time out") // This line should only execute if no error occurred
	}()

	cancel() // Cancel the context to trigger a timeout
	<-sync   // Wait for the goroutine to signal completion

	if ctx.Err() != context.Canceled {
		t.Errorf("Expected canceled context; actual %q", ctx.Err())
	}
}
