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
}

func (o *middlewareOptions) apply(opts []MiddlewareOption) {
	for _, opt := range opts {
		opt(o)
	}

	if o.LimitExceededErrorCode == 0 {
		o.LimitExceededErrorCode = http.StatusTooManyRequests
	}
	if o.RequestsRemainingHeader == "" {
		o.RequestsRemainingHeader = defaultRequestsRemainingHeader
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
)

func Middleware(limit uint64, period time.Duration, opts ...MiddlewareOption) func(http.Handler) http.Handler {
	limiter := New(limit, period)

	options := new(middlewareOptions)
	options.apply(opts)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Available() {
				http.Error(w, ErrLimitExceeded.Error(), options.LimitExceededErrorCode)

				return
			}

			next.ServeHTTP(w, r)

			if options.RequestsRemainingHeader != "" {
				w.Header().Set(options.RequestsRemainingHeader, strconv.FormatUint(limit-limiter.Count(), 10))
			}
		})
	}
}
