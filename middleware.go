package hardlimit

import (
	"net/http"
	"strconv"
	"time"
)

const defaultRequestsRemainingHeader = "X-Requests-Remaining"

type MiddlewareOption func(*middlewareOptions)

type middlewareOptions struct {
	RequestsRemainingHeader string
	LimitExceededErrorCode  int
	GetOrCreateFunc         func(r *http.Request) *Limiter
}

func (o *middlewareOptions) apply(limit uint64, period time.Duration, opts []MiddlewareOption) {
	for _, opt := range opts {
		opt(o)
	}

	if o.LimitExceededErrorCode == 0 {
		o.LimitExceededErrorCode = http.StatusTooManyRequests
	}
	if o.RequestsRemainingHeader == "" {
		o.RequestsRemainingHeader = defaultRequestsRemainingHeader
	}
	if o.GetOrCreateFunc == nil {
		commonLimiter := New(limit, period)
		o.GetOrCreateFunc = func(_ *http.Request) *Limiter {
			return commonLimiter
		}
	}
}

var (
	WithCustomRequestsRemainingHeader = func(header string) MiddlewareOption {
		return func(o *middlewareOptions) {
			o.RequestsRemainingHeader = header
		}
	}
	WithCustomLimitExceededErrorCode = func(code int) MiddlewareOption {
		return func(o *middlewareOptions) {
			o.LimitExceededErrorCode = code
		}
	}
	WithGetOrCreateFunc = func(f func(r *http.Request) *Limiter) MiddlewareOption {
		return func(o *middlewareOptions) {
			o.GetOrCreateFunc = f
		}
	}
)

func Middleware(limit uint64, period time.Duration, opts ...MiddlewareOption) func(http.Handler) http.Handler {
	options := new(middlewareOptions)
	options.apply(limit, period, opts)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiter := options.GetOrCreateFunc(r)

			if !limiter.Available() {
				http.Error(w, ErrLimitExceeded.Error(), options.LimitExceededErrorCode)

				return
			}

			next.ServeHTTP(w, r)

			if options.RequestsRemainingHeader != "" {
				requestsRemaining := limit
				if c := limiter.Count(); c > limit { // possible in case of race condition
					requestsRemaining = 0
				} else {
					requestsRemaining = limit - c
				}

				w.Header().Set(options.RequestsRemainingHeader, strconv.FormatUint(requestsRemaining, 10))
			}
		})
	}
}
