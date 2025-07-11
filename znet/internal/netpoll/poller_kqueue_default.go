//go:build (darwin || dragonfly || freebsd || netbsd || openbsd) && !poll_opt

package netpoll

import (
	"errors"
	"github.com/meta-apex/zenith/core/zerror"
	"github.com/meta-apex/zenith/zlog"
	"github.com/meta-apex/zenith/znet/internal/queue"
	"os"
	"runtime"
	"sync/atomic"

	"golang.org/x/sys/unix"
)

// Poller represents a poller which is in charge of monitoring file-descriptors.
type Poller struct {
	fd                          int
	pipe                        []int
	wakeupCall                  int32
	asyncTaskQueue              queue.AsyncTaskQueue // queue with low priority
	urgentAsyncTaskQueue        queue.AsyncTaskQueue // queue with high priority
	highPriorityEventsThreshold int32                // threshold of high-priority events
}

// OpenPoller instantiates a poller.
func OpenPoller() (poller *Poller, err error) {
	poller = new(Poller)
	if poller.fd, err = unix.Kqueue(); err != nil {
		poller = nil
		err = os.NewSyscallError("kqueue", err)
		return
	}
	if err = poller.addWakeupEvent(); err != nil {
		_ = poller.Close()
		poller = nil
		err = os.NewSyscallError("kevent | pipe2", err)
		return
	}
	poller.asyncTaskQueue = queue.NewLockFreeQueue()
	poller.urgentAsyncTaskQueue = queue.NewLockFreeQueue()
	poller.highPriorityEventsThreshold = MaxPollEventsCap
	return
}

// Close closes the poller.
func (p *Poller) Close() error {
	if len(p.pipe) == 2 {
		_ = unix.Close(p.pipe[0])
		_ = unix.Close(p.pipe[1])
	}
	return os.NewSyscallError("close", unix.Close(p.fd))
}

// Trigger enqueues task and wakes up the poller to process pending tasks.
// By default, any incoming task will enqueued into urgentAsyncTaskQueue
// before the threshold of high-priority events is reached. When it happens,
// any asks other than high-priority tasks will be shunted to asyncTaskQueue.
//
// Note that asyncTaskQueue is a queue of low-priority whose size may grow large and tasks in it may backlog.
func (p *Poller) Trigger(priority queue.EventPriority, fn queue.Func, param any) (err error) {
	task := queue.GetTask()
	task.Exec, task.Param = fn, param
	if priority > queue.HighPriority && p.urgentAsyncTaskQueue.Length() >= p.highPriorityEventsThreshold {
		p.asyncTaskQueue.Enqueue(task)
	} else {
		// There might be some low-priority tasks overflowing into urgentAsyncTaskQueue in a flash,
		// but that's tolerable because it ought to be a rare case.
		p.urgentAsyncTaskQueue.Enqueue(task)
	}
	if atomic.CompareAndSwapInt32(&p.wakeupCall, 0, 1) {
		err = p.wakePoller()
	}
	return os.NewSyscallError("kevent | write", err)
}

// Polling blocks the current goroutine, monitoring the registered file descriptors and waiting for network I/O.
// When I/O occurs on any of the file descriptors, the provided callback function is invoked.
func (p *Poller) Polling(callback PollEventHandler) error {
	el := newEventList(InitPollEventsCap)

	var (
		ts       unix.Timespec
		tsp      *unix.Timespec
		doChores bool
	)
	for {
		n, err := unix.Kevent(p.fd, nil, el.events, tsp)
		if n == 0 || (n < 0 && err == unix.EINTR) {
			tsp = nil
			runtime.Gosched()
			continue
		} else if err != nil {
			zlog.Error().Msgf("error occurs in kqueue: %v", os.NewSyscallError("kevent wait", err))
			return err
		}
		tsp = &ts

		for i := 0; i < n; i++ {
			ev := &el.events[i]
			if fd := int(ev.Ident); fd == 0 { // poller is awakened to run tasks in queues
				doChores = true
				p.drainWakeupEvent()
			} else {
				err = callback(fd, ev.Filter, ev.Flags)
				if errors.Is(err, zerror.ErrAcceptSocket) || errors.Is(err, zerror.ErrEngineShutdown) {
					return err
				}
			}
		}

		if doChores {
			doChores = false
			task := p.urgentAsyncTaskQueue.Dequeue()
			for ; task != nil; task = p.urgentAsyncTaskQueue.Dequeue() {
				err = task.Exec(task.Param)
				if errors.Is(err, zerror.ErrEngineShutdown) {
					return err
				}
				queue.PutTask(task)
			}
			for i := 0; i < MaxAsyncTasksAtOneTime; i++ {
				if task = p.asyncTaskQueue.Dequeue(); task == nil {
					break
				}
				err = task.Exec(task.Param)
				if errors.Is(err, zerror.ErrEngineShutdown) {
					return err
				}
				queue.PutTask(task)
			}
			atomic.StoreInt32(&p.wakeupCall, 0)
			if (!p.asyncTaskQueue.IsEmpty() || !p.urgentAsyncTaskQueue.IsEmpty()) && atomic.CompareAndSwapInt32(&p.wakeupCall, 0, 1) {
				if err = p.wakePoller(); err != nil {
					doChores = true
				}
			}
		}

		if n == el.size {
			el.expand()
		} else if n < el.size>>1 {
			el.shrink()
		}
	}
}

// AddReadWrite registers the given file descriptor with readable and writable events to the poller.
func (p *Poller) AddReadWrite(pa *PollAttachment, edgeTriggered bool) error {
	var flags IOFlags = unix.EV_ADD
	if edgeTriggered {
		flags |= unix.EV_CLEAR
	}
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: keventIdent(pa.FD), Flags: flags, Filter: unix.EVFILT_READ},
		{Ident: keventIdent(pa.FD), Flags: flags, Filter: unix.EVFILT_WRITE},
	}, nil, nil)
	return os.NewSyscallError("kevent add", err)
}

// AddRead registers the given file descriptor with readable event to the poller.
func (p *Poller) AddRead(pa *PollAttachment, edgeTriggered bool) error {
	var flags IOFlags = unix.EV_ADD
	if edgeTriggered {
		flags |= unix.EV_CLEAR
	}
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: keventIdent(pa.FD), Flags: flags, Filter: unix.EVFILT_READ},
	}, nil, nil)
	return os.NewSyscallError("kevent add", err)
}

// AddWrite registers the given file descriptor with writable event to the poller.
func (p *Poller) AddWrite(pa *PollAttachment, edgeTriggered bool) error {
	var flags IOFlags = unix.EV_ADD
	if edgeTriggered {
		flags |= unix.EV_CLEAR
	}
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: keventIdent(pa.FD), Flags: flags, Filter: unix.EVFILT_WRITE},
	}, nil, nil)
	return os.NewSyscallError("kevent add", err)
}

// ModRead modifies the given file descriptor with readable event in the poller.
func (p *Poller) ModRead(pa *PollAttachment, _ bool) error {
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: keventIdent(pa.FD), Flags: unix.EV_DELETE, Filter: unix.EVFILT_WRITE},
	}, nil, nil)
	return os.NewSyscallError("kevent delete", err)
}

// ModReadWrite modifies the given file descriptor with readable and writable events in the poller.
func (p *Poller) ModReadWrite(pa *PollAttachment, edgeTriggered bool) error {
	var flags IOFlags = unix.EV_ADD
	if edgeTriggered {
		flags |= unix.EV_CLEAR
	}
	_, err := unix.Kevent(p.fd, []unix.Kevent_t{
		{Ident: keventIdent(pa.FD), Flags: flags, Filter: unix.EVFILT_WRITE},
	}, nil, nil)
	return os.NewSyscallError("kevent add", err)
}

// Delete removes the given file descriptor from the poller.
func (*Poller) Delete(_ int) error {
	return nil
}
