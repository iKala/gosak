package timeutil

// inspired by https://github.com/jonboulle/clockwork
// but clockwork does not provide method to set now

import (
	"sync"
	"time"
)

// NewRealClock creates real clock
func NewRealClock() Clock {
	return &realClock{}
}

// NewFakeClock for testing
func NewFakeClock() FakeClock {
	return &fakeClock{
		now: time.Now(),
	}
}

// Clock defines an interface for clock
// The purpose of this interface is to create an abstract
// layer to easy testing with time
type Clock interface {
	// Now returns the current time
	Now() time.Time
}

// FakeClock is for testing
type FakeClock interface {
	Clock
	// SetNow set current time
	SetNow(time.Time)
	// Incr increase curren time
	Incr(time.Duration) time.Time
}

type realClock struct{}

func (*realClock) Now() time.Time {
	return time.Now()
}

type fakeClock struct {
	now  time.Time
	lock sync.RWMutex
}

func (f *fakeClock) Now() time.Time {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.now
}

func (f *fakeClock) SetNow(now time.Time) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.now = now
}

func (f *fakeClock) Incr(d time.Duration) time.Time {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.now = f.now.Add(d)
	return f.now
}
