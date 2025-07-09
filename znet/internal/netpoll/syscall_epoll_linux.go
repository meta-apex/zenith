//go:build poll_opt

package netpoll

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

func epollCtl(epfd int, op int, fd int, event *epollevent) error {
	_, _, errno := unix.RawSyscall6(unix.SYS_EPOLL_CTL, uintptr(epfd), uintptr(op), uintptr(fd), uintptr(unsafe.Pointer(event)), 0, 0)
	if errno != 0 {
		return errnoErr(errno)
	}
	return nil
}
