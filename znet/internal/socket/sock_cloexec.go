//go:build dragonfly || freebsd || linux

package socket

import "golang.org/x/sys/unix"

func sysSocket(family, sotype, proto int) (int, error) {
	return unix.Socket(family, sotype|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, proto)
}

func sysAccept(fd int) (nfd int, sa unix.Sockaddr, err error) {
	return unix.Accept4(fd, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
}
