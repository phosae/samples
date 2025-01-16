package main

/*
#cgo CFLAGS: -g -Wall -I.
#cgo LDFLAGS: -L. -lleaky
#include <stdlib.h>
#include "leaky.h"
*/
import "C"

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
	"unsafe"
)

func main() {
	// Enable profiling
	go func() {
		fmt.Println("pprof server started at http://localhost:6060/debug/pprof/")
		http.ListenAndServe("localhost:6060", nil)
	}()

	http.HandleFunc("/leaky", leakyHandler)
	http.HandleFunc("/fixed", fixedHandler)
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

func leakyHandler(w http.ResponseWriter, r *http.Request) {
	userName := C.CString("Leaky User")
	defer C.free(unsafe.Pointer(userName))

	user := C.createUser(userName, 123)

	// Simulate using the user object
	C.printUser(user)

	// We "free" the user, but due to the bug in freeUser, a leak occurs
	C.freeUser(user)

	fmt.Fprintf(w, "Processed leaky request\n")
}

func fixedHandler(w http.ResponseWriter, r *http.Request) {
	userName := C.CString("Fixed User")
	defer C.free(unsafe.Pointer(userName))

	user := C.createUser(userName, 456)

	// Simulate using the user object
	C.printUser(user)

	// Correctly free the user with the fixed function
	C.freeUserCorrectly(user)

	fmt.Fprintf(w, "Processed fixed request\n")
}

/*
gcc -g -Wall -shared -fPIC -o libleaky.so leaky.c
go build -o cgo_leaky_example main.go


hey -n 1000 -c 50 -z 1m http://localhost:8080/leaky
go tool pprof http://localhost:6060/debug/pprof/heap

pprof primarily shows Go-managed memory.
It might indirectly hint at C leaks if you see continuously growing memory usage correlated with CGo calls,
but it doesn't give you precise details about C-allocated memory.

(pprof) top
Showing nodes accounting for 514kB, 100% of 514kB total
      flat  flat%   sum%        cum   cum%
     514kB   100%   100%      514kB   100%  bufio.NewWriterSize (inline)
         0     0%   100%      514kB   100%  net/http.(*conn).serve
         0     0%   100%      514kB   100%  net/http.newBufioWriterSize

# macOS amd64
brew tap LouisBrunner/valgrind
brew install --HEAD LouisBrunner/valgrind/valgrind

# Linux
apt install valgrind

LD_LIBRARY_PATH=$LD_LIBRARY_PATH:. valgrind --leak-check=full --show-leak-kinds=all ./cgo_leaky_example

LD_LIBRARY_PATH=$LD_LIBRARY_PATH:. ./cgo_leaky_example

///~
...
C: User: Leaky User, id: 123
C: Freeing user at 0x697e7610
Alloc = 3 MiB	TotalAlloc = 92 MiB	Sys = 16 MiB	NumGC = 35
Alloc = 3 MiB	TotalAlloc = 92 MiB	Sys = 16 MiB	NumGC = 35
Alloc = 3 MiB	TotalAlloc = 92 MiB	Sys = 16 MiB	NumGC = 35
Alloc = 3 MiB	TotalAlloc = 92 MiB	Sys = 16 MiB	NumGC = 35
Alloc = 0 MiB	TotalAlloc = 92 MiB	Sys = 16 MiB	NumGC = 36
...
==18099== Process terminating with default action of signal 2 (SIGINT)
==18099==    at 0x47B268: runtime.raise.abi0 (sys_linux_arm64.s:158)
==18099==
==18099== HEAP SUMMARY:
==18099==     in use at exit: 957,696 bytes in 58,858 blocks
==18099==   total heap usage: 176,525 allocs, 117,667 frees, 2,254,748 bytes allocated
==18099==
==18099== 160 bytes in 10 blocks are definitely lost in loss record 1 of 120
==18099==    at 0x4885250: malloc (in /usr/libexec/valgrind/vgpreload_memcheck-arm64-linux.so)
==18099==    by 0x48C081F: createUser (leaky.c:8)
==18099==    by 0x6514C7: _cgo_6b7719a2a370_Cfunc_createUser (cgo-gcc-prolog:56)
==18099==    by 0x47A22B: runtime.asmcgocall.abi0 (asm_arm64.s:1000)
==18099==    by 0x40002301BF: ???
==18099==
==18099== 160 bytes in 10 blocks are definitely lost in loss record 2 of 120
==18099==    at 0x4885250: malloc (in /usr/libexec/valgrind/vgpreload_memcheck-arm64-linux.so)
==18099==    by 0x48C081F: createUser (leaky.c:8)
==18099==    by 0x6514C7: _cgo_6b7719a2a370_Cfunc_createUser (cgo-gcc-prolog:56)
==18099==    by 0x47A22B: runtime.asmcgocall.abi0 (asm_arm64.s:1000)
==18099==    by 0x40004CD33F: ???
*/
