package socket

import (
	"github.com/meta-apex/zenith/core/zerror"
)

// SetKeepAlivePeriod is not implemented on OpenBSD because there are
// no equivalents of Linux's TCP_KEEPIDLE, TCP_KEEPINTVL, and TCP_KEEPCNT.
func SetKeepAlivePeriod(_, _ int) error {
	// OpenBSD has no user-settable per-socket TCP keepalive options.
	return zerror.ErrUnsupportedOp
}

// SetKeepAlive is not implemented on OpenBSD.
func SetKeepAlive(_ int, _ bool, _, _, _ int) error {
	// OpenBSD has no user-settable per-socket TCP keepalive options.
	return zerror.ErrUnsupportedOp
}
