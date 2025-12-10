package streamgo

import "sync"

type StreamController[T any] struct {
	input  chan StreamEvent[T]
	mu     sync.Mutex
	subs   []chan StreamEvent[T]
	closed bool
	wg     sync.WaitGroup
}

func NewStreamController[T any]() *StreamController[T] {
	c := &StreamController[T]{

		input: make(chan StreamEvent[T]),
		subs:  make([]chan StreamEvent[T], 0),
	}

	go c.run()
	return c
}

func (c *StreamController[T]) run() {

	c.wg.Add(1)
	defer c.wg.Done()

	for event := range c.input {
		c.mu.Lock()

		for _, ch := range c.subs {
			// **BACKPRESSURE HANDLING (NON-BLOCKING):**
			select {
			case ch <- event:
				// Message sent successfully
			default:
				// The subscriber's buffer is full. We discard the message to
				// avoid blocking the 'run' goroutine and other subscribers.
			}
		}
		c.mu.Unlock()
	}

	// Closing all subscriber channels.
	c.mu.Lock()
	for _, ch := range c.subs {
		close(ch)
	}
	c.mu.Unlock()
}

func (c *StreamController[T]) Stream() *Stream[T] {

	const bufferSize = 2
	output := make(chan StreamEvent[T], bufferSize)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.subs = append(c.subs, output)

	return &Stream[T]{
		output:     output,
		controller: c,
	}
}

func (c *StreamController[T]) Add(data T) {
	if c.closed {
		return
	}

	event := StreamEvent[T]{
		Data:    data,
		Err:     nil,
		IsError: false,
	}
	c.input <- event
}

func (c *StreamController[T]) AddError(err error) {
	if c.closed {
		return
	}

	event := StreamEvent[T]{

		Err:     err,
		IsError: true,
	}
	c.input <- event
}

func (c *StreamController[T]) ForceClose() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}
	c.closed = true

	close(c.input)
}

func (c *StreamController[T]) Close() {
	c.ForceClose()
	c.Wait()
}

func (c *StreamController[T]) Wait() {
	c.wg.Wait()
}
