package concurrent

import (
	"runtime"
	"sync"
	"time"
)

func BlockQueueWrapper[T any](queue Queue[T]) BlockQueue[T] {
	if q, ok := queue.(BlockQueue[T]); ok {
		return q
	}
	mu := &sync.Mutex{}
	return &blockQueue[T]{
		Queue: queue,
		mu:    mu,
		cond:  sync.NewCond(mu),
	}
}

type blockQueue[T any] struct {
	Queue[T]
	_     [6]int64
	waiter int32
	mu    *sync.Mutex
	cond  *sync.Cond
}

func (q *blockQueue[T]) Enqueue(v T) {
	q.Queue.Enqueue(v)
	q.mu.Lock()
	w := q.waiter
	if w > 1 {
		q.cond.Broadcast()
	} else if w > 0 {
		q.cond.Signal()
	}
	q.mu.Unlock()
}

func (q *blockQueue[T]) DequeueBlock(timeout ...time.Duration) (T, bool) {
	if len(timeout) > 0 {
		return q.dequeueBlockWithTimeout(timeout[0])
	}
	return q.dequeueBlockForever()
}

func (q *blockQueue[T]) dequeueBlockForever() (T, bool) {
	if td, ok := q.Queue.(TryDequeuer[T]); ok {
		// 快速路径: 无锁spin
		for i := 0; i < 4; i++ {
			v, b := td.TryDequeue()
			if b {
				return v, true
			}
			runtime.Gosched()
		}
		// 慢路径: Cond等待
		q.mu.Lock()
		q.waiter++
		for {
			v, b := td.TryDequeue()
			if b {
				q.waiter--
				q.mu.Unlock()
				return v, true
			}
			q.cond.Wait()
		}
	}
	for {
		v, b := q.Dequeue()
		if b {
			return v, true
		}
		runtime.Gosched()
	}
}

func (q *blockQueue[T]) dequeueBlockWithTimeout(timeout time.Duration) (T, bool) {
	if td, ok := q.Queue.(TryDequeuer[T]); ok {
		return q.dequeueWithTimeoutTry(td, timeout)
	}
	return q.dequeueWithTimeoutFallback(timeout)
}

func (q *blockQueue[T]) dequeueWithTimeoutTry(td TryDequeuer[T], timeout time.Duration) (T, bool) {
	deadline := time.Now().Add(timeout)
	for i := 0; i < 4; i++ {
		v, b := td.TryDequeue()
		if b {
			return v, true
		}
		runtime.Gosched()
	}
	q.mu.Lock()
	q.waiter++
	defer func() {
		q.waiter--
		q.mu.Unlock()
	}()
	timer := time.AfterFunc(timeout, func() { q.cond.Broadcast() })
	defer timer.Stop()
	for {
		v, b := td.TryDequeue()
		if b {
			return v, true
		}
		if time.Now().After(deadline) {
			var r T
			return r, false
		}
		q.cond.Wait()
	}
}

func (q *blockQueue[T]) dequeueWithTimeoutFallback(timeout time.Duration) (T, bool) {
	deadline := time.Now().Add(timeout)
	for {
		v, b := q.Dequeue()
		if b {
			return v, true
		}
		if time.Now().After(deadline) {
			return v, false
		}
		runtime.Gosched()
	}
}
