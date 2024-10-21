package ch03

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestDeadline(t *testing.T) {
	// Channel to synchronize between goroutines
	sync := make(chan struct{})

	// Start a listener on an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close() // Ensure the listener is closed after the test

	// Goroutine to handle server-side connection
	go func() {
		// Accept an incoming connection
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer func() {
			conn.Close() // Close the connection when done
			close(sync)  // Notify the main routine that the connection was handled
		}()

		// Set a deadline of 5 seconds for the connection
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		// Attempt to read from the connection (will block until data is received)
		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		nErr, ok := err.(net.Error)
		if !ok || !nErr.Timeout() {
			// Verify that the error is a timeout
			t.Errorf("Expected timeout error; actual: %v", err)
		}

		// Signal that the connection has been processed
		sync <- struct{}{}

		// Set another deadline and try reading again
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			t.Error(err)
			return
		}

		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
		}
	}()

	// Client-side: Dial the listener
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close() // Ensure the client connection is closed

	// Wait for the server to handle the connection
	<-sync

	// Write data to the connection
	_, err = conn.Write([]byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	// Try reading from the connection again
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err != io.EOF {
		// We expect an EOF error indicating the server closed the connection
		t.Errorf("Expected server termination; actual: %v", err)
	}
}

