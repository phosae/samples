# How to profile memory

Ask: how to figure out Golang appplication's memory leak ?

Answer by Gemini:

Memory leaks in Go applications, while less common than in languages with manual memory management, 
can still occur due to lingering references that prevent the garbage collector from reclaiming unused memory. 
Here's a comprehensive guide on how to identify and fix memory leaks in your Go code:

**1. Understanding Memory Leaks in Go**

*   **Garbage Collection:** Go has automatic garbage collection (GC), which periodically identifies and reclaims memory that is no longer reachable.
*   **How Leaks Happen:** Memory leaks in Go typically occur when you unintentionally keep references to objects that you no longer need. This prevents the GC from reclaiming that memory.
*   **Common Causes:**
    *   **Global Variables:** Global variables holding references to large data structures that are no longer used.
    *   **Long-Lived Goroutines:** Goroutines that are blocked indefinitely or have unintentional references to large objects, preventing them from being garbage-collected.
    *   **Unclosed Channels:**  A goroutine blocked on sending to or receiving from a channel that is never closed or read from.
    *   **Timers/Tickers:** `time.Timer` or `time.Ticker` that are not stopped when no longer needed.
    *   **Cgo:** Improper handling of memory allocated in C code when using Cgo.
    *   **Slices:** Appending to slices within loops without considering the underlying array's potential to grow unnecessarily large.
    *   **Closures:** Unintentionally capturing large variables in closures.
    *   **Finalizers:** Using `runtime.SetFinalizer` incorrectly. It's generally recommended to avoid finalizers unless absolutely necessary.

**2. Tools for Detecting Memory Leaks**

*   **`pprof` (Profiling):**
    *   **Gold Standard:** `pprof` is the most important tool for profiling Go applications and identifying memory leaks.
    *   **Heap Profiles:** It can collect heap profiles that show you the memory allocation patterns in your application.
    *   **How to Use:**
        1.  **Import:** Add `import _ "net/http/pprof"` to your main package. This automatically registers pprof's handlers with the default HTTP server.
        2.  **Start HTTP Server (if not already running):** If your application doesn't already have an HTTP server, you can start one:
            ```go
            go func() {
                log.Println(http.ListenAndServe("localhost:6060", nil))
            }()
            ```
        3.  **Collect a Heap Profile:**
            *   **Option 1 (Interactive):**  Open your web browser and go to `http://localhost:6060/debug/pprof/`. Click on "heap" and let it collect data. You can then analyze interactively on the UI, or download the file
            *   **Option 2 (Command Line):**
                ```bash
                go tool pprof http://localhost:6060/debug/pprof/heap
                ```
                This will download the heap profile and start an interactive `pprof` session in your terminal.
        4.  **Analyze the Profile:**
            *   **`top`:** Shows the functions that allocated the most memory.
            *   **`web`:** Generates a visual graph (if you have Graphviz installed) showing allocation hotspots. This command will try to open a web browser to display the graph.
                ```bash
                # Install Graphviz on macOS using Homebrew:
                brew install graphviz

                # Install Graphviz on Ubuntu/Debian:
                sudo apt-get install graphviz

                # Install Graphviz on Windows:
                # Download from https://graphviz.org/download/ and add to PATH.
                ```
            *   **`list <function>`:** Shows the source code of a function with memory allocation annotations.
            *   **`inuse_space` (default):** Focuses on memory that is currently in use.
            *   **`alloc_space`:** Shows the total memory allocated, even if it has been garbage collected. This can help find places where you're allocating a lot of memory that's quickly discarded (churn).

*   **Go Test with `-memprofile`:**
    *   **Targeted Profiling:** You can collect memory profiles during test runs, making it easier to isolate leaks related to specific test cases.
    *   **Example:**
        ```bash
        go test -memprofile memprofile.out ./your_package
        go tool pprof memprofile.out
        ```
<!-- fix Gemini outdate -->
<!-- go-torch is deprecated, Go 1.11 later we use `go tool pprof`, others can use `go get -u github.com/google/pprof`. -->
*   **`go tool pprof` (Flame Graphs):**
    *   **Visual Analysis:**  `go tool pprof` generates flame graphs from `pprof` profiles. Flame graphs provide an intuitive, visual way to understand memory allocation patterns.
    *   **Usage:**
        1.  Collect a `pprof` heap profile (see above).
        2.  Run `go-torch`:
            ```bash
            # This will listen on :8081 and open a browser.
            # Change :8081 to a port of your choice.
            $ go tool pprof -http=":8081" [binary] [profile]
            ```
        This will typically open your web browser with an interactive flame graph.

*   **Delve (Debugger):**
    *   **Interactive Debugging:** Delve is a powerful debugger for Go. While not primarily for memory leaks, it can help inspect memory usage at specific points in your program.
    *   **Installation:**
        ```bash
        go install github.com/go-delve/delve/cmd/dlv@latest
        ```
    *   **Usage:** You can use Delve to set breakpoints, step through code, and inspect variables, which might help you identify why memory isn't being released.

**3. Strategies for Investigating with `pprof`**

1.  **Identify Top Allocators:** Use the `top` command in `pprof` to find the functions or parts of your code that are allocating the most memory.
2.  **Focus on `inuse_space`:** By default, `pprof` displays `inuse_space`. This shows you the memory that is currently allocated and not yet garbage collected, which is most relevant for leak detection.
3.  **Visualize with `web` or `go tool pprof`:** Use the `web` command (or `go tool pprof`) to generate a graphical representation of the heap profile. This makes it easier to identify allocation hotspots and follow the flow of memory through your application.
4.  **Inspect Source Code with `list`:** Use the `list` command to examine the source code of functions that appear to be allocating a lot of memory. Look for clues like:
    *   Large objects being created but not explicitly released.
    *   Appending to slices inside loops without pre-allocating or clearing them.
    *   Global variables that might be holding onto references.
5.  **Compare Profiles:** Take multiple heap profiles at different points in time (e.g., before and after a suspected leak). Use the `-base` option in `pprof` to compare profiles and see what has changed:
    ```bash
    # Take the first profile (baseline)
    go tool pprof http://localhost:6060/debug/pprof/heap
    (pprof) top
    (pprof) exit
    # run the test, wait for some time.
    # Take the second profile
    go tool pprof http://localhost:6060/debug/pprof/heap
    (pprof) top
    (pprof) exit
    # Compare the second profile to the first
    go tool pprof -base <baseline_profile_file> <new_profile_file>
    (pprof) top # Show the difference
    ```
6.  **Isolate the Leak:** Try to isolate the part of your code that is causing the leak by creating minimal, reproducible test cases.

**4. Common Leak Patterns and Fixes**

*   **Global Variables:**
    *   **Problem:** If a global variable holds a reference to an object, that object will never be garbage collected, even if it's no longer needed.
    *   **Solution:**
        *   Minimize the use of global variables.
        *   Explicitly set global variables to `nil` when the referenced objects are no longer needed.
        *   Use a `sync.Pool` to reuse objects if appropriate.

*   **Long-Lived Goroutines:**
    *   **Problem:** A goroutine that runs indefinitely but is blocked or unintentionally holding references to large objects can cause a leak.
    *   **Solution:**
        *   Ensure that goroutines have a way to terminate gracefully. Use context cancellation or other signaling mechanisms.
        *   Review goroutines for unintended references to large objects. Make sure they don't capture variables in closures that they don't need.

*   **Unclosed Channels:**
    *   **Problem:** If a goroutine is blocked on a send or receive operation on an unclosed channel, it will never be garbage collected.
    *   **Solution:**
        *   Close channels when they are no longer needed, especially after the last send operation.
        *   Use `select` statements with `default` cases or timeouts to avoid indefinite blocking on channel operations.

*   **Timers/Tickers:**
    *   **Problem:** `time.NewTimer` and `time.NewTicker` create timers that continue to run in the background until explicitly stopped.
    *   **Solution:**
        *   Call `timer.Stop()` or `ticker.Stop()` when the timer or ticker is no longer needed.
        *   Use `defer timer.Stop()` or `defer ticker.Stop()` to ensure they are stopped even if errors occur.

*   **Slices (Appending in Loops):**
    *   **Problem:** Appending to a slice inside a loop can cause the underlying array to be repeatedly reallocated and copied, leading to excessive memory usage.
    *   **Solution:**
        *   If you know the approximate size of the slice in advance, pre-allocate it using `make([]T, 0, capacity)`.
        *   If you need to repeatedly grow and shrink a slice, consider using a `bytes.Buffer` or a `list.List` (from the `container/list` package) instead.
        *   Set slices that are no longer needed to `nil` so that they can be garbage collected.

*   **Closures:**
    *   **Problem:** Closures can capture variables from the enclosing scope. If a closure captures a large variable, that variable won't be garbage collected until the closure itself is garbage collected, even if it's not actively used.
    *   **Solution:**
        *   Be mindful of what variables your closures capture. If possible, pass only the necessary values as arguments to the closure instead of capturing large structures.

*   **Cgo:**
    *   **Problem:** Memory allocated in C code using `malloc` must be explicitly freed using `free`.
    *   **Solution:**
        *   Use `defer C.free(unsafe.Pointer(ptr))` to ensure that C memory is freed as soon as the Go function exits.
        *   Be very careful when passing Go pointers to C code, as the garbage collector might move or deallocate the memory while the C code is still using it.

**5. Example of Fixing a Leak (Unstopped Timer)**

**Leaky Code:**

```go
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func leakyFunction() {
	timer := time.NewTimer(5 * time.Second)
	go func() {
		<-timer.C
		fmt.Println("Timer fired (but I might never be called!)")
	}()

	// ... timer is not stopped, keeping the goroutine alive ...
}

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	for i := 0; i < 100000; i++ {
		leakyFunction()
	}

	time.Sleep(30 * time.Second) // Keep the program running for profiling
}
```

**Fixed Code:**

```go
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func nonLeakyFunction() {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop() // Stop the timer when the function exits

	go func() {
		<-timer.C // When this channel is closed after 5s, it can send.
		fmt.Println("Timer fired")
	}()
}

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	for i := 0; i < 100000; i++ {
		nonLeakyFunction()
	}

	time.Sleep(30 * time.Second) // Keep the program running for profiling
}
```

By systematically using `pprof` and understanding these common leak patterns, you can effectively track down and fix memory leaks in your Go applications, ensuring they run efficiently and reliably. Remember that the key is to understand how the Go garbage collector works and to be mindful of the references you create in your code.


Then, analyze the profile file using:

```bash
go tool pprof heap.prof
```

**3. Using `-memprofile` and `-cpuprofile` with `go test`**

You can also collect profiles during test runs using the `-memprofile` and `-cpuprofile` flags:

```bash
# Memory profile
go test -memprofile mem.prof ./your_package

# CPU profile
go test -cpuprofile cpu.prof ./your_package
```

Analyze the profiles using:

```bash
go tool pprof mem.prof
go tool pprof cpu.prof
```

**4. When You Can't Modify the Code (Post-Mortem Debugging)**

If you encounter a memory leak in a production environment where you can't easily modify the code to enable `pprof`, it becomes more challenging:

*   **Containerized Environments (e.g., Kubernetes):**
    *   **Sidecar Container:** You might be able to add a sidecar container to your pod that runs an HTTP server and exposes the `pprof` endpoints. This sidecar would need to communicate with your main application container to trigger profile collection.
    *   **Ephemeral Containers:** If your Kubernetes version supports ephemeral containers (still an alpha/beta feature in some versions), you might be able to inject a temporary container into the running pod to collect a profile.
*   **Remote Debugging (Advanced):** Tools like Delve (dlv) can, in theory, attach to running processes, but this is a more complex and less reliable approach for memory profiling. It's generally used for interactive debugging of issues rather than performance profiling. It could be too late to attach a debugger when a leak happens in production.

**Important Considerations:**

*   **Production Overhead:** Profiling can introduce overhead, so be cautious when enabling it in production. For high-traffic services, you might need to sample profiles or use a less intrusive profiling technique. It's often better to reproduce the issue in a staging or testing environment for detailed profiling.
*   **Security:** If you expose `pprof` endpoints publicly, make sure to secure them properly (e.g., using authentication or network restrictions) to prevent unauthorized access.

**In Summary:**

The easiest way to use `pprof` is to enable it in your code and start an HTTP server. If that's not possible, you can use `runtime/pprof` directly or leverage testing flags. For post-mortem debugging in production without code changes, options are limited and more complex, often involving containerization features or remote debugging. The best approach depends on your specific environment and constraints.

## what if pprof have not been enabled?

If you haven't explicitly enabled `pprof` in your Go application, you won't be able to collect profiling data using the standard `net/http/pprof` endpoints. Here's what you need to do in that situation and alternative approaches:

**1. Enabling `pprof`**

*   **Import `net/http/pprof`:**  The simplest way to enable `pprof` is to add the following import to your application's main package (or any package that is imported by your main package):

    ```go
    import _ "net/http/pprof"
    ```

    This blank import registers the `pprof` handlers with the default HTTP server (`http.DefaultServeMux`).

*   **Start an HTTP Server (if necessary):** If your application doesn't already have an HTTP server running, you'll need to start one. You can use a dedicated server on a separate port (often port 6060 by convention) for profiling:

    ```go
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    ```
    Or you can use your existing server to do that

*   **Access `pprof` Endpoints:** Once the server is running, you can access the `pprof` endpoints in your browser or using the `go tool pprof` command:
    *   Browser: `http://localhost:6060/debug/pprof/`
    *   Command Line: `go tool pprof http://localhost:6060/debug/pprof/heap` (or other profile types like `profile` for CPU profiling).

**2. Alternative: Using `runtime/pprof` Directly**

If you can't or don't want to start an HTTP server, you can use the `runtime/pprof` package to collect profiles programmatically and write them to files:

```go
import (
	"log"
	"os"
	"runtime/pprof"
)

func main() {
	// ... your application code ...

	// Create a file to write the heap profile to
	f, err := os.Create("heap.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close()

	// Collect a heap profile
	// runtime.GC() // Optionally trigger garbage collection before profiling
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

	// ... your application code ...

    // For CPU profiling
    // f, err := os.Create("cpu.prof")
    // if err != nil {
    //     log.Fatal("could not create CPU profile: ", err)
    // }
    // defer f.Close()
    // if err := pprof.StartCPUProfile(f); err != nil {
    //     log.Fatal("could not start CPU profile: ", err)
    // }
    // defer pprof.StopCPUProfile()
}
```

Then, analyze the profile file using:

```bash
go tool pprof heap.prof
```

**3. Using `-memprofile` and `-cpuprofile` with `go test`**

You can also collect profiles during test runs using the `-memprofile` and `-cpuprofile` flags:

```bash
# Memory profile
go test -memprofile mem.prof ./your_package

# CPU profile
go test -cpuprofile cpu.prof ./your_package
```

Analyze the profiles using:

```bash
go tool pprof mem.prof
go tool pprof cpu.prof
```

**4. When You Can't Modify the Code (Post-Mortem Debugging)**

If you encounter a memory leak in a production environment where you can't easily modify the code to enable `pprof`, it becomes more challenging:

*   **Containerized Environments (e.g., Kubernetes):**
    *   **Sidecar Container:** You might be able to add a sidecar container to your pod that runs an HTTP server and exposes the `pprof` endpoints. This sidecar would need to communicate with your main application container to trigger profile collection.
    *   **Ephemeral Containers:** If your Kubernetes version supports ephemeral containers (still an alpha/beta feature in some versions), you might be able to inject a temporary container into the running pod to collect a profile.
*   **Remote Debugging (Advanced):** Tools like Delve (dlv) can, in theory, attach to running processes, but this is a more complex and less reliable approach for memory profiling. It's generally used for interactive debugging of issues rather than performance profiling. It could be too late to attach a debugger when a leak happens in production.

**Important Considerations:**

*   **Production Overhead:** Profiling can introduce overhead, so be cautious when enabling it in production. For high-traffic services, you might need to sample profiles or use a less intrusive profiling technique. It's often better to reproduce the issue in a staging or testing environment for detailed profiling.
*   **Security:** If you expose `pprof` endpoints publicly, make sure to secure them properly (e.g., using authentication or network restrictions) to prevent unauthorized access.

**In Summary:**

The easiest way to use `pprof` is to enable it in your code and start an HTTP server. If that's not possible, you can use `runtime/pprof` directly or leverage testing flags. For post-mortem debugging in production without code changes, options are limited and more complex, often involving containerization features or remote debugging. The best approach depends on your specific environment and constraints.
