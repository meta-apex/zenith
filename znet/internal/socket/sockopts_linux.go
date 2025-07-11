package socket

import (
	"os"

	"golang.org/x/sys/unix"
)

// SetBindToDevice binds the socket to a specific network interface.
//
// SO_BINDTODEVICE on Linux works in both directions: only process packets
// received from the particular interface along with sending them through
// that interface, instead of following the default route.
func SetBindToDevice(fd int, ifname string) error {
	return os.NewSyscallError("setsockopt", unix.BindToDevice(fd, ifname))
}
