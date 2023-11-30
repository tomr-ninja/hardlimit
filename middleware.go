package hardlimit

import (
	"net/http"
	"strconv"
	"time"
)

const DefaultRequestsRemainingHeader = "X-Requests-Remaining"

type MiddlewareOption func(*middlewareOptions)

type middlewareOptions struct {
	RequestsRemainingHeader string
	LimitExceededErrorCode  int
}

func newOptions(opts []MiddlewareOption) middlewareOptions {
	o := middlewareOptions{}
	for _, opt := range opts {
		opt(&o)
	}

	if o.LimitExceededErrorCode == 0 {
		o.LimitExceededErrorCode = http.StatusTooManyRequests
	}
	if o.RequestsRemainingHeader == "" {
		o.RequestsRemainingHeader = DefaultRequestsRemainingHeader
	}

	return o
}

// Middleware - construct a stdlib-compatible middleware with a rate limiter.
func Middleware(limiterForRequest func(r *http.Request) *Limiter, opts ...MiddlewareOption) func(http.Handler) http.Handler {
	options := newOptions(opts)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiter := limiterForRequest(r)

			if !limiter.Available() {
				http.Error(w, ErrLimitExceeded.Error(), options.LimitExceededErrorCode)

				return
			}

			limiter.Inc()
			next.ServeHTTP(w, r)

			if options.RequestsRemainingHeader != "" {
				limit := limiter.Limit()
				requestsRemaining := uint64(0)
				if c := limiter.Count(); c <= limit { // c > limit is possible in case of race condition
					requestsRemaining = limit - c
				}

				w.Header().Set(options.RequestsRemainingHeader, strconv.FormatUint(requestsRemaining, 10))
			}
		})
	}
}

// SimpleMiddleware - simplified version of Middleware with static (not request-wise) limiter.
func SimpleMiddleware(limit uint64, period time.Duration, opts ...MiddlewareOption) func(http.Handler) http.Handler {
	return Middleware(StaticLimiter(New(limit, period)), opts...)
}

// StaticLimiter is a simplest limiter getter to configure a middleware. It returns the same limiter for every request.
func StaticLimiter(l *Limiter) func(r *http.Request) *Limiter {
	return func(_ *http.Request) *Limiter {
		return l
	}
}

// WithRequestsRemainingHeader - middleware option to set a custom header name for requests remaining.
func WithRequestsRemainingHeader(header string) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.RequestsRemainingHeader = header
	}
}

// WithLimitExceededErrorCode - middleware option to set a custom error code for limit exceeded error.
func WithLimitExceededErrorCode(code int) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.LimitExceededErrorCode = code
	}
}
