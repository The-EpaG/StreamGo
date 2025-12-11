package streamgo

// StreamEvent[T] encapsulates the value (of type T) or an error for the communication flow.
type StreamEvent[T any] struct {
	Data T
	Err  error
}

type Stream[T any] struct {
	controller *StreamController[T]
	bufferSize int
}

// Listen starts listening on the Stream, using a callback function.
// onData is the callback for data (type T), onError is the callback for errors.
// It returns a StreamSubscription object for lifecycle management.
// The WaitGroup is managed internally by the Controller.
func (s *Stream[T]) Listen(onData func(T), onError func(error)) *StreamSubscription {
	// Register with the controller to get a new subscriber
	sub := s.controller.subscribe(s.bufferSize)
	if sub == nil {
		// Controller is closed or max subscribers reached.
		// Return a closed subscription (or nil? Pattern usually returns object).
		// Returning a dummy subscription that is already "done" is safer.
		dummy := NewStreamSubscription(nil)
		dummy.Cancel()
		return dummy
	}

	cleanupFunc := func() {
		s.controller.removeSubscriber(sub)
	}

	streamSub := NewStreamSubscription(cleanupFunc)

	s.controller.wg.Go(func() {
		for {
			select {
			case event, ok := <-sub.ch:
				if !ok {
					// End of Stream (Controller.Close())
					return
				}
				if event.Err != nil {
					if onError != nil {
						onError(event.Err)
					}
				} else {
					if onData != nil {
						onData(event.Data)
					}
				}
			case <-streamSub.Done:
				return
			}
		}
	})

	return streamSub
}
