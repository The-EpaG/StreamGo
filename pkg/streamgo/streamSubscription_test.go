package streamgo

import (
	"sync"
	"testing"
	"time"
)

func TestStreamSubscription_Cancel(t *testing.T) {
	var callCount int
	var mu sync.Mutex

	onCancel := func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	sub := NewStreamSubscription(onCancel)

	// cancels first time
	sub.Cancel()

	// cancels second time (idempotent)
	sub.Cancel()

	mu.Lock()
	if callCount != 1 {
		t.Errorf("Expected onCancel to be called exactly once, got %d", callCount)
	}
	mu.Unlock()

	// Verify channel is closed
	select {
	case <-sub.Done:
		// OK
	default:
		t.Error("Expected sub.Done to be closed")
	}
}

func TestStreamSubscription_Wait(t *testing.T) {
	sub := NewStreamSubscription(nil)

	go func() {
		time.Sleep(10 * time.Millisecond)
		sub.Cancel()
	}()

	select {
	case <-sub.Done:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Error("Timed out waiting for subscription to be cancelled")
	}
}
