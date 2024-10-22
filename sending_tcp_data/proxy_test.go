package main

import (
	"io"
	"net"
	"sync"
	"testing"
)

// proxy sets up a bidirectional copy between a reader and a writer.
func proxy(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer) // Type assertion for from to io.Writer
	toReader, toIsReader := to.(io.Reader)       // Type assertion for to to io.Reader

	if toIsReader && fromIsWriter {
		// Start a goroutine to copy data from 'to' to 'from'.
		go func() { _, _ = io.Copy(fromWriter, toReader) }()
	}

	// Copy data from 'from' to 'to' and return any error encountered.
	_, err := io.Copy(to, from)
	return err
}

func TestProxy(t *testing.T) {
	var wg sync.WaitGroup

	// Server listens for a "ping" message and responds with "pong"
	server, err := net.Listen("tcp", "127.0.0.1:0") // Use port 0 to automatically allocate an available port
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close() // Ensure the server is closed at the end of the test

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()

				for {
					buf := make([]byte, 1024)
					n, err := c.Read(buf)
					if err != nil {
						if err == io.EOF {
							return // Exit if there's no more data to read
						}
						t.Error(err)
						return
					}

					// Handle the message read from the connection
					switch msg := string(buf[:n]); msg {
					case "ping":
						_, err = c.Write([]byte("pong")) // Respond with "pong"
					default:
						_, err = c.Write(buf[:n]) // Echo the received message
					}

					if err != nil {
						t.Error(err)
						return
					}
				}
			}(conn)
		}
	}()

	// Proxy server proxies messages from client connections to the destination server
	proxyServer, err := net.Listen("tcp", "127.0.0.1:0") // Use port 0 to automatically allocate an available port
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done() // Ensure WaitGroup counter is decremented

		conn, err := proxyServer.Accept()
		if err != nil {
			return
		}

		go func(from net.Conn) {
			defer from.Close()

			to, err := net.Dial("tcp", server.Addr().String())
			if err != nil {
				t.Error(err)
				return
			}
			defer to.Close()

			err = proxy(from, to)
			if err != nil && err != io.EOF {
				t.Error(err)
			}
		}(conn)
	}()

	// Create a client connection to the proxy server
	conn, err := net.Dial("tcp", proxyServer.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	// Test messages to send and their expected replies
	msgs := []struct {
		Message string
		Reply   string
	}{
		{"ping", "pong"},
		{"pong", "pong"},
		{"echo", "echo"},
		{"ping", "pong"},
	}

	// Send messages and check replies
	for i, m := range msgs {
		_, err := conn.Write([]byte(m.Message))
		if err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		actual := string(buf[:n])
		t.Logf("%q -> proxy -> %q", m.Message, actual)

		if actual != m.Reply {
			t.Errorf("%d: expected reply: %q; actual: %q", i, m.Reply, actual)
		}
	}

	// Close connections
	_ = conn.Close()
	_ = proxyServer.Close()
	_ = server.Close()

	wg.Wait() // Wait for all goroutines to finish
}
