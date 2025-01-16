package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

/*
This is because the response bodies are not being closed,
so the underlying connections and associated resources are not being released and are not eligible for garbage collection.
You might also see a growing number of goroutines in the goroutine profile.

go run detect-http-conn-leak.go
go tool pprof http://localhost:6060/debug/pprof/goroutine                                          ─╯
*/

func leakyHTTPClient() {
	for i := 0; i < 10_000; i++ {
		resp, err := http.Get("http://localhost:6060/debug/pprof/") // Accessing pprof endpoint as an example
		if err != nil {
			log.Println("Error:", err)
			continue
		}

		// Simulate doing something with the response...

		// LEAK: We forgot to close the response body
		// resp.Body.Close()

		fmt.Println("Request:", i, "Status:", resp.Status)
	}
}

func main() {
	// Start pprof server
	go func() {
		log.Fatalln(http.ListenAndServe("localhost:6060", nil))
	}()
	time.Sleep(100 * time.Millisecond)

	// Run the leaky function
	go leakyHTTPClient()

	// Keep the program running for a while to observe the leak
	fmt.Println("Running... Press Ctrl+C to exit.")
	time.Sleep(300 * time.Second)
}
