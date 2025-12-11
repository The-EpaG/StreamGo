# StreamGo

StreamGo is a simple Go library that emulates the logic of Dart's Streams, providing a convenient way to handle sequences of asynchronous events.

## Features

-   **Generic Streams**: Create streams of any data type using Go generics.
-   **Error Handling**: Propagate and handle errors within the stream.
-   **Subscription Management**: Cancel subscriptions to stop listening for events.
-   **Backpressure Handling**: Robust blocking backpressure to ensure zero data loss when subscribers are slow.
-   **Configurable Buffering**: detailed control over buffer sizes per stream.

## Installation

To use StreamGo in your project, you can use `go get`:

```bash
go get github.com/The-EpaG/StreamGo
```

## Testing

To run the tests for this project, you can use the standard Go test command.

### Run all tests
```bash
go test -v ./pkg/streamgo/...
```

### Run tests with race detection
It is recommended to run tests with the `-race` flag to ensure concurrency safety.
```bash
go test -v -race ./pkg/streamgo/...
```

## Usage

Here is a simple example of how to use StreamGo:

```go
package main

import (
    "fmt"
    "time"

    "github.com/The-EpaG/StreamGo/pkg/streamgo"
)

func main() {
    // 1. Create a new StreamController
    controller := streamgo.NewStreamController[int]()

    // 2. Get a Stream from the controller
    // You can specify a buffer size (e.g., 10) or use default (Stream())
    stream := controller.StreamWithBuffer(10)

    // 3. Listen to the stream
    subscription := stream.Listen(
        func(data int) {
            fmt.Printf("Received data: %d\n", data)
        },
        func(err error) {
            fmt.Printf("Received error: %v\n", err)
        },
    )

    // 4. Add data and errors to the stream
    // This will BLOCK if the buffer is full!
    controller.Add(1)
    controller.Add(2)
    controller.AddError(fmt.Errorf("something went wrong"))
    controller.Add(3)

    // 5. Close the stream when you're done
    controller.Close()

    // You can also cancel a subscription
    subscription.Cancel()
}
```

### Example with a custom struct

```go
package main

import (
    "fmt"

    "github.com/The-EpaG/StreamGo/pkg/streamgo"
)

type UserMetric struct {
    UserID string
    Value  float64
    Event  string
}

func main() {
    fmt.Println("--- Stream of Structs (UserMetric) ---")
    metricController := streamgo.NewStreamController[UserMetric]()
    metricStream := metricController.Stream()

    metricStream.Listen(
        func(metric UserMetric) { // T is UserMetric
            fmt.Printf("   [METRIC Listener] Received metric: %s for User %s\n", metric.Event, metric.UserID)
        },
        func(err error) { // onError
            fmt.Printf("   [METRIC Listener] 🚨 RECEIVED ERROR: %s\n", err.Error())
        },
    )

    metricController.Add(UserMetric{UserID: "A101", Value: 5.4, Event: "Login"})
    metricController.Add(UserMetric{UserID: "B202", Value: 12.1, Event: "Logout"})
    metricController.AddError(fmt.Errorf("critical metric error"))
    metricController.Close()
}
```

## API Overview

### `StreamController[T]`

The `StreamController` is the main entry point for creating and managing streams.

-   `NewStreamController[T]() *StreamController[T]`: Creates a new stream controller.
-   `Stream() *Stream[T]`: Returns a new stream with default buffering.
-   `StreamWithBuffer(size int) *Stream[T]`: Returns a new stream with specified buffer size.
-   `Add(data T)`: Adds data to the stream. Blocks if any subscriber's buffer is full.
-   `AddError(err error)`: Adds an error to the stream. Blocks if any subscriber's buffer is full.
-   `Close()`: Closes the stream and waits for all listeners to finish.
-   `ForceClose()`: Closes the stream without waiting for listeners.
-   `Wait()`: Waits for all listeners to finish.

### `Stream[T]`

The `Stream` represents a sequence of asynchronous events.

-   `Listen(onData func(T), onError func(error)) *StreamSubscription`: Registers a listener for data and error events.

### `StreamSubscription`

The `StreamSubscription` represents a subscription to a stream.

-   `Cancel()`: Cancels the subscription.

## License

This project is licensed under the GPL-3.0 License - see the [LICENSE](LICENSE) file for details.
