package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func leakyFunction() {
	timer := time.NewTimer(5 * time.Second)
	// defer timer.Stop() // Stop the timer when the function exits

	go func() {
		<-timer.C // When this channel is closed after 5s, it can send.
		fmt.Println("Timer fired")
	}()

	time.Sleep(60 * time.Second)
}

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	for i := 0; i < 1_000_000; i++ {
		go leakyFunction()
	}

	time.Sleep(60 * time.Second) // Keep the program running for profiling
}
