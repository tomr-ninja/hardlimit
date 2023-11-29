package hardlimit

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type (
	Limiter struct {
		limit   uint64
		counter atomic.Uint64
		subs    subscriptions
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
		subs: subscriptions{
			chans: make([]chan struct{}, 0, 64),
			mux:   sync.RWMutex{},
		},
	}

	go func() {
		for range time.Tick(period) {
			limiter.Reset()
			limiter.subs.notify()
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

func (l *Limiter) Wait() {
	if l.Available() {
		return
	}

	<-l.subs.add()
}

type subscriptions struct {
	chans []chan struct{}
	mux   sync.RWMutex
}

func (s *subscriptions) add() <-chan struct{} {
	s.mux.Lock()
	defer s.mux.Unlock()

	ch := make(chan struct{})
	s.chans = append(s.chans, ch)

	return ch
}

func (s *subscriptions) notify() {
	s.mux.RLock()
	n := len(s.chans)
	s.mux.RUnlock()

	if n == 0 {
		return
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	for _, ch := range s.chans {
		close(ch)
	}

	s.chans = s.chans[:0]
}
