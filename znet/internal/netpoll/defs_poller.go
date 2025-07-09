//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package netpoll

// PollEventHandler is the callback for I/O events notified by the poller.
type PollEventHandler func(int, IOEvent, IOFlags) error

// PollAttachment is the user data which is about to be stored in "void *ptr" of epoll_data or "void *udata" of kevent.
type PollAttachment struct {
	FD       int
	Callback PollEventHandler
}
