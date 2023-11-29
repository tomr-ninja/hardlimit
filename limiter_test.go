package main

import (
	"errors"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	limit := uint64(10)
	period := time.Millisecond
	limiter := New(limit, period)

	for i := 0; i < 9; i++ {
		limiter.Inc()

		if !limiter.Available() {
			t.Errorf("expected to be available")
		}
	}

	limiter.Inc()
	if limiter.Available() {
		t.Errorf("expected to be unavailable")
	}

	time.Sleep(period)

	if !limiter.Available() {
		t.Errorf("expected to be available")
	}
}

func TestLimiter_Exec(t *testing.T) {
	limit := uint64(10)
	period := time.Millisecond
	limiter := New(limit, period)

	for i := 0; i < 10; i++ {
		c, err := limiter.Exec(func() error {
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if c != uint64(i+1) {
			t.Errorf("expected %d, got %d", i+1, c)
		}
	}

	c, err := limiter.Exec(func() error {
		return nil
	})
	if !errors.Is(err, ErrLimitExceeded) {
		t.Errorf("expected %v, got %v", ErrLimitExceeded, err)
	}
	if c != limit {
		t.Errorf("expected %d, got %d", limit, c)
	}
	if limiter.Available() {
		t.Errorf("expected to be unavailable")
	}
}

func TestInvalidInit(t *testing.T) {
	t.Run("negative period", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected to panic")
			}
		}()

		New(10, -time.Millisecond)
	})

	t.Run("zero limit", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected to panic")
			}
		}()

		New(0, time.Millisecond)
	})
}
