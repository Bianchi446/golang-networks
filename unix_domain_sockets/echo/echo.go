// streamingEchoServer creates a generic stream-based echo server.
// It listens on the specified network and address, and echoes back any data
// received from clients. The server shuts down when the provided context is cancelled.

package echo

import (
	"context"
	"net"
	"os"
)

// streamingEchoServer starts a TCP/UDP echo server on the given network and address.
// The server runs until the provided context is canceled.
// Parameters:
// - ctx: Context to handle server shutdown.
// - network: Network type (e.g., "tcp", "udp").
// - addr: Address to listen on (e.g., "localhost:8080").
// Returns:
// - net.Addr: The address the server is bound to.
// - error: Any error encountered during server setup.
func streamingEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	// Listen for incoming connections on the specified network and address.
	s, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	// Goroutine to handle context cancellation and server shutdown.
	go func() {
		go func() {
			<-ctx.Done() // Wait for the context to be cancelled.
			_ = s.Close() // Gracefully close the listener when context is cancelled.
		}()

		// Accept incoming client connections in a loop.
		for {
			conn, err := s.Accept()
			if err != nil {
				return // Exit if there's an error accepting a connection.
			}

			// Goroutine to handle each client connection.
			go func() {
				defer func() { _ = conn.Close() }() // Ensure the connection is closed.

				// Echo loop: read from the connection and write back the same data.
				for {
					buf := make([]byte, 1024) // Buffer to store client data.
					n, err := conn.Read(buf)  // Read data from the connection.
					if err != nil {
						return // Exit if there's an error reading from the connection.
					}

					_, err = conn.Write(buf[:n]) // Write the received data back to the client.
					if err != nil {
						return // Exit if there's an error writing to the connection.
					}
				}
			}()
		}
	}()

	return s.Addr(), nil // Return the server's listening address.
}

func datagramEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	s, err := net.ListenPacket(network, addr) // Listen for incoming packets on the specified network and address
	if err != nil {
		return nil, err // Return error if unable to listen
	}

	go func() {
		// This goroutine will close the socket when the context is done
		go func() {
			<-ctx.Done() // Wait for the context to be canceled
			_ = s.Close() // Close the socket
			if network == "unixgram" {
				_ = os.Remove(addr) // Remove the socket file if it's a Unix domain socket
			}
		}()

		buf := make([]byte, 1024) // Buffer to hold incoming data
		for {
			n, clientAddr, err := s.ReadFrom(buf) // Read from the socket
			if err != nil {
				return // Exit the goroutine on error
			}

			_, err = s.WriteTo(buf[:n], clientAddr) // Echo the data back to the client
			if err != nil {
				return // Exit the goroutine on error
			}
		}
	}()

	return s.LocalAddr(), nil // Return the address the server is listening on
}
