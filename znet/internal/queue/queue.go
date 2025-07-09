package queue

import "sync"

// Func is the callback function executed by poller.
type Func func(any) error

// Task is a wrapper that contains function and its argument.
type Task struct {
	Exec  Func
	Param any
}

var taskPool = sync.Pool{New: func() any { return new(Task) }}

// GetTask gets a cached Task from pool.
func GetTask() *Task {
	return taskPool.Get().(*Task)
}

// PutTask puts the trashy Task back in pool.
func PutTask(task *Task) {
	task.Exec, task.Param = nil, nil
	taskPool.Put(task)
}

// AsyncTaskQueue is a queue storing asynchronous tasks.
type AsyncTaskQueue interface {
	Enqueue(*Task)
	Dequeue() *Task
	IsEmpty() bool
	Length() int32
}

// EventPriority is the priority of an event.
type EventPriority int

const (
	// HighPriority is for the tasks expected to be executed
	// as soon as possible.
	HighPriority EventPriority = iota
	// LowPriority is for the tasks that won't matter much
	// even if they are deferred a little bit.
	LowPriority
)
