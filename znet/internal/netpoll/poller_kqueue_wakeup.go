//go:build darwin || dragonfly || freebsd

package netpoll

import (
	"github.com/meta-apex/zenith/zlog"
	"golang.org/x/sys/unix"
)

func (p *Poller) addWakeupEvent() error {
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Flags:  unix.EV_ADD | unix.EV_CLEAR,
	}}, nil, nil)
	return err
}

func (p *Poller) wakePoller() error {
retry:
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{{
		Ident:  0,
		Filter: unix.EVFILT_USER,
		Fflags: unix.NOTE_TRIGGER,
	}}, nil, nil)
	if err == nil {
		return nil
	}
	if err == unix.EINTR {
		// All changes contained in the changelist should have been applied
		// before returning EINTR. But let's be skeptical and retry it anyway,
		// to make a 100% commitment.
		goto retry
	}
	zlog.Warn().Msgf("failed to wake up the poller: %v", err)
	return err
}

func (p *Poller) drainWakeupEvent() {}
