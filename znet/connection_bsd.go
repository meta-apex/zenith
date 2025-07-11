//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package znet

import (
	"github.com/meta-apex/zenith/znet/internal/netpoll"
	"io"

	"golang.org/x/sys/unix"
)

func (c *conn) processIO(_ int, filter netpoll.IOEvent, flags netpoll.IOFlags) (err error) {
	el := c.loop
	switch filter {
	case unix.EVFILT_READ:
		err = el.read(c)
	case unix.EVFILT_WRITE:
		err = el.write(c)
	}
	// EV_EOF indicates that the remote has closed the connection.
	// We check for EV_EOF after processing the read/write event
	// to ensure that nothing is left out on this event filter.
	if flags&unix.EV_EOF != 0 && c.opened && err == nil {
		switch filter {
		case unix.EVFILT_READ:
			// Received the event of EVFILT_READ|EV_EOF, but the previous eventloop.read
			// failed to drain the socket buffer, so we make sure we get it done this time.
			c.isEOF = true
			err = el.read(c)
		case unix.EVFILT_WRITE:
			// On macOS, the kqueue in either LT or ET mode will notify with one event for the
			// EOF of the TCP remote: EVFILT_READ|EV_ADD|EV_CLEAR|EV_EOF. But for some reason,
			// two events will be issued in ET mode for the EOF of the Unix remote in this order:
			// 1) EVFILT_WRITE|EV_ADD|EV_CLEAR|EV_EOF, 2) EVFILT_READ|EV_ADD|EV_CLEAR|EV_EOF.
			err = el.write(c)
		default:
			c.outboundBuffer.Release() // don't bother to write to a connection that is already broken
			err = el.close(c, io.EOF)
		}
	}
	return
}
