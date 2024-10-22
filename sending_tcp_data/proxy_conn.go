// This function copies any data sent form the 
// source node to the destination node 

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net"
	"time"
)

func proxyConn(source, destination string) error {
	connSource, err := net.Dial("tcp", source)
	if err != nil {
		return err
	}
	defer connSource.Close()

	connDestination, err := net.Dial("tcp", destination)
	if err != nil {
		return err
	}
	defer connDestination.Close()

	// connDestination replies to connSource
	go func() {
		_, _ = io.Copy(connSource, connDestination)
	}()

	// connSource messages to connDestination
	_, err = io.Copy(connDestination, connSource) // Corrected to use connSource
	return err
}

func main() {
	// Start a local server on port 8080 for testing
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello from the local server!")
		})
		fmt.Println("Starting server on :8080...")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Give the server some time to start
	time.Sleep(1 * time.Second)

	// Define source and destination addresses
	sourceAddress := "127.0.0.1:8080"        // Local server
	destinationAddress := "example.com:80" // Destination, e.g., HTTP server

	// Call the proxyConn function
	err := proxyConn(sourceAddress, destinationAddress)
	if err != nil {
		log.Fatal("Error in proxying connection:", err)
	} else {
		fmt.Println("Proxy connection successful")
	}
}
