package streamgo

import (
	"testing"
	"time"
)

func TestStream_BasicFlow(t *testing.T) {
	ctrl := NewStreamController[int]()
	defer ctrl.Close()

	stream := ctrl.Stream()

	received := make(chan int, 1)

	stream.Listen(func(data int) {
		received <- data
	}, nil)

	time.Sleep(10 * time.Millisecond)
	ctrl.Add(123)

	select {
	case val := <-received:
		if val != 123 {
			t.Errorf("Expected 123, got %d", val)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for data")
	}
}
