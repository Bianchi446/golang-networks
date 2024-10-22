package ch04

import (
	"bufio"
	"net"
	"reflect"
	"testing"
)

const payload = "The bigger the interface the weaker the abstraction."

func TestScanner(t *testing.T) {
	// Start a TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Use port 0 to select an available port
	if err != nil {
		t.Fatal(err)
		return
	}
	defer listener.Close() // Ensure the listener is closed

	// Goroutine to accept connections and send the payload
	go func() {
		conn, err := listener.Accept() // Accept incoming connection
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close() // Ensure the connection is closed

		// Write the payload to the connection
		_, err = conn.Write([]byte(payload))
		if err != nil {
			t.Error(err)
		}
	}()

	// Dial the listener
	conn, err := net.Dial("tcp", listener.Addr().String()) // Use the listener's address
	if err != nil {
		t.Fatal(err) // Fatal if unable to connect
	}
	defer conn.Close() // Ensure the connection is closed

	// Set up a scanner to read from the connection
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanWords) // Split the input by words

	var words []string

	// Scan the connection input and collect words
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	// Check for any scanning errors
	err = scanner.Err()
	if err != nil {
		t.Error(err)
	}

	// Expected list of words from the payload
	expected := []string{"The", "bigger", "the", "interface", "the", "weaker", "the", "abstraction."}

	// Compare the scanned words with the expected list
	if !reflect.DeepEqual(words, expected) {
		t.Fatalf("Inaccurate scanned words list. Got %#v, expected %#v", words, expected)
	}

	// Log the scanned words
	t.Logf("Scanned words: %#v", words)
}
