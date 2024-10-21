package main

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Use :0 for an available port
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer func() { done <- struct{}{} }()

		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err == io.EOF {
							break
						}
						t.Error(err)
						return
					}
					t.Logf("Received: %q", buf[:n])
				}
			}(conn)
		}
	}()

	// Client-side dialing
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	// Send data to the listener
	_, err = conn.Write([]byte("Hello"))
	if err != nil {
		t.Fatal(err)
	}

	conn.Close() // Close client connection
	<-done       // Wait for the server to handle the connection

	listener.Close() // Close the listener
	<-done           // Wait for the listener goroutine to finish
}
