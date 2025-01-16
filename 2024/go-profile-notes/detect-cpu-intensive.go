package main

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"
)

/*
Example 1:

	go run detect-cpu-high.go
	go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile

Example 2:

	go tool pprof http://localhost:6060/debug/pprof/profile

	Fetching profile over HTTP from http://localhost:6060/debug/pprof/profile
	Saved profile in /Users/user/pprof/pprof.detect-cpu-high.samples.cpu.002.pb.gz
	(pprof) top
	Showing nodes accounting for 24.24s, 100% of 24.25s total
	Dropped 8 nodes (cum <= 0.12s)
	Showing top 10 nodes out of 13
      flat  flat%   sum%        cum   cum%
    23.48s 96.82% 96.82%     24.08s 99.30%  main.sortRandomly (inline)
     0.60s  2.47% 99.30%      0.60s  2.47%  runtime.asyncPreempt
     0.16s  0.66%   100%      0.16s  0.66%  runtime.kevent
         0     0%   100%     24.08s 99.30%  main.computationallyIntensiveTask
         0     0%   100%      0.14s  0.58%  runtime.(*timer).maybeAdd
         0     0%   100%      0.14s  0.58%  runtime.(*timer).modify
         0     0%   100%      0.14s  0.58%  runtime.(*timer).reset (inline)
         0     0%   100%      0.17s   0.7%  runtime.mcall
         0     0%   100%      0.14s  0.58%  runtime.netpollBreak (inline)
         0     0%   100%      0.17s   0.7%  runtime.park_m
	(pprof) web  # This will generate an SVG file of the graph using your local Graphviz installation and attempt to open it in your default web browser.
*/

func main() {
	// Start pprof server
	go func() {
		fmt.Println("pprof server started at http://localhost:6060")
		http.ListenAndServe("localhost:6060", nil)
	}()

	// Simulate some work that leads to high CPU usage
	go computationallyIntensiveTask()

	// Keep the main goroutine alive
	fmt.Println("Application running... Press Ctrl+C to exit.")
	select {}
}

func computationallyIntensiveTask() {
	for {
		// Simulate a task that involves random sorting, which can be CPU-intensive
		data := generateRandomData(10000)
		sortRandomly(data)

		// Introduce a small delay to avoid pegging the CPU at 100% constantly
		// Remove this in a real scenario to see maximum CPU usage
		time.Sleep(10 * time.Millisecond)
	}
}

// generateRandomData generates a slice of random integers
func generateRandomData(size int) []int {
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = rand.Intn(size)
	}
	return data
}

// sortRandomly sorts the data using a bubble sort algorithm for demonstration
func sortRandomly(data []int) {
	for i := 0; i < len(data); i++ {
		for j := i + 1; j < len(data); j++ {
			if data[i] > data[j] {
				data[i], data[j] = data[j], data[i]
			}
		}
	}
}
