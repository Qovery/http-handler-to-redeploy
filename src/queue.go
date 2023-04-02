package src

import (
	"sync"
)

type Queue[T any] struct {
	items          []T
	cond           *sync.Cond
	capacity       int
	circularBuffer bool
}

func NewQueue[T any](capacity int, circularBuffer bool) *Queue[T] {
	return &Queue[T]{
		items:          make([]T, 0),
		cond:           sync.NewCond(&sync.Mutex{}),
		capacity:       capacity,
		circularBuffer: circularBuffer,
	}
}

func (q *Queue[T]) Enqueue(item T) bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if len(q.items) >= q.capacity {
		if !q.circularBuffer {
			return false
		}

		// remove first item
		q.items = q.items[1:]
	}

	q.items = append(q.items, item)
	q.cond.Signal()
	return true
}

func (q *Queue[T]) Dequeue() (T, bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for len(q.items) == 0 {
		q.cond.Wait()
	}

	if len(q.items) > 0 {
		item := q.items[0]
		q.items = q.items[1:]
		return item, true
	}

	var zero T
	return zero, false
}
