package socket

import (
	"errors"
	"github.com/meta-apex/zenith/core/zerror"
	"os"

	"golang.org/x/sys/unix"
)

// SetKeepAlivePeriod enables the SO_KEEPALIVE option on the socket and sets
// TCP_KEEPIDLE/TCP_KEEPALIVE to the specified duration in seconds, TCP_KEEPCNT
// to 5, and TCP_KEEPINTVL to secs/5.
func SetKeepAlivePeriod(fd, secs int) error {
	if secs <= 0 {
		return errors.New("invalid time duration")
	}

	interval := secs / 5
	if interval == 0 {
		interval = 1
	}

	return SetKeepAlive(fd, true, secs, interval, 5)
}

// SetKeepAlive enables/disables the TCP keepalive feature on the socket.
func SetKeepAlive(fd int, enabled bool, idle, intvl, cnt int) error {
	if enabled && (idle <= 0 || intvl <= 0 || cnt <= 0) {
		return errors.New("invalid time duration")
	}

	var on int
	if enabled {
		on = 1
	}

	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, on); err != nil {
		return os.NewSyscallError("setsockopt", err)
	}

	if !enabled {
		// If keepalive is disabled, ignore the TCP_KEEP* options.
		return nil
	}

	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPALIVE, idle); err != nil {
		return os.NewSyscallError("setsockopt", err)
	}

	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, intvl); err != nil {
		return os.NewSyscallError("setsockopt", err)
	}

	return os.NewSyscallError("setsockopt", unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPCNT, cnt))
}

// SetBindToDevice is not implemented on macOS because there is
// no equivalent of Linux's SO_BINDTODEVICE.
func SetBindToDevice(_ int, _ string) error {
	return zerror.ErrUnsupportedOp
}
