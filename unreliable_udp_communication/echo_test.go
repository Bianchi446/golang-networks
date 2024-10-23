package echo

import (
	"bytes"
	"context"
	"net"
	"testing"
)

// TestEchoServerUDP tests the echoServerUDP function.
func TestEchoServerUDP(t *testing.T) {
	// Create a cancellable context for the server.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the echo server and bind it to a local address.
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err) // Fail the test if the server fails to start.
	}

	// Create a UDP client.
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err) // Fail the test if the client cannot bind to an address.
	}
	defer func() { _ = client.Close() }() // Ensure the client connection is closed.

	// Send a "ping" message from the client to the server.
	msg := []byte("ping")
	_, err = client.WriteTo(msg, serverAddr)
	if err != nil {
		t.Fatal(err) // Fail if sending the message fails.
	}

	// Create a buffer to store the server's response.
	buf := make([]byte, 1024)
	n, addr, err := client.ReadFrom(buf) // Read the server's reply.
	if err != nil {
		t.Fatal(err) // Fail if reading the response fails.
	}

	// Check if the response came from the expected server address.
	if addr.String() != serverAddr.String() {
		t.Fatalf("Received reply from %q instead of %q", addr, serverAddr)
	}

	// Verify if the received message matches the sent message.
	if !bytes.Equal(msg, buf[:n]) {
		t.Errorf("Expected reply %q; actual reply %q", msg, buf[:n])
	}
}
