package ch03

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 1 * time.Second // Default interval for pings

func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration

	// First, check if a reset value is provided or if the context is done
	select {
	case <-ctx.Done():
		return
	case interval = <-reset:
	default:
	}

	// Set the interval to the default if it's zero or negative
	if interval <= 0 {
		interval = defaultPingInterval
	}

	timer := time.NewTimer(interval) // Create a timer for the interval
	defer func() {
		if !timer.Stop() { // Stop the timer if it's running
			<-timer.C // Drain the channel if it was already fired
		}
	}()

	for {
		select {
		case <-ctx.Done(): // If the context is done, exit
			return
		case newInterval := <-reset: // Listen for a new interval
			if !timer.Stop() { // Stop the current timer
				<-timer.C // Drain the channel if it was already fired
			}
			if newInterval > 0 {
				interval = newInterval // Update the interval
			}
		case <-timer.C: // When the timer fires
			if _, err := w.Write([]byte("ping")); err != nil {
				// Handle write error, potentially track and act on consecutive timeouts here
				return
			}

			_ = timer.Reset(interval) // Reset the timer with the updated interval
		}
	}
}
