//go:build linux

package netpoll

import "golang.org/x/sys/unix"

// IOFlags represents the flags of IO events.
type IOFlags = uint16

// IOEvent is the integer type of I/O events on Linux.
type IOEvent = uint32

const (
	// InitPollEventsCap represents the initial capacity of poller event-list.
	InitPollEventsCap = 128
	// MaxPollEventsCap is the maximum limitation of events that the poller can process.
	MaxPollEventsCap = 1024
	// MinPollEventsCap is the minimum limitation of events that the poller can process.
	MinPollEventsCap = 32
	// MaxAsyncTasksAtOneTime is the maximum amount of asynchronous tasks that the event-loop will process at one time.
	MaxAsyncTasksAtOneTime = 256
	// ReadEvents represents readable events that are polled by epoll.
	ReadEvents = unix.EPOLLIN | unix.EPOLLPRI
	// WriteEvents represents writeable events that are polled by epoll.
	WriteEvents = unix.EPOLLOUT
	// ReadWriteEvents represents both readable and writeable events.
	ReadWriteEvents = ReadEvents | WriteEvents
	// ErrEvents represents exceptional events that occurred.
	ErrEvents = unix.EPOLLERR | unix.EPOLLHUP
)

// IsReadEvent checks if the event is a read event.
func IsReadEvent(event IOEvent) bool {
	return event&ReadEvents != 0
}

// IsWriteEvent checks if the event is a write event.
func IsWriteEvent(event IOEvent) bool {
	return event&WriteEvents != 0
}

// IsErrorEvent checks if the event is an error event.
func IsErrorEvent(event IOEvent, _ IOFlags) bool {
	return event&ErrEvents != 0
}

type eventList struct {
	size   int
	events []epollevent
}

func newEventList(size int) *eventList {
	return &eventList{size, make([]epollevent, size)}
}

func (el *eventList) expand() {
	if newSize := el.size << 1; newSize <= MaxPollEventsCap {
		el.size = newSize
		el.events = make([]epollevent, newSize)
	}
}

func (el *eventList) shrink() {
	if newSize := el.size >> 1; newSize >= MinPollEventsCap {
		el.size = newSize
		el.events = make([]epollevent, newSize)
	}
}
