package echo

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

func TestDial(t *testing.T) {
	// Create a cancellable context to control the echo server's lifecycle.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the context is cancelled at the end of the test.

	// Start a UDP echo server. If it fails, the test will stop.
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	// Dial the echo server as a client. If connection fails, the test stops.
	client, err := net.Dial("udp", serverAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	// Ensure the client connection is closed at the end of the test.
	defer func() { _ = client.Close() }()

	// Create an "interloper" to interrupt communication by sending a message to the client.
	interloper, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	// Ensure the interloper is properly closed.
	defer func() { _ = interloper.Close() }()

	// Interloper sends a message to the client before the echo server replies.
	interrupt := []byte("Pardon me")
	n, err := interloper.WriteTo(interrupt, client.LocalAddr())
	if err != nil {
		t.Fatal(err)
	}

	// Ensure the correct number of bytes was sent by the interloper.
	if len(interrupt) != n {
		t.Fatalf("Wrote %d bytes of %d", n, len(interrupt))
	}

	// Client sends a "ping" message to the echo server.
	ping := []byte("ping")
	_, err = client.Write(ping)
	if err != nil {
		t.Fatal(err)
	}

	// Buffer to read the server's reply.
	buf := make([]byte, 1024)

	// Client reads the server's reply.
	n, err = client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the server echoed back the exact "ping" message.
	if !bytes.Equal(ping, buf[:n]) {
		t.Errorf("Expected reply %q; actual reply %q", ping, buf[:n])
	}

	// Set a deadline for the client to stop reading after 1 second.
	err = client.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	// Attempt to read again, expecting no further packets after the deadline.
	_, err = client.Read(buf)
	if err == nil {
		t.Fatal("Unexpected packet received after deadline")
	}
}
