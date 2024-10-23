package echo

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

// echoServerUDP starts a simple UDP echo server that listens for incoming messages
// and sends the same message back to the client (echoes it).
func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	// Bind to the specified UDP address
	s, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("Binding to udp %s: %w", addr, err)
	}

	go func() {
		go func() {
		// Ensure the server shuts down when the context is canceled
		<-ctx.Done()
		_ = s.Close()
	}()

	// Buffer to store incoming data
	buf := make([]byte, 1024)

	// Main server loop: handle incoming messages
	
		for {
			// Read data from the client
			n, clientAddr, err := s.ReadFrom(buf) // client to server
			if err != nil {
				return
			}

			// Echo the message back to the client
			_, err = s.WriteTo(buf[:n], clientAddr) // server to client
			if err != nil {
				return
			}
		}
	}()
	// Return the server's local address
	return s.LocalAddr(), nil
}

func main() {
	// Create a context to manage the server's lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the UDP echo server on 127.0.0.1:8080
	serverAddr, err := echoServerUDP(ctx, "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Failed to start UDP echo server: %v", err)
	}

	fmt.Println("Echo server listening on:", serverAddr.String())

	// Set up a UDP client to interact with the server
	conn, err := net.Dial("udp", serverAddr.String())
	if err != nil {
		log.Fatalf("Failed to connect to echo server: %v", err)
	}
	defer conn.Close()

	// Message to send to the server
	message := []byte("Hello, UDP Echo Server!")
	_, err = conn.Write(message) // Send the message to the server
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// Buffer to store the echoed message from the server
	buf := make([]byte, 1024)
	n, err := conn.Read(buf) // Read the echoed message
	if err != nil {
		log.Fatalf("Failed to receive echo: %v", err)
	}

	// Print the received echo message
	fmt.Printf("Received echo: %s\n", string(buf[:n]))

	// Allow the server some time to process before shutting down
	time.Sleep(1 * time.Second)
	// Cancel the context to stop the server
	cancel()
}
