// Reading data from a network connection into a byte slice

package main

import (
	"crypto/rand"
	"io"
	"net"
	"testing"
)

func TestReadIntoBuffer(t *testing.T) {
	// Create a payload of 16 MB
	payload := make([]byte, 1<<24) // 16 MB
	_, err := rand.Read(payload) // Generate a random payload
	if err != nil {
		t.Fatal(err) // Fatal if there's an error reading random bytes
	}

	// Set up a TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Use port 0 for an available port
	if err != nil {
		t.Fatal(err) // Fatal if unable to start the listener
	}
	defer listener.Close() // Ensure the listener is closed

	// Start a goroutine to accept connections
	go func() {
		conn, err := listener.Accept() // Accept incoming connection
		if err != nil {
			t.Log(err) // Log if there's an error accepting
			return
		}
		defer conn.Close() // Ensure the connection is closed

		// Write the payload to the connection
		_, err = conn.Write(payload)
		if err != nil {
			t.Error(err) // Log error if unable to write payload
		}
	}()

	// Dial the listener
	conn, err := net.Dial("tcp", listener.Addr().String()) // Use the listener's address
	if err != nil {
		t.Fatal(err) // Fatal if unable to connect
	}
	defer conn.Close() // Ensure the connection is closed

	// Create a buffer for reading data
	buf := make([]byte, 1<<19) // 512 KB
	for {
		n, err := conn.Read(buf) // Read data from the connection
		if err != nil {
			if err != io.EOF {
				t.Error(err) // Log error if not an EOF error
			}
			break // Exit loop on EOF
		}
		t.Logf("Read %d bytes", n) // Log the number of bytes read
	}
	conn.Close()
}
