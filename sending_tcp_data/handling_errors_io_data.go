// How to check for temporary errors while writting data to network conn

package main

import (
	"errors"
	"log"
	"net"
	"time"
)

var (
	err error
	n   int
	i   = 7 // Threshold for temporary write retries
)

func WriteData(conn net.Conn) error {
	for i := 7; i > 0; i-- { // Iterating with retry attempts
		n, err = conn.Write([]byte("Hello world"))
		if err != nil {
			if nErr, ok := err.(net.Error); ok && nErr.Temporary() { // Correct type assertion
				log.Println("Temporary error: ", nErr)
				time.Sleep(10 * time.Second)
				continue
			}
			return err // Return the error if it's not temporary
		}
		break // Exit the loop on success
	}

	if i == 0 {
		return errors.New("Temporary write failure threshold exceeded") // Error if retries exhausted
	}

	log.Printf("Wrote %d bytes to %s\n", n, conn.RemoteAddr()) // Log the successful write
	return nil
}

func main() {
	// Establish a connection to a local TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Accept a connection in a separate goroutine
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			return
		}
		defer conn.Close()

		err = WriteData(conn)
		if err != nil {
			log.Println("WriteData error:", err)
		}
	}()

	// Dial to the listener
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Read the response (optional)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Error reading:", err)
	} else {
		log.Printf("Received: %s\n", string(buf[:n]))
	}
}
