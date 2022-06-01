package main

import (
	"fmt"
	"github.com/nats-io/stan.go"
	"strconv"
)

func main() {
	sc, err := stan.Connect("test-cluster", "client123")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Simple Async Subscriber
	for i := 1; i < 100; i++ {
		sc.Publish("data", []byte(strconv.Itoa(i)))
	}

	// Close connection
	sc.Close()
}
