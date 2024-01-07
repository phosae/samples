# Go Profile 

## generate pprof|trace data

1. Manually generate trace data in Go application via runtime/trace or runtime/pprof

```
go run main.go
```

2. Profiling web services via net/http/pprof

Run the service in a container with cpu and memory limit:

```
docker run --cpus=2 -m 200m  -p 8080:8080 $(ko build -L main.go)
```

use [hey](https://github.com/rakyll/hey) to load test the service

```
hey -z 60s http://localhost:8080
```

Fetch pprof/trace data using HTTP:

```
wget -O trace.out http://localhost:8080/debug/pprof/trace?seconds=15
```

3. Trace or Profile data via go test

```
go test -trace trace.out -run ^TestDoParallelTask$ example.zeng.dev
```

```
go test  -cpuprofile cpu.prof -memprofile mem.prof  -bench Calculate
```

## data analysis

analyze trace

```
go tool trace -http :6060 trace.out
```

analyze cpu profile

```
go tool pprof -http :6060 http://localhost:8080/debug/pprof/profile
```

```
go tool pprof -http :6060 cpu.prof
```

see full documentation at
- [net/http/pprof]
- [runtime/pprof]

## References
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Goroutine and Preemption](https://medium.com/a-journey-with-go/go-goroutine-and-preemption-d6bc2aa2f4b7)
- [Understanding Go Execution Tracer by Example](https://www.sobyte.net/post/2022-03/go-execution-tracer-by-example/)
- [pkg/profile](https://github.com/pkg/profile)

[runtime/pprof]: https://pkg.go.dev/runtime/pprof
[net/http/pprof]: https://pkg.go.dev/net/http/pprof