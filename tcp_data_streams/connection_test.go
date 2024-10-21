// TCP server capable of listening to incoming requests

package main

import (
	"net"
	"testing"
)

func TestListener(t *testing.T) {
	Listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = Listener.Close() }()

	t.Logf("Bound to %q", Listener.Addr())

	for {
		conn, err := Listener.Accept()
		if err != nil {
			t.Fatal(err) // Use t.Fatal to handle errors during testing
		}

		go func(c net.Conn) {
			defer c.Close()

			// Handle the connection (e.g., read/write data)
			// Add connection handling logic here

		}(conn)
	}
}
