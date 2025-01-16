package main

/*
go tool pprof http://localhost:6060/debug/pprof/goroutine

Fetching profile over HTTP from http://localhost:6060/debug/pprof/goroutine
Saved profile in /Users/root/pprof/pprof.goroutine-blocked-on-channel.goroutine.002.pb.gz
File: goroutine-blocked-on-channel
Type: goroutine
Time: Jan 16, 2025 at 12:37pm (CST)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 1000003, 100% of 1000004 total
Dropped 30 nodes (cum <= 5000)
      flat  flat%   sum%        cum   cum%
   1000003   100%   100%    1000003   100%  runtime.gopark
         0     0%   100%    1000000   100%  main.leakySend.func1
         0     0%   100%    1000000   100%  runtime.chansend
         0     0%   100%    1000000   100%  runtime.chansend1
*/
import (
	"net/http"
	_ "net/http/pprof"
	"time"
)

func leakySend() {
	ch := make(chan int) // Unbuffered channel

	go func() {
		ch <- 10 // This will block indefinitely if there's no receiver
	}()

	// The goroutine is now blocked, and any memory it holds won't be released
}

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	for i := 0; i < 1_000_000; i++ {
		leakySend()
	}
	time.Sleep(30 * time.Second)
}
