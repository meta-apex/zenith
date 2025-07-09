//go:build darwin || dragonfly || freebsd || openbsd

package netpoll

// IOFlags represents the flags of IO events.
type IOFlags = uint16

// IOEvent is the integer type of I/O events on BSD's.
type IOEvent = int16
