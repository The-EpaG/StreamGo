package streamgo

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestBackpressure_DataLoss(t *testing.T) {
	controller := NewStreamController[int]()
	stream := controller.Stream()

	receivedCount := 0
	expectedCount := 100

	// Helper to signal completion
	done := make(chan struct{})

	stream.Listen(func(data int) {
		// Simulate slow subscriber
		time.Sleep(10 * time.Millisecond)
		receivedCount++
		if receivedCount == expectedCount {
			close(done)
		}
	}, func(err error) {
		t.Errorf("Error: %v", err)
	})

	// Fast producer
	go func() {
		for i := 0; i < expectedCount; i++ {
			controller.Add(i)
			// Very short sleep to allow *some* processing but clearly overwhelm the 10ms subscriber
			time.Sleep(1 * time.Millisecond)
		}
		controller.Close()
	}()

	// Wait for a bit to see how many we get
	select {
	case <-done:
		// Success! Received all events.
	case <-time.After(5 * time.Second): // Increased timeout for CI/test reliability
		t.Errorf("Timed out. Received %d out of %d events.\n", receivedCount, expectedCount)
	}
}

func TestBackpressure_CancelUnblocks(t *testing.T) {
	// 1. Create Controller
	controller := NewStreamController[int]()

	// 2. Add a blocked subscriber (Buffer 0)
	streamBlocked := controller.StreamWithBuffer(0)
	subBlocked := streamBlocked.Listen(func(data int) {
		// Never consumes
		t.Logf("Blocked Sub received: %d (Should not happen often)\n", data)
		time.Sleep(1 * time.Hour)
	}, nil)

	// 3. Add a fast subscriber (to verify it gets processed after unblock)
	streamFast := controller.StreamWithBuffer(2)
	var receivedFast int32
	subFast := streamFast.Listen(func(data int) {
		atomic.AddInt32(&receivedFast, 1)
		t.Logf("Fast Sub received: %d\n", data)
	}, nil)
	defer subFast.Cancel()

	// 4. Start Producer in background
	go func() {
		t.Log("Producer: Sending 1")
		controller.Add(1) // Goes to buffer (fast) or wait (blocked)
		t.Log("Producer: Sent 1")

		t.Log("Producer: Sending 2")
		controller.Add(2) // Should BLOCK here because blocking sub has buffer 0 and won't consume
		t.Log("Producer: Sent 2")

		t.Log("Producer: Sending 3")
		controller.Add(3)
		t.Log("Producer: Sent 3")

		controller.Close()
	}()

	// 5. Wait to ensure producer is blocked
	time.Sleep(500 * time.Millisecond)
	t.Log("Main: Cancelling blocked subscriber...")

	// 6. CANCEL the blocked subscriber -> Should unblock producer
	subBlocked.Cancel()

	// 7. Wait for completion
	time.Sleep(1 * time.Second)

	if atomic.LoadInt32(&receivedFast) >= 2 {
		// Success
	} else {
		t.Error("Failure: Producer likely deadlocked.")
	}
}
