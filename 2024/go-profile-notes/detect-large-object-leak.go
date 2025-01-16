package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"time"
)

/*
go run detect-large-object-leak.go

hey -n 100 -c 30 http://localhost:8080/

go tool pprof http://localhost:6060/debug/pprof/heap

(pprof) top
Showing nodes accounting for 1.56GB, 99.94% of 1.56GB total
Dropped 14 nodes (cum <= 0.01GB)
      flat  flat%   sum%        cum   cum%
    1.56GB 99.94% 99.94%     1.56GB 99.94%  main.handler
         0     0% 99.94%     1.56GB 99.94%  net/http.(*ServeMux).ServeHTTP
         0     0% 99.94%     1.56GB 99.94%  net/http.(*conn).serve
         0     0% 99.94%     1.56GB 99.94%  net/http.HandlerFunc.ServeHTTP
         0     0% 99.94%     1.56GB 99.94%  net/http.serverHandler.ServeHTTP

we've found problem in `main.handler`

(pprof) list main.handler
Total: 880MB
ROUTINE ======================== main.handler in /Users/xu/go/src/github.com/phosae/samples/2024/go-profile-notes/detect-large-object-leak.go
     880MB      880MB (flat, cum)   100% of Total
         .          .     50:func handler(w http.ResponseWriter, r *http.Request) {
         .          .     51:   // Simulate processing that creates a large object and stores it in a map
     880MB      880MB     52:   largeObject := make([]byte, 10*1024*1024) // 10MB
         .          .     53:
         .          .     54:   counterMu.Lock()
         .          .     55:   currentCounter := counter
         .          .     56:   counter++
         .          .     57:   counterMu.Unlock()

*/

type LeakyData struct {
	data map[int][]byte
	mu   sync.Mutex
}

var leakyData = LeakyData{
	data: make(map[int][]byte),
}
var counter = 0
var counterMu sync.Mutex

func handler(w http.ResponseWriter, r *http.Request) {
	// Simulate processing that creates a large object and stores it in a map
	largeObject := make([]byte, 10*1024*1024) // 10MB

	counterMu.Lock()
	currentCounter := counter
	counter++
	counterMu.Unlock()

	leakyData.mu.Lock()
	leakyData.data[currentCounter] = largeObject
	leakyData.mu.Unlock()

	// Simulate some work (in real scenario, the large object might be used here)
	time.Sleep(100 * time.Millisecond)

	// Potential fix: delete the entry after use
	//leakyData.mu.Lock()
	//delete(leakyData.data, currentCounter)
	//leakyData.mu.Unlock()

	fmt.Fprintf(w, "Processed request %d\n", currentCounter)

	// In this example, the largeObject is never explicitly released
	// and remains in the leakyData map.
}

func main() {
	// Enable profiling
	go func() {
		fmt.Println("pprof server started at http://localhost:6060/debug/pprof/")
		http.ListenAndServe("localhost:6060", nil)
	}()

	// Start a simple HTTP server
	http.HandleFunc("/", handler)
	fmt.Println("Server started at http://localhost:8080/")
	go http.ListenAndServe(":8080", nil)

	// Periodically print memory stats
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
		fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
		fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
		fmt.Printf("\tNumGC = %v\n", m.NumGC)
		time.Sleep(5 * time.Second)
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
