package queue

import (
	"sync/atomic"
	"unsafe"
)

// lockFreeQueue is a simple, fast, and practical non-blocking and concurrent queue with no lock.
type lockFreeQueue struct {
	head   unsafe.Pointer
	tail   unsafe.Pointer
	length int32
}

type node struct {
	value *Task
	next  unsafe.Pointer
}

// NewLockFreeQueue instantiates and returns a lockFreeQueue.
func NewLockFreeQueue() AsyncTaskQueue {
	n := unsafe.Pointer(&node{})
	return &lockFreeQueue{head: n, tail: n}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *lockFreeQueue) Enqueue(task *Task) {
	n := &node{value: task}
retry:
	tail := load(&q.tail)
	next := load(&tail.next)
	// Are tail and next consistent?
	if tail == load(&q.tail) {
		if next == nil {
			// Try to link node at the end of the linked list.
			if cas(&tail.next, next, n) { // enqueue is done.
				// Try to swing tail to the inserted node.
				cas(&q.tail, tail, n)
				atomic.AddInt32(&q.length, 1)
				return
			}
		} else { // tail was not pointing to the last node
			// Try to swing tail to the next node.
			cas(&q.tail, tail, next)
		}
	}
	goto retry
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *lockFreeQueue) Dequeue() *Task {
retry:
	head := load(&q.head)
	tail := load(&q.tail)
	next := load(&head.next)
	// Are head, tail, and next consistent?
	if head == load(&q.head) {
		// Is queue empty or tail falling behind?
		if head == tail {
			// Is queue empty?
			if next == nil {
				return nil
			}
			cas(&q.tail, tail, next) // tail is falling behind, try to advance it.
		} else {
			// Read value before CAS, otherwise another dequeue might free the next node.
			task := next.value
			if cas(&q.head, head, next) { // dequeue is done, return value.
				atomic.AddInt32(&q.length, -1)
				return task
			}
		}
	}
	goto retry
}

// IsEmpty indicates whether this queue is empty or not.
func (q *lockFreeQueue) IsEmpty() bool {
	return atomic.LoadInt32(&q.length) == 0
}

// Length returns the number of elements in the queue.
func (q *lockFreeQueue) Length() int32 {
	return atomic.LoadInt32(&q.length)
}

func load(p *unsafe.Pointer) (n *node) {
	return (*node)(atomic.LoadPointer(p))
}

func cas(p *unsafe.Pointer, old, new *node) bool { //nolint:revive
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
