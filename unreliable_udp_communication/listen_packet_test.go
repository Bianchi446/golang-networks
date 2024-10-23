package echo

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestListenPacketUDP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the UDP echo server
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// Create a UDP client
	client, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Create another "interloper" UDP connection to send an unexpected message
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// Send an interrupting message from the interloper to the client
	interrupt := []byte("Pardon me")
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}
	_ = interloper.Close()

	// Check if the interrupt message was sent correctly
	if l := len(interrupt); l != n {
		t.Fatalf("Wrote %d bytes of %d", n, l)
	}

	// Prepare a buffer to receive messages
	buf := make([]byte, 1024)

	// Read from client (which should receive the interrupting message)
	n, addr, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the interrupt message matches the expected data
	if !bytes.Equal(interrupt, buf[:n]) {
		t.Errorf("Expected reply %q; actual reply %q", interrupt, buf[:n])
	}

	// Check if the sender is the interloper
	if addr.String() != interloper.LocalAddr().String() {
		t.Errorf("Expected message from %q; actual sender is %q", interloper.LocalAddr(), addr)
	}

	// Now send a "ping" message from the client to the server
	ping := []byte("Ping")
	_, err = client.WriteTo(ping, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	// Read the server's response to the ping
	n, addr, err = client.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the server echoed the ping message correctly
	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("Expected reply %q; actual reply %q", ping, buf[:n])
	}

	// Check if the reply came from the correct server address
	if addr.String() != serverAddr.String() {
		t.Errorf("Expected message from %q; actual sender is %q", serverAddr, addr)
	}
}
