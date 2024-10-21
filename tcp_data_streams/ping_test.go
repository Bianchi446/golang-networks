// Using incoming messages to advance the deadline 

package ch03

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestPingerAdvanceDeadline(t *testing.T) {
	done := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Use port 0 to allow the OS to assign an available port
	if err != nil {
		t.Fatal(err)
	}
	begin := time.Now()

	go func() {
		defer close(done) // Close done channel when the goroutine exits

		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			conn.Close() // Correctly close the connection
		}()

		resetTimer := make(chan time.Duration, 1)
		resetTimer <- time.Second
		go Pinger(ctx, conn, resetTimer) // Start the Pinger in a separate goroutine

		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			t.Logf("[%s] %s",
				time.Since(begin).Truncate(time.Second), buf[:n])

			resetTimer <- 0 // Reset the timer by sending 0
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for i := 0; i < 4; i++ { // Read up to four pings
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}

	_, err = conn.Write([]byte("PONG!!!")) // Reset the ping timer 
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 4; i++ { // Read up to four pings
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.Fatal(err)
			}
			break
		}
		t.Logf("[%s] %s", time.Since(begin).Truncate(time.Second), buf[:n])
	}

	<-done // Wait for the listener goroutine to finish
	end := time.Since(begin).Truncate(time.Second)
	t.Logf("[%s] done", end)
	if end != 9*time.Second {
		t.Fatalf("Expected EOF at 9 seconds; actual %s", end)
	}
}
