package ch04

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

// TestPayloads tests sending multiple payloads over a TCP connection.
func TestPayloads(t *testing.T) {
	b1 := Binary("Clear is better than clever.") // First binary payload
	b2 := Binary("Don't panic")                  // Second binary payload
	s1 := String("Errors are values.")           // String payload
	payloads := []Payload{&b1, &s1, &b2}         // Collecting payloads into a slice

	// Start a TCP listener on a random local port
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// Goroutine to handle the server side of the connection
	go func() {
		conn, err := listener.Accept() // Accept the connection
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		// Write each payload to the connection
		for _, p := range payloads {
			_, err = p.WriteTo(conn) // Use WriteTo method to send data
			if err != nil {
				t.Error(err)
				break
			}
		}
	}()

	// Client side: Dial the listener and connect
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Read and decode each payload from the connection
	for i := 0; i < len(payloads); i++ {
		actual, err := decode(conn) // Custom decode function to read from conn
		if err != nil {
			t.Fatal(err)
		}
		// Compare the actual payload with the expected one
		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("value mismatch: %v != %v", expected, actual)
			continue
		}

		t.Logf("[%T] %[1]q", actual) // Log the payload type and value
	}
}

// TestPayloadSize tests payload size limits by trying to create an oversized payload.
func TestPayloadSize(t *testing.T) {
	buf := new(bytes.Buffer) // Buffer to hold the payload
	err := buf.WriteByte(BinaryType) // Write the binary type
	if err != nil {
		t.Fatal(err)
	}
	// Write a large size (1 GB) to the buffer, exceeding the allowed payload size
	err = binary.Write(buf, binary.BigEndian, uint32(1<<30)) // 1 GB size
	if err != nil {
		t.Fatal(err)
	}

	var b Binary
	// Try to read the payload and check if it exceeds the max payload size
	_, err = b.ReadFrom(buf)
	if err != ErrMaxPayloadSize { // Expecting an error due to exceeding size
		t.Fatalf("Expected ErrMaxPayloadSize; actual: %v", err)
	}
}

