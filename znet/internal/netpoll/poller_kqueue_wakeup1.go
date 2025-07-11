//go:build netbsd || openbsd

package netpoll

import (
	"github.com/meta-apex/zenith/zlog"
	"golang.org/x/sys/unix"
)

// TODO(panjf2000): NetBSD didn't implement EVFILT_USER for user-established events
// until NetBSD 10.0, check out https://www.netbsd.org/releases/formal-10/NetBSD-10.0.html
// Therefore we use the pipe to wake up the kevent on NetBSD at this point. Get back here
// and switch to EVFILT_USER when we bump up the minimal requirement of NetBSD to 10.0.
// Alternatively, maybe we can use EVFILT_USER on the NetBSD by checking the kernel version
// via uname(3) and fall back to the pipe if the kernel version is older than 10.0.

func (p *Poller) addWakeupEvent() error {
	p.pipe = make([]int, 2)
	if err := unix.Pipe2(p.pipe[:], unix.O_NONBLOCK|unix.O_CLOEXEC); err != nil {
		zlog.Fatal().Msgf("failed to create pipe for wakeup event: %v", err)
	}
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{{
		Ident:  uint64(p.pipe[0]),
		Filter: unix.EVFILT_READ,
		Flags:  unix.EV_ADD,
	}}, nil, nil)
	return err
}

func (p *Poller) wakePoller() error {
retry:
	_, err := unix.Write(p.pipe[1], []byte("x"))
	if err == nil || err == unix.EAGAIN {
		return nil
	}
	if err == unix.EINTR {
		goto retry
	}
	zlog.Warn().Msgf("failed to write to the wakeup pipe: %v", err)
	return err
}

func (p *Poller) drainWakeupEvent() {
	var buf [8]byte
	_, _ = unix.Read(p.pipe[0], buf[:])
}
