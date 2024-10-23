package echo

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// TestEchoServerUnix tests a Unix domain socket echo server using datagram sockets.
func TestEchoServerUnix(t *testing.T) {
	// Create a temporary directory for the Unix socket files.
	dir, err := ioutil.TempDir("", "echo_unixgram") // Removed space in the first argument
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		// Clean up the temporary directory after the test completes.
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	// Create a cancellable context for managing server lifetime.
	ctx, cancel := context.WithCancel(context.Background()) // Changed 'channel' to 'cancel' for clarity.
	sSocket := filepath.Join(dir, fmt.Sprintf("s%d.sock", os.Getpid())) // Server socket path

	// Start the datagram echo server on Unix domain socket.
	serverAddr, err := datagramEchoServer(ctx, "unixgram", sSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel() // Ensure the context gets canceled to stop the server.

	// Set permissions for the server socket to allow access.
	err = os.Chmod(sSocket, os.ModeSocket|0622) // Use proper permissions for the socket
	if err != nil {
		t.Fatal(err)
	}

	// Create a client socket for sending messages to the server.
	cSocket := filepath.Join(dir, fmt.Sprintf("c%d.sock", os.Getpid())) // Client socket path
	client, err := net.ListenPacket("unixgram", cSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }() // Ensure the client socket is closed.

	// Set permissions for the client socket.
	err = os.Chmod(cSocket, os.ModeSocket|0622) // Use proper permissions for the socket
	if err != nil {
		t.Fatal(err)
	}

	// Prepare the message to send to the server.
	msg := []byte("ping")
	for i := 0; i < 3; i++ { // Changed 'o' to '0' to ensure the loop runs correctly
		_, err = client.WriteTo(msg, serverAddr) // Send the message to the server
		if err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 1024) // Buffer to hold replies from the server
	for i := 0; i < 3; i++ {
		n, addr, err := client.ReadFrom(buf) // Read replies from the server
		if err != nil {
			t.Fatal(err)
		}

		// Check if the reply is from the expected server address.
		if addr.String() != serverAddr.String() { // Changed to '!=' for the correct condition
			t.Fatalf("Received reply from unexpected address %q instead of %q", addr, serverAddr)
		}

		// Validate the received message matches the expected message.
		if !bytes.Equal(msg, buf[:n]) {
			t.Fatalf("expected reply %q; actual reply %q", msg, buf[:n])
		}
	}
}
