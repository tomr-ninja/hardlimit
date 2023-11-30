package hardlimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tomr-ninja/hardlimit"
)

func BenchmarkMiddleware(b *testing.B) {
	createHandler := func(withMiddleware bool) http.Handler {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		if !withMiddleware {
			return handler
		}

		middleware := hardlimit.Middleware(hardlimit.StaticLimiter(createLimitlessLimiter()))

		return middleware(handler)
	}

	benchServer := func(b *testing.B, server http.Handler) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			rec := httptest.NewRecorder()

			server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		}
	}

	b.Run("no middleware", func(b *testing.B) {
		benchServer(b, createHandler(false))
	})

	b.Run("with middleware", func(b *testing.B) {
		benchServer(b, createHandler(true))
	})
}
