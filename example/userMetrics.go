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
	fmt.Println("--- Start of streamgo library example ---")

	// --- Example 1: Stream of integers (int) ---
	fmt.Println("\n--- Stream of integers ---")
	intController := streamgo.NewStreamController[int]()
	intStream := intController.Stream()

	intStream.Listen(
		func(i int) { // T is int
			fmt.Printf("   [INT Listener] Received DATA: %d\n", i)
		},
		func(err error) { // onError
			fmt.Printf("   [INT Listener] 🚨 RECEIVED ERROR: %s\n", err.Error())
		},
	)

	intController.Add(10)
	intController.Add(20)
	intController.AddError(fmt.Errorf("int sample error"))
	intController.Close()

	// --- Example 2: Stream of Structs (UserMetric) ---
	fmt.Println("\n--- Stream of Structs (UserMetric) ---")
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

	fmt.Println("\n--- streamgo library example finished. ---")
}
