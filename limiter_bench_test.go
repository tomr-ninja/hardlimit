package hardlimit_test

import (
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tomr-ninja/hardlimit"
)

func BenchmarkLimiter(b *testing.B) {
	b.Run("no limiter", func(b *testing.B) {
		job := createJob()

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = job()
		}
	})

	b.Run("with limiter", func(b *testing.B) {
		limiter := createLimitlessLimiter()
		job := createJob()

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = limiter.Exec(job)
		}
	})
}

func BenchmarkLimiterParallel(b *testing.B) {
	b.Run("no limiter", func(b *testing.B) {
		job := createJob()

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = job()
			}
		})
	})

	b.Run("with limiter", func(b *testing.B) {
		limiter := createLimitlessLimiter()
		job := createJob()

		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = limiter.Exec(job)
			}
		})
	})
}

func createLimitlessLimiter() *hardlimit.Limiter {
	return hardlimit.New(math.MaxUint64, time.Hour) // should never be exceeded during the benchmark
}

func createJob() func() error {
	v := atomic.Uint64{}

	return func() error {
		v.CompareAndSwap(0, 1)
		v.CompareAndSwap(1, 0)

		return nil
	}
}
