package auth

import (
	"log"
	"net"
	"strings" // Added import for strings package
	"golang.org/x/sys/unix"
	"os/user" // Importing user package for user.LookupId
)

// allowed checks if the Unix connection is from an allowed group.
func allowed(conn *net.UnixConn, groups map[string]struct{}) bool { // Corrected function signature
	if conn == nil || groups == nil || len(groups) == 0 {
		return false // Return false if connection or groups are invalid
	}

	file, _ := conn.File() // Get the file descriptor for the connection
	defer func() { _ = conn.Close() }() // Ensure the connection is closed

	var (
		err   error
		ucred *unix.Ucred // Corrected type from Unix.Ucred to unix.Ucred
	)

	for {
		ucred, err = unix.GetsockoptUcred(int(file.Fd()), unix.SOL_SOCKET, // Corrected SOL_SOCKET
			unix.SO_PEERCRED) // Get the credentials of the peer
		if err == unix.EINTR {
			continue // syscall interrupted; retry
		}
		if err != nil {
			log.Println(err) // Log the error if GetsockoptUcred fails
			return false
		}
		break // Break the loop if no errors occurred
	}

	// Lookup the user based on UID from ucred
	u, err := user.LookupId(string(ucred.Uid))
	if err != nil {
		log.Println(err) // Log the error if user lookup fails
		return false
	}

	// Get the group IDs for the user
	gids, err := u.GroupIds()
	if err != nil {
		log.Println(err) // Log the error if fetching group IDs fails
		return false
	}

	// Check if the user belongs to any allowed group
	for _, gid := range gids {
		if _, ok := groups[gid]; ok {
			return true // Return true if the user is in an allowed group
		}
	}
	return false // Return false if no matching group found
}
