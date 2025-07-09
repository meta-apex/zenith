//go:build dragonfly || freebsd || netbsd || openbsd

package socket

import (
	"github.com/meta-apex/zenith/core/zerror"
)

// SetBindToDevice is not implemented on *BSD because there is
// no equivalent of Linux's SO_BINDTODEVICE.
func SetBindToDevice(_ int, _ string) error {
	return zerror.ErrUnsupportedOp
}
