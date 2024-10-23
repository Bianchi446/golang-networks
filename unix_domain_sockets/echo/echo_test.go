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

// TestEchoServerUnix tests the Unix domain socket-based echo server.
func TestEchoServerUnix(t *testing.T) {
	// Create a temporary directory for Unix socket files.
	dir, err := ioutil.TempDir("", "echo_unix") // Fixed space in the first argument (should be an empty string)
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the context gets canceled to stop the server.

	// Define the Unix socket file path.
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid())) // Use process ID to name the socket file.

	// Start the streaming echo server using Unix domain socket.
	rAddr, err := streamingEchoServer(ctx, "unix", socket)
	if err != nil {
		t.Fatal(err)
	}

	// Change permissions for the socket to allow read/write access for everyone.
	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	// Dial the server using the Unix domain socket address.
	conn, err := net.Dial("unix", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }() // Ensure the connection is closed.

	// Send three "Ping" messages to the server.
	msg := []byte("Ping")
	for i := 0; i < 3; i++ {
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Read the server's reply.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	// The server should echo back three "Ping" messages.
	expected := bytes.Repeat(msg, 3)
	if !bytes.Equal(expected, buf[:n]) {
		t.Fatalf("expected reply %q; actual reply %q", expected, buf[:n])
	}

	// Since there is no `closer` defined, we'll remove this line.
	// <-done is also unnecessary as we don't have a channel to wait for, removing it too.
}

