//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package znet

import (
	"github.com/meta-apex/zenith/core/zerror"
	"github.com/meta-apex/zenith/znet/internal/netpoll"
	"github.com/meta-apex/zenith/znet/internal/queue"
	"github.com/meta-apex/zenith/znet/internal/socket"
	"runtime"

	"golang.org/x/sys/unix"
)

func (el *eventloop) accept0(fd int, _ netpoll.IOEvent, _ netpoll.IOFlags) error {
	for {
		nfd, sa, err := socket.Accept(fd)
		switch err {
		case nil:
		case unix.EAGAIN: // the Accept queue has been drained out, we can return now
			return nil
		case unix.EINTR, unix.ECONNRESET, unix.ECONNABORTED:
			// ECONNRESET or ECONNABORTED could indicate that a socket
			// in the Accept queue was closed before we Accept()ed it.
			// It's a silly error, let's retry it.
			continue
		default:
			el.getLogger().Error().Msgf("Accept() failed due to error: %v", err)
			return zerror.ErrAcceptSocket
		}

		remoteAddr := socket.SockaddrToTCPOrUnixAddr(sa)
		network := el.listeners[fd].network
		if opts := el.engine.opts; opts.TCPKeepAlive > 0 && network == "tcp" &&
			(runtime.GOOS != "linux" && runtime.GOOS != "freebsd" && runtime.GOOS != "dragonfly") {
			// TCP keepalive options are not inherited from the listening socket
			// on platforms other than Linux, FreeBSD, or DragonFlyBSD.
			// We therefore need to set them on the accepted socket explicitly.
			//
			// Check out https://github.com/nginx/nginx/pull/337 for details.
			if err = setKeepAlive(
				nfd,
				true,
				opts.TCPKeepAlive,
				opts.TCPKeepInterval,
				opts.TCPKeepCount); err != nil {
				el.getLogger().Error().Msgf("failed to set TCP keepalive on fd=%d: %v", fd, err)
			}
		}

		el := el.engine.eventLoops.next(remoteAddr)
		c := newStreamConn(network, nfd, el, sa, el.listeners[fd].addr, remoteAddr)
		err = el.poller.Trigger(queue.HighPriority, el.register, c)
		if err != nil {
			el.getLogger().Error().Msgf("failed to enqueue the accepted socket fd=%d to poller: %v", c.fd, err)
			_ = unix.Close(nfd)
			c.release()
		}
	}
}

func (el *eventloop) accept(fd int, ev netpoll.IOEvent, flags netpoll.IOFlags) error {
	network := el.listeners[fd].network
	if network == "udp" {
		return el.readUDP(fd, ev, flags)
	}

	nfd, sa, err := socket.Accept(fd)
	switch err {
	case nil:
	case unix.EINTR, unix.EAGAIN, unix.ECONNRESET, unix.ECONNABORTED:
		// ECONNRESET or ECONNABORTED could indicate that a socket
		// in the Accept queue was closed before we Accept()ed it.
		// It's a silly error, let's retry it.
		return nil
	default:
		el.getLogger().Error().Msgf("Accept() failed due to error: %v", err)
		return zerror.ErrAcceptSocket
	}

	remoteAddr := socket.SockaddrToTCPOrUnixAddr(sa)
	if opts := el.engine.opts; opts.TCPKeepAlive > 0 && el.listeners[fd].network == "tcp" &&
		(runtime.GOOS != "linux" && runtime.GOOS != "freebsd" && runtime.GOOS != "dragonfly") {
		// TCP keepalive options are not inherited from the listening socket
		// on platforms other than Linux, FreeBSD, or DragonFlyBSD.
		// We therefore need to set them on the accepted socket explicitly.
		//
		// Check out https://github.com/nginx/nginx/pull/337 for details.
		if err = setKeepAlive(
			nfd,
			true,
			opts.TCPKeepAlive,
			opts.TCPKeepInterval,
			opts.TCPKeepCount); err != nil {
			el.getLogger().Error().Msgf("failed to set TCP keepalive on fd=%d: %v", fd, err)
		}
	}

	c := newStreamConn(network, nfd, el, sa, el.listeners[fd].addr, remoteAddr)
	return el.register0(c)
}
