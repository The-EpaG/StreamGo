package streamgo

import (
	"sync"
	"testing"
	"time"
)

func TestTopic_Stress(t *testing.T) {
	// 1. Concurrent Add and Close (Fix Send Panic)
	t.Run("Concurrent Add and Close", func(t *testing.T) {
		ctrl := NewStreamController[int]()
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				ctrl.Add(i) // Should not panic even if closed
				time.Sleep(100 * time.Microsecond)
			}
		}()

		go func() {
			defer wg.Done()
			time.Sleep(20 * time.Millisecond)
			ctrl.Close()
		}()

		wg.Wait()
	})

	// 2. Double Close Panic (Subscribe/Cancel vs Close)
	t.Run("Double Close Scenario", func(t *testing.T) {
		ctrl := NewStreamController[int]()
		stream := ctrl.Stream()

		var wg sync.WaitGroup
		wg.Add(2)

		// Subscriber
		sub := stream.Listen(func(data int) {}, func(err error) {})

		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			sub.Cancel() // Cancel while closing
		}()

		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			ctrl.Close() // Close while canceling
		}()

		wg.Wait()
	})

	// 3. DoS Limit
	t.Run("Max Subscribers Limit", func(t *testing.T) {
		ctrl := NewStreamController[int]()
		defer ctrl.Close()

		// MaxSubscribers is 128
		for i := 0; i < 128; i++ {
			s := ctrl.Stream()
			// Actually subscribe to fill the slots
			s.Listen(func(d int) {}, func(e error) {})
		}

		// 129th should fail at Listen time
		s := ctrl.Stream()
		// s is not nil anymore, but Listen should fail

		sub := s.Listen(func(d int) {}, func(e error) {})

		select {
		case <-sub.Done:
			// Success: Subscription immediately cancelled/rejected
		default:
			t.Errorf("Expected subscription to be rejected due to limit, but it's active")
		}
	})
}
