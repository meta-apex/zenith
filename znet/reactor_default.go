//go:build (darwin || dragonfly || freebsd || linux || netbsd || openbsd) && !poll_opt

package znet

import (
	"errors"
	"github.com/meta-apex/zenith/core/zerror"
	"github.com/meta-apex/zenith/znet/internal/netpoll"
	"runtime"
)

func (el *eventloop) rotate() error {
	if el.engine.opts.LockOSThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	err := el.poller.Polling(el.accept0)
	if errors.Is(err, zerror.ErrEngineShutdown) {
		el.getLogger().Debug().Msgf("main reactor is exiting in terms of the demand from user, %v", err)
		err = nil
	} else if err != nil {
		el.getLogger().Error().Msgf("main reactor is exiting due to error: %v", err)
	}

	el.engine.shutdown(err)

	return err
}

func (el *eventloop) orbit() error {
	if el.engine.opts.LockOSThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	err := el.poller.Polling(func(fd int, ev netpoll.IOEvent, flags netpoll.IOFlags) error {
		c := el.connections.getConn(fd)
		if c == nil {
			// For kqueue, this might happen when the connection has already been closed,
			// the file descriptor will be deleted from kqueue automatically as documented
			// in the manual pages.
			// For epoll, it somehow notified with an event for a stale fd that is not in
			// our connection set. We need to explicitly delete it from the epoll set.
			// Also print a warning log for this kind of irregularity.
			el.getLogger().Warn().Msgf("received event[fd=%d|ev=%d|flags=%d] of a stale connection from event-loop(%d)", fd, ev, flags, el.idx)
			return el.poller.Delete(fd)
		}
		return c.processIO(fd, ev, flags)
	})
	if errors.Is(err, zerror.ErrEngineShutdown) {
		el.getLogger().Debug().Msgf("event-loop(%d) is exiting in terms of the demand from user, %v", el.idx, err)
		err = nil
	} else if err != nil {
		el.getLogger().Error().Msgf("event-loop(%d) is exiting due to error: %v", el.idx, err)
	}

	el.closeConns()
	el.engine.shutdown(err)

	return err
}

func (el *eventloop) run() error {
	if el.engine.opts.LockOSThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	err := el.poller.Polling(func(fd int, ev netpoll.IOEvent, flags netpoll.IOFlags) error {
		c := el.connections.getConn(fd)
		if c == nil {
			if _, ok := el.listeners[fd]; ok {
				return el.accept(fd, ev, flags)
			}
			// For kqueue, this might happen when the connection has already been closed,
			// the file descriptor will be deleted from kqueue automatically as documented
			// in the manual pages.
			// For epoll, it somehow notified with an event for a stale fd that is not in
			// our connection set. We need to explicitly delete it from the epoll set.
			// Also print a warning log for this kind of irregularity.
			el.getLogger().Warn().Msgf("received event[fd=%d|ev=%d|flags=%d] of a stale connection from event-loop(%d)", fd, ev, flags, el.idx)
			return el.poller.Delete(fd)
		}
		return c.processIO(fd, ev, flags)
	})
	if errors.Is(err, zerror.ErrEngineShutdown) {
		el.getLogger().Debug().Msgf("event-loop(%d) is exiting in terms of the demand from user, %v", el.idx, err)
		err = nil
	} else if err != nil {
		el.getLogger().Error().Msgf("event-loop(%d) is exiting due to error: %v", el.idx, err)
	}

	el.closeConns()
	el.engine.shutdown(err)

	return err
}
