package io

import "golang.org/x/sys/unix"

// Writev calls writev() on Linux.
func Writev(fd int, iov [][]byte) (int, error) {
	if len(iov) == 0 {
		return 0, nil
	}
	return unix.Writev(fd, iov)
}

// Readv calls readv() on Linux.
func Readv(fd int, iov [][]byte) (int, error) {
	if len(iov) == 0 {
		return 0, nil
	}
	return unix.Readv(fd, iov)
}
