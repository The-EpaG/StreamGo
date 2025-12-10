package streamgo

// StreamEvent[T] encapsulates the value (of type T) or an error for the communication flow.
type StreamEvent[T any] struct {
	Data    T
	Err     error
	IsError bool
}
