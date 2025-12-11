package streamgo

import (
	"sync"
	"testing"
	"time"
)

func TestStreamController_Broadcast(t *testing.T) {
	ctrl := NewStreamController[int]()
	defer ctrl.Close()

	stream := ctrl.Stream()

	var wg sync.WaitGroup
	wg.Add(2)

	received1 := 0
	received2 := 0

	sub1 := stream.Listen(func(data int) {
		received1 = data
		wg.Done()
	}, nil)
	defer sub1.Cancel()

	sub2 := stream.Listen(func(data int) {
		received2 = data
		wg.Done()
	}, nil)
	defer sub2.Cancel()

	// Wait for subscriptions to be ready (async nature of subscribe/listen might need brief sync,
	// but controller.Add blocks if no subscribers? No, Add is async-ish or blocks?
	// Controller logic: copy subs, try send. If buffered, non-blocking.
	// If unbuffered, blocking until sub reads.
	// Default buffer is small (2).

	// Let's sleep briefly to ensure subs are registered
	time.Sleep(10 * time.Millisecond)

	ctrl.Add(42)

	wg.Wait()

	if received1 != 42 {
		t.Errorf("Sub1 expected 42, got %d", received1)
	}
	if received2 != 42 {
		t.Errorf("Sub2 expected 42, got %d", received2)
	}
}

func TestStreamController_Close(t *testing.T) {
	ctrl := NewStreamController[int]()
	stream := ctrl.Stream()

	done := make(chan struct{})
	stream.Listen(func(data int) {}, func(err error) {
		// No error expected on close
	})

	// Monitor stream flow
	// Stream.Listen loop returns on channel close

	// We need to verify that Listen loop exits.
	// We can't easily hook into "Listen finished" without a Done callback or similar in Listen which currently returns *StreamSubscription.
	// BUT, if we peek the implementation, Listen loop waits for sub.ch close.
	// Controller.Close closes sub.ch.

	go func() {
		// Send some data
		ctrl.Add(1)
		ctrl.Close()
		close(done)
	}()

	select {
	case <-done:
		// Controller.Close returned
	case <-time.After(100 * time.Millisecond):
		t.Error("Controller.Close didn't return in time")
	}

	// Verify Listeners stop?
	// We rely on the fact that existing tests check concurrent close.
}

func TestStreamController_Error(t *testing.T) {
	ctrl := NewStreamController[string]()
	stream := ctrl.Stream()
	defer ctrl.Close()

	var errReceived error
	var wg sync.WaitGroup
	wg.Add(1)

	stream.Listen(func(data string) {}, func(err error) {
		errReceived = err
		wg.Done()
	})

	time.Sleep(10 * time.Millisecond)
	testErr := &testError{msg: "oops"}
	ctrl.AddError(testErr)

	wg.Wait()

	if errReceived != testErr {
		t.Errorf("Expected error %v, got %v", testErr, errReceived)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
