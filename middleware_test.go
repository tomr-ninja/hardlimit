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
	t.Run("default options", func(t *testing.T) {
		limit := uint64(10)
		period := time.Millisecond

		server := hardlimit.Middleware(limit, period)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

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
}
