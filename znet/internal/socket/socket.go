//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

// Package socket provides some handy socket-related functions.
package socket

import (
	"net"

	"golang.org/x/sys/unix"
)

// Option is used for setting an option on socket.
type Option[T int | string] struct {
	SetSockOpt func(int, T) error
	Opt        T
}

func execSockOpts[T int | string](fd int, opts []Option[T]) error {
	for _, opt := range opts {
		if err := opt.SetSockOpt(fd, opt.Opt); err != nil {
			return err
		}
	}
	return nil
}

// TCPSocket creates a TCP socket and returns a file descriptor that refers to it.
// The given socket options will be set on the returned file descriptor.
func TCPSocket(proto, addr string, passive bool, sockOptInts []Option[int], sockOptStrs []Option[string]) (int, net.Addr, error) {
	return tcpSocket(proto, addr, passive, sockOptInts, sockOptStrs)
}

// UDPSocket creates a UDP socket and returns a file descriptor that refers to it.
// The given socket options will be set on the returned file descriptor.
func UDPSocket(proto, addr string, connect bool, sockOptInts []Option[int], sockOptStrs []Option[string]) (int, net.Addr, error) {
	return udpSocket(proto, addr, connect, sockOptInts, sockOptStrs)
}

// UnixSocket creates a Unix socket and returns a file descriptor that refers to it.
// The given socket options will be set on the returned file descriptor.
func UnixSocket(proto, addr string, passive bool, sockOptInts []Option[int], sockOptStrs []Option[string]) (int, net.Addr, error) {
	return udsSocket(proto, addr, passive, sockOptInts, sockOptStrs)
}

// Accept accepts the next incoming socket along with setting
// O_NONBLOCK and O_CLOEXEC flags on it.
func Accept(fd int) (int, unix.Sockaddr, error) {
	return sysAccept(fd)
}
