package zsync

import (
	"github.com/meta-apex/zenith/core/zcast"
	"github.com/meta-apex/zenith/core/ztime"
	"time"
)

// A Cond is used to wait for conditions.
type Cond struct {
	signal chan zcast.PlaceholderType
}

// NewCond returns a Cond.
func NewCond() *Cond {
	return &Cond{
		signal: make(chan zcast.PlaceholderType),
	}
}

// WaitWithTimeout wait for signal return remain wait time or timed out.
func (cond *Cond) WaitWithTimeout(timeout time.Duration) (time.Duration, bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	begin := ztime.Now()
	select {
	case <-cond.signal:
		elapsed := ztime.Since(begin)
		remainTimeout := timeout - elapsed
		return remainTimeout, true
	case <-timer.C:
		return 0, false
	}
}

// Wait waits for signals.
func (cond *Cond) Wait() {
	<-cond.signal
}

// Signal wakes one goroutine waiting on c, if there is any.
func (cond *Cond) Signal() {
	select {
	case cond.signal <- zcast.Placeholder:
	default:
	}
}
