package ch03

import (
	"io"
	"time"
	"context"
)

const defaultPingInterval = 30 * time.Second

// pinger sends periodic "ping" messages to the provided io.Writer
// at an adjustable interval, controlled by the reset channel.
func pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration

	// Check for an initial interval from the reset channel or use the default.
	select {
	case <-ctx.Done():
		return // Exit if the context is canceled.
	case interval = <-reset:
	default:
		// Use default interval if no value is received.
	}

	// Set to default if the interval is zero or negative.
	if interval <= 0 {
		interval = defaultPingInterval
	}

	// Create a timer for the specified interval.
	timer := time.NewTimer(interval)
	defer func() {
		// Stop the timer when pinger exits.
		if !timer.Stop() {
			<-timer.C // Drain the timer channel.
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return // Exit if the context is canceled.
		case newInterval := <-reset:
			// Stop the timer and reset it with a new interval if provided.
			if !timer.Stop() {
				<-timer.C // Drain the timer channel to prevent unexpected behavior.
			}
			if newInterval > 0 {
				interval = newInterval
			}
			timer.Reset(interval) // Reset the timer with the new interval.

		case <-timer.C: // Timer expired, send a "ping".
			if _, err := w.Write([]byte("ping")); err != nil {
				// Track and act on consecutive timeouts here (e.g., log or alert).
				return
			}
		}
		_= timer.Reset(interval)
	}
}
