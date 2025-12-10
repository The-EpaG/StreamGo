package streamgo

type Stream[T any] struct {
	output     chan StreamEvent[T]
	controller *StreamController[T]
}

// Listen starts listening on the Stream, using a callback function.
// onData is the callback for data (type T), onError is the callback for errors.
// It returns a StreamSubscription object for lifecycle management.
// The WaitGroup is managed internally by the Controller.
func (s *Stream[T]) Listen(onData func(T), onError func(error)) *StreamSubscription {
	sub := NewStreamSubscription()

	s.controller.wg.Add(1)

	go func() {
		defer s.controller.wg.Done()

		for {
			select {
			case event, ok := <-s.output:
				if !ok {
					// End of Stream (Controller.Close())
					return
				}
				if event.IsError {
					if onError != nil {
						onError(event.Err)
					}
				} else {
					onData(event.Data)
				}
			case <-sub.Done:
				return
			}
		}
	}()

	return sub
}
