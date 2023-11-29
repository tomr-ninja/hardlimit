Simple rate limiter.

## Usage

> [!TIP]
> If you came here for an HTTP middleware, just scroll down to [this](#http-middleware).

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
for !limiter.Available() { // or if; may not worth it to simulate locks really
    limiter.Wait()
}
job.Do()
```

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
    handler := http.HandlerFunc(myHandler)
    mux.Handle("/", hardlimit.Middleware(100, time.Minute)(handler))	
}
```

### Advanced HTTP middleware

```go
// it's just a dumb example, don't do this in production
// use a sync.Map or something
limiters := map[string]*hardlimit.Limiter{
    "42.42.42.42": hardlimit.New(100, time.Minute),
    "42.42.42.43": hardlimit.New(100, time.Minute),
}

func main() {
    mux := http.NewServeMux()
    handler := http.HandlerFunc(myHandler)
    middleware := hardlimit.Middleware(
        100, time.Minute,
        WithGetOrCreateFunc(func(r *http.Request) *hardlimit.Limiter {
            ip := r.Header.Get("X-Real-Ip") // or r.RemoteAddr, or some token, whatever you want, you have the request
            limiter, ok := limiters[ip]
            if !ok {
                // create new limiter and store it
                ...
            }
            return limiter
        }),
    )
    mux.Handle("/", middleware(handler))
}
```
