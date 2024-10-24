package main

import (
	"fmt"
	"net/http"
	"time"
)

// Function to fetch the Date header and calculate the time skew
func CheckTimeSkew() error {
	// Send a HEAD request to time.gov
	resp, err := http.Head("https://www.time.gov/")
	if err != nil {
		return fmt.Errorf("failed to fetch time.gov: %w", err)
	}
	defer resp.Body.Close() // Ensure response body is closed properly

	// Capture the current time rounded to the nearest second
	now := time.Now().Round(time.Second)

	// Get the Date header from the response
	date := resp.Header.Get("Date")
	if date == "" {
		return fmt.Errorf("no date header received from time.gov")
	}

	// Parse the Date header into time.Time format
	dt, err := time.Parse(time.RFC1123, date)
	if err != nil {
		return fmt.Errorf("failed to parse date header: %w", err)
	}

	// Calculate the skew between local time and time.gov server time
	skew := now.Sub(dt)

	// Output the results
	fmt.Printf("time.gov server time: %s (local time skew: %s)\n", dt, skew)
	return nil
}

func main() {
	// Call the function to check the time skew
	err := CheckTimeSkew()
	if err != nil {
		// Handle any errors that may occur
		fmt.Println("Error:", err)
	}
}
