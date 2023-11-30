package hardlimit_test

import (
	"math"
	"testing"
	"time"

	"github.com/tomr-ninja/hardlimit"
)

func BenchmarkLimiter(b *testing.B) {
	limiter := createLimitlessLimiter()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = limiter.Exec(func() error {
			return nil
		})
	}
}

func BenchmarkLimiterParallel(b *testing.B) {
	limiter := createLimitlessLimiter()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = limiter.Exec(func() error {
				return nil
			})
		}
	})
}

func createLimitlessLimiter() *hardlimit.Limiter {
	return hardlimit.New(math.MaxUint64, time.Hour) // should never be exceeded during the benchmark
}
