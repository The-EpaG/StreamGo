package streamgo

import "sync"

type Subscriber[T any] struct {
	ch   chan StreamEvent[T]
	quit chan struct{} // Closed when subscription is cancelled
}

type StreamController[T any] struct {
	input     chan StreamEvent[T]
	mu        sync.Mutex
	subs      []*Subscriber[T]
	wg        sync.WaitGroup
	done      chan struct{}
	closeOnce sync.Once

	// Default buffer size for new streams if not specified
	defaultBufferSize int
}

func NewStreamController[T any]() *StreamController[T] {
	c := &StreamController[T]{
		input:             make(chan StreamEvent[T]),
		subs:              make([]*Subscriber[T], 0),
		done:              make(chan struct{}),
		defaultBufferSize: 2, // Default small buffer
	}

	c.run()
	return c
}

const MaxSubscribers = 128

func (c *StreamController[T]) run() {
	c.wg.Go(func() {
		defer func() {
			// Controller closed
			// Closing all subscriber channels.
			c.mu.Lock()
			for _, sub := range c.subs {
				close(sub.ch)
			}
			c.subs = nil // Prevent double-close in removeSubscriber
			c.mu.Unlock()
		}()

		for {
			select {
			case event := <-c.input:
				c.mu.Lock()
				// Snapshot subscribers to release lock during blocking send?
				// NO. If we release lock, removeSubscriber might happen concurrently.
				// However, blocking with lock held prevents new subscribers from joining (Add/Remove blocked).
				// But blocking with lock held is safer for list integrity.
				// BUT if we block, Add() call at source is blocked. That's fine.
				// Wait, if we hold lock, removeSubscriber cannot acquire lock.
				// If removeSubscriber cannot acquire lock, it cannot close 'quit'.
				// If 'quit' is not closed, we deadlock here waiting for 'ch' which is full,
				// and 'removeSubscriber' waits for us to release lock.
				// **DEADLOCK DANGER**.

				// SOLUTION: Copy subscribers and release lock.
				currentSubs := make([]*Subscriber[T], len(c.subs))
				copy(currentSubs, c.subs)
				c.mu.Unlock()

				for _, sub := range currentSubs {
					// **BACKPRESSURE HANDLING (BLOCKING):**
					// We block until the subscriber accepts the event OR cancels.
					select {
					case sub.ch <- event:
					case <-sub.quit:
						// Subscriber cancelled, stop waiting for this subscriber
					}
				}

			case <-c.done:
				return
			}
		}
	})
}

// subscribe creates a new Subscriber and adds it to the list.
func (c *StreamController[T]) subscribe(bufferSize int) *Subscriber[T] {
	// Check if the controller is already closed before proceeding.
	select {
	case <-c.done:
		return nil
	default:
	}

	if bufferSize < 0 {
		bufferSize = c.defaultBufferSize
	}

	output := make(chan StreamEvent[T], bufferSize)
	sub := &Subscriber[T]{
		ch:   output,
		quit: make(chan struct{}),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring the lock
	select {
	case <-c.done:
		// Clean up potentially created resources
		close(output)
		return nil
	default:
		if len(c.subs) >= MaxSubscribers {
			close(output)
			return nil
		}
		c.subs = append(c.subs, sub)
	}

	return sub
}

// Stream creates a new Stream with default buffering.
func (c *StreamController[T]) Stream() *Stream[T] {
	return c.StreamWithBuffer(-1) // Use default
}

// StreamWithBuffer creates a new Stream with specified buffer size.
func (c *StreamController[T]) StreamWithBuffer(bufferSize int) *Stream[T] {
	return &Stream[T]{
		controller: c,
		bufferSize: bufferSize,
	}
}

func (c *StreamController[T]) Add(data T) {
	event := StreamEvent[T]{
		Data: data,
		Err:  nil,
	}

	select {
	case c.input <- event:
		// Event sent successfully.
	case <-c.done:
		// Controller is closed, do nothing.
	}
}

func (c *StreamController[T]) AddError(err error) {
	event := StreamEvent[T]{
		Err: err,
	}

	select {
	case c.input <- event:
		// Event sent successfully.
	case <-c.done:
		// Controller is closed, do nothing.
	}
}

func (c *StreamController[T]) ForceClose() {
	c.closeOnce.Do(func() {
		close(c.done)
		// close(c.input) // Do NOT close input to avoid panic in Add/AddError
	})
}

func (c *StreamController[T]) Close() {
	c.ForceClose()
	c.Wait()
}

func (c *StreamController[T]) Wait() {
	c.wg.Wait()
}

func (c *StreamController[T]) removeSubscriber(sub *Subscriber[T]) {
	// Signal the 'run' loop to stop waiting for this subscriber
	// We must do this BEFORE acquiring the lock if we wanted to break a deadlock,
	// BUT 'quit' is only read by 'run' loop which doesn't hold the lock during send.
	// So we can do it safely.

	// Close quit to unblock any pending send in run() loop
	// We use strictly one-time close via select or just ignored if double close?
	// Better to just close it. But closing closed channel panics.
	// Let's assume removeSubscriber is called once per subscription.
	// Actually, StreamSubscription.Cancel() uses sync.Once, so this is safe.
	close(sub.quit)

	c.mu.Lock()
	defer c.mu.Unlock()

	for i, s := range c.subs {
		if s == sub {
			c.subs = append(c.subs[:i], c.subs[i+1:]...)
			// close(s.ch) // DO NOT Close the data channel here. It causes a race with the run loop sending to it.
			// The run loop handles stale subscribers via the 'quit' channel.
			// The listener exits via streamSub.Done checks.
			break
		}
	}
}
