package hardlimit_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/tomr-ninja/hardlimit"
)

func TestMiddleware(t *testing.T) {
	createHandler := func(limit uint64, period time.Duration, mwOptions ...hardlimit.MiddlewareOption) http.Handler {
		middleware := hardlimit.SimpleMiddleware(limit, period, mwOptions...)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		return middleware(handler)
	}

	t.Run("default options", func(t *testing.T) {
		limit := uint64(10)
		period := 10 * time.Millisecond
		server := createHandler(limit, period)

		for i := 0; i < int(limit); i++ {
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusOK {
				t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
			}

			expectedRequestsRemaining := strconv.Itoa(int(limit) - i - 1)
			actualRequestsRemaining := rec.Header().Get(hardlimit.DefaultRequestsRemainingHeader)
			if actualRequestsRemaining != expectedRequestsRemaining {
				t.Errorf("expected %s, got %s", expectedRequestsRemaining, actualRequestsRemaining)
			}
		}

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusTooManyRequests {
			t.Errorf("expected %d, got %d", http.StatusTooManyRequests, rec.Code)
		}
	})

	t.Run("WithRequestsRemainingHeader option", func(t *testing.T) {
		limit := uint64(10)
		period := 10 * time.Millisecond
		customHeader := "X-Remaining"
		server := createHandler(limit, period, hardlimit.WithRequestsRemainingHeader(customHeader))

		rec := httptest.NewRecorder()

		server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
		}

		expectedRequestsRemaining := strconv.Itoa(int(limit) - 1)
		actualRequestsRemaining := rec.Header().Get(customHeader)
		if actualRequestsRemaining != expectedRequestsRemaining {
			t.Errorf("expected %s, got %s", expectedRequestsRemaining, actualRequestsRemaining)
		}
	})

	t.Run("WithLimitExceededErrorCode option", func(t *testing.T) {
		limit := uint64(10)
		period := 10 * time.Millisecond
		customErrorCode := http.StatusForbidden
		server := createHandler(limit, period, hardlimit.WithLimitExceededErrorCode(customErrorCode))

		for i := 0; i < int(limit); i++ {
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusOK {
				t.Errorf("expected %d, got %d", http.StatusOK, rec.Code)
			}
		}

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != customErrorCode {
			t.Errorf("expected %d, got %d", customErrorCode, rec.Code)
		}
	})
}
