Simple rate limiter.

## Usage

### Basic usage

```go
limiter := hardlimit.New(100, time.Minute)

if !limiter.Available() {
    limiter.Wait() // or just return error, why not
}
job.Do()
limiter.Inc()
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

fucn main() {
    mux := http.NewServeMux()
    handler := http.HandlerFunc(myHandler)
    mux.Handle("/", hardlimit.Middleware(100, time.Minute)(handler))	
}
```
