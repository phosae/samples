package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	rpprof "runtime/pprof"
	"runtime/trace"
	"sync"
)

var s *bool = flag.Bool("s", false, "if true, start web service")

func main() {
	flag.Parse()

	if !*s {
		TraceManually()
		PprofManually()
		return
	}
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, 1024*1024)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Allocated %d bytes", len(data))))
	}))

	addr := ":8080"
	fmt.Printf("listen and serve on %s\n", addr)
	http.ListenAndServe(addr, http.DefaultServeMux)
}

func TraceManually() {
	f, _ := os.Create("trace.out")
	defer f.Close()
	trace.Start(f)
	defer trace.Stop()

	DoTasks()
}

func PprofManually() {
	f, _ := os.Create("cpu.prof")
	defer f.Close()
	rpprof.StartCPUProfile(f)
	defer rpprof.StopCPUProfile()

	Calculate()
}

func DoTasks() {
	var total int
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1e6; j++ {
				total++
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

var generators [20]*rand.Rand

func init() {
	for i := int64(0); i < 20; i++ {
		generators[i] = rand.New(rand.NewSource(i).(rand.Source64))
	}
}

type gen int

//go:noinline
func (g gen) readNumber() int {
	return generators[int(g)].Intn(10)
}

func Calculate() {
	var total int
	var wg sync.WaitGroup

	for i := gen(0); i < 20; i++ {
		wg.Add(1)
		go func(g gen) {
			for j := 0; j < 1e7; j++ {
				total += g.readNumber()
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
