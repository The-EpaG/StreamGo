package streamgo

import "sync"

type StreamSubscription struct {
	Done     chan struct{}
	once     sync.Once
	onCancel func()
}

// NewStreamSubscription creates a new subscription.
// It's public for external use, although it's mainly for internal package use.
func NewStreamSubscription(onCancel func()) *StreamSubscription {
	return &StreamSubscription{
		Done:     make(chan struct{}),
		onCancel: onCancel,
	}
}

// Cancel (similar to cancel() in Dart) stops the subscriber from listening.
func (s *StreamSubscription) Cancel() {
	s.once.Do(func() {
		close(s.Done)

		if s.onCancel != nil {
			s.onCancel()
		}
	})
}
