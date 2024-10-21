package ch03

import (
	"context"
	"fmt"
	"io"
	"time"
)


// ExamplePinger demonstrates the use of the Pinger function.
func ExamplePinger() {
	ctx, cancel := context.WithCancel(context.Background())
	r, w := io.Pipe() 
	done := make(chan struct{})
	resetTimer := make(chan time.Duration, 1)
	resetTimer <- time.Second

	// Start the Pinger in a separate goroutine.
	go func() {
		Pinger(ctx, w, resetTimer)
		close(done) // Signal completion.
	}()

	receivePing := func(d time.Duration, r io.Reader) {
		if d >= 0 {
			fmt.Printf("Resetting timer (%s)\n", d)
			resetTimer <- d // Send new duration to resetTimer.
		}

		now := time.Now()
		buf := make([]byte, 1024)
		m, err := r.Read(buf)
		if err != nil {
			fmt.Printf("Error reading from pipe: %v\n", err)
			return
		}

		// Print the received ping message with elapsed time.
		fmt.Printf("Received %q (%s)\n", buf[:m], time.Since(now).Round(100*time.Millisecond))
	}

	// Simulate different reset intervals.
	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1, -1} {
		fmt.Printf("Run %d:\n", i+1)
		receivePing(time.Duration(v)*time.Millisecond, r) // Call receivePing with simulated intervals.
	}

	cancel() // Cancel the context to stop the Pinger.
	<-done   // Wait for the Pinger to finish.
}
