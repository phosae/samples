package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

/*

go run go run detect-cpu-deadlock.go
go tool pprof http://localhost:6060/debug/pprof/goroutine

Fetching profile over HTTP from http://localhost:6060/debug/pprof/goroutine
Saved profile in /Users/user/pprof/pprof.detect-cpu-deadlock.goroutine.001.pb.gz
File: detect-cpu-deadlock
Type: goroutine
Time: Jan 16, 2025 at 1:13pm (CST)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 6, 100% of 6 total
Showing top 10 nodes out of 39
      flat  flat%   sum%        cum   cum%
         5 83.33% 83.33%          5 83.33%  runtime.gopark
         1 16.67%   100%          1 16.67%  runtime.goroutineProfileWithLabels
         0     0%   100%          1 16.67%  internal/poll.(*FD).Accept
         0     0%   100%          1 16.67%  internal/poll.(*FD).Read
         0     0%   100%          2 33.33%  internal/poll.(*pollDesc).wait
         0     0%   100%          2 33.33%  internal/poll.(*pollDesc).waitRead (inline)
         0     0%   100%          2 33.33%  internal/poll.runtime_pollWait
         0     0%   100%          1 16.67%  main.main
         0     0%   100%          1 16.67%  main.main.func1
         0     0%   100%          1 16.67%  main.process1
(pprof) web
*/

var resourceA sync.Mutex
var resourceB sync.Mutex

func process1(wg *sync.WaitGroup) {
	defer wg.Done()

	resourceA.Lock()
	fmt.Println("Process 1 acquired resource A")

	time.Sleep(1 * time.Second) // Simulate some work

	resourceB.Lock()
	fmt.Println("Process 1 acquired resource B")

	// ... do something with resourceA and resourceB ...

	resourceB.Unlock()
	resourceA.Unlock()
}

func process2(wg *sync.WaitGroup) {
	defer wg.Done()

	resourceB.Lock()
	fmt.Println("Process 2 acquired resource B")

	time.Sleep(1 * time.Second) // Simulate some work

	resourceA.Lock()
	fmt.Println("Process 2 acquired resource A")

	// ... do something with resourceA and resourceB ...

	resourceA.Unlock()
	resourceB.Unlock()
}

func main() {
	// Start pprof server
	go func() {
		fmt.Println("pprof server started at http://localhost:6060")
		http.ListenAndServe("localhost:6060", nil)
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go process1(&wg)
	go process2(&wg)

	wg.Wait()
	fmt.Println("Done")
}
