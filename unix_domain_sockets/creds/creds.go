package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"

	"github.com/Bianchi446/golang-networks/unix_domain_sockets/creds/auth"
)

// init initializes the command-line flags and usage information
func init() {
	flag.Usage = func() { // Corrected from flag.usage to flag.Usage
		_, _ = fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: \n\t%s <group names>\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults() // Print the default flag values
	}
}

// parseGroupNames parses the command-line arguments and returns a map of group names to structs
func parseGroupNames(args []string) map[string]struct{} {
	groups := make(map[string]struct{})

	for _, arg := range args {
		grp, err := user.LookupGroup(arg) // Lookup the group by name
		if err != nil {
			log.Println(err) // Log error if group lookup fails
			continue
		}
		groups[grp.Gid] = struct{}{} // Add the group ID to the map
	}
	return groups // Return the map of groups
}

func main() {
	flag.Parse() // Parse command-line flags

	groups := parseGroupNames(flag.Args()) // Get the groups from command-line arguments
	socket := filepath.Join(os.TempDir(), "creds.sock") // Create a socket file in the temp directory
	addr, err := net.ResolveUnixAddr("unix", socket) // Resolve the Unix socket address
	if err != nil {
		log.Fatal(err) // Log fatal error if address resolution fails
	}

	s, err := net.ListenUnix("unix", addr) // Listen for Unix domain socket connections
	if err != nil {
		log.Fatal(err) // Log fatal error if listening fails
	}

	c := make(chan os.Signal, 1) // Channel to receive OS signals
	signal.Notify(c, os.Interrupt) // Notify the channel on interrupt signals

	go func() {
		<-c // Wait for an interrupt signal
		_ = s.Close() // Close the socket on interrupt
	}()

	fmt.Printf("Listening on %s ... \n", socket) // Print the socket address being listened on

	for {
		conn, err := s.AcceptUnix() // Accept a connection from a Unix domain socket
		if err != nil {
			break // Break the loop on error
		}

		if auth.Allowed(conn, groups) { // Check if the connection is allowed
			_, err := conn.Write([]byte("Welcome\n")) // Send welcome message
			if err != nil {
				log.Println(err) // Log error if writing fails
			}
		}

		_ = conn.Close() // Close the connection after handling
	}
}
