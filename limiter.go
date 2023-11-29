package main

import (
	"errors"
	"sync/atomic"
	"time"
)

type (
	Limiter struct {
		limit   uint64
		counter atomic.Uint64
	}
)

var (
	ErrLimitExceeded = errors.New("limit exceeded")
)

func New(limit uint64, period time.Duration) *Limiter {
	if period <= 0 {
		panic("invalid period: must be positive")
	}
	if limit == 0 {
		panic("invalid limit: must be positive")
	}

	limiter := &Limiter{
		limit:   limit,
		counter: atomic.Uint64{},
	}

	go func() {
		for range time.Tick(period) {
			limiter.Reset()
		}
	}()

	return limiter
}

func (l *Limiter) Exec(cb func() error) (uint64, error) {
	if !l.Available() {
		return l.Count(), ErrLimitExceeded
	}

	return l.Inc(), cb()
}

func (l *Limiter) Inc() uint64 {
	return l.counter.Add(1)
}

func (l *Limiter) Available() bool {
	return l.Count() < l.limit
}

func (l *Limiter) Reset() {
	l.counter.Store(0)
}

func (l *Limiter) Count() uint64 {
	return l.counter.Load()
}
