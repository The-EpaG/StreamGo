package streamgo

type StreamSubscription struct {
	Done chan struct{}
}

// NewStreamSubscription creates a new subscription.
// It's public for external use, although it's mainly for internal package use.
func NewStreamSubscription() *StreamSubscription {
	return &StreamSubscription{
		Done: make(chan struct{}),
	}
}

// Cancel (similar to cancel() in Dart) stops the subscriber from listening.
func (s *StreamSubscription) Cancel() {
	select {
	case <-s.Done:
	default:
		close(s.Done)
	}
}
