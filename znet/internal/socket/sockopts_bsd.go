//go:build dragonfly || freebsd || netbsd || openbsd

package socket

import (
	errorx "github.com/meta-apex/zenith/znet/internal/errors"
)

// SetBindToDevice is not implemented on *BSD because there is
// no equivalent of Linux's SO_BINDTODEVICE.
func SetBindToDevice(_ int, _ string) error {
	return errorx.ErrUnsupportedOp
}
