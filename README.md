Simple rate limiter.

## Usage

> [!TIP]
> If you came here for an HTTP middleware, just scroll down to [this](#http-middleware).

> [|TIP]
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
    rateLimiter := hardlimit.Middleware(100, time.Minute)
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
    middleware := hardlimit.Middleware(
        100, time.Minute,
        WithGetOrCreateFunc(getOrCreateLimiter), // see middleware.go for more options
    )
    mux.Handle("/", middleware(handler))
    http.ListenAndServe(":3000", mux)
}
```
