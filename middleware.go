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
	GetOrCreateFunc         func(r *http.Request) *Limiter
}

func newOptions(limit uint64, period time.Duration, opts []MiddlewareOption) middlewareOptions {
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
	if o.GetOrCreateFunc == nil {
		commonLimiter := New(limit, period)
		o.GetOrCreateFunc = func(_ *http.Request) *Limiter {
			return commonLimiter
		}
	}

	return o
}

func Middleware(limit uint64, period time.Duration, opts ...MiddlewareOption) func(http.Handler) http.Handler {
	options := newOptions(limit, period, opts)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiter := options.GetOrCreateFunc(r)

			if !limiter.Available() {
				http.Error(w, ErrLimitExceeded.Error(), options.LimitExceededErrorCode)

				return
			}

			limiter.Inc()
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

func WithCustomRequestsRemainingHeader(header string) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.RequestsRemainingHeader = header
	}
}

func WithCustomLimitExceededErrorCode(code int) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.LimitExceededErrorCode = code
	}
}

func WithGetOrCreateFunc(f func(r *http.Request) *Limiter) MiddlewareOption {
	return func(o *middlewareOptions) {
		o.GetOrCreateFunc = f
	}
}
