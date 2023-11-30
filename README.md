Simple rate limiter.

## Usage

> [!TIP]
> If you came here for an HTTP middleware, just scroll down to [this](#http-middleware).

> [!TIP]
> Rate limiter != [circuit breaker](https://en.wikipedia.org/wiki/Circuit_breaker_design_pattern). If you need a circuit breaker, there are other libriaries, e.g. [gobreaker](https://github.com/sony/gobreaker).

### Basic usage

```go
limiter := hardlimit.New(100, time.Minute)

...

if !limiter.Available() {
    return hardlimit.ErrLimitExceeded
}
job.Do()
limiter.Inc()
```

You can also wait for the next available slot:

```go
for !limiter.Available() { // or 'if'; keep in mind that there are no locks
    limiter.Wait()
}
job.Do()
```

`Wait()` is relatively bad for performance, so don't use it if you are not sure you need it.

### Using `Exec()` wrapper

```go
limiter := hardlimit.New(100, time.Minute)

err := limiter.Exec(job.Do)
if errors.Is(err, hardlimit.ErrLimitExceeded) {
    // handle limiter error
} else if err != nil {
    // handle job error
}
```

### HTTP middleware

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
}

func main() {
    mux := http.NewServeMux()
    rateLimiter := hardlimit.SimpleMiddleware(100, time.Minute)
    handler := http.HandlerFunc(myHandler)
    mux.Handle("/", rateLimiter(handler))
    http.ListenAndServe(":3000", mux)
}
```

### Advanced HTTP middleware

```go
// it's just an oversimplified example, don't do this in production
// use a sync.Map or something
var limiters = map[string]*hardlimit.Limiter{
    "42.42.42.42": hardlimit.New(100, time.Minute),
    "42.42.42.43": hardlimit.New(100, time.Minute),
}

func getOrCreateLimiter(r *http.Request) *hardlimit.Limiter {
    ip := r.Header.Get("X-Real-Ip") // or r.RemoteAddr, or some token, whatever you want, you have the request
    limiter, ok := limiters[ip]
    if !ok { ... } // create new limiter and store it

    return limiter	
}

func main() {
    mux := http.NewServeMux()
    handler := http.HandlerFunc(myHandler)
    middleware := hardlimit.Middleware(getOrCreateLimiter)
    mux.Handle("/", middleware(handler))
    http.ListenAndServe(":3000", mux)
}
```

## Benchmarks

```
goos: linux
goarch: amd64
pkg: github.com/tomr-ninja/hardlimit
cpu: AMD Ryzen 5 5600H with Radeon Graphics         
BenchmarkLimiter
BenchmarkLimiter/no_limiter
BenchmarkLimiter/no_limiter-12    	291471141	         4.099 ns/op
BenchmarkLimiter/with_limiter
BenchmarkLimiter/with_limiter-12  	208200771	         5.811 ns/op
BenchmarkLimiterParallel
BenchmarkLimiterParallel/no_limiter
BenchmarkLimiterParallel/no_limiter-12         	44305459	        27.16 ns/op
BenchmarkLimiterParallel/with_limiter
BenchmarkLimiterParallel/with_limiter-12       	31016211	        39.02 ns/op
BenchmarkMiddlewareParallel
BenchmarkMiddlewareParallel/no_middleware
BenchmarkMiddlewareParallel/no_middleware-12   	 1202026	      1006 ns/op
BenchmarkMiddlewareParallel/with_middleware
BenchmarkMiddlewareParallel/with_middleware-12 	 1000000	      1064 ns/op
```
