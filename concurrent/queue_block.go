package concurrent

import (
	"runtime"
	"sync"
	"sync/atomic"
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
	_      [6]int64
	waiter int32
	mu     *sync.Mutex
	cond   *sync.Cond
}

func (q *blockQueue[T]) Enqueue(v T) {
	q.Queue.Enqueue(v)
	// 快速路径: 无等待者时跳过锁, 高吞吐场景(worker全忙碌)零Mutex开销
	if atomic.LoadInt32(&q.waiter) == 0 {
		return
	}
	q.mu.Lock()
	if w := atomic.LoadInt32(&q.waiter); w > 1 {
		q.cond.Broadcast()
	} else if w > 0 {
		q.cond.Signal()
	}
	q.mu.Unlock()
}

func (q *blockQueue[T]) WaiterCount() int32 {
	return atomic.LoadInt32(&q.waiter)
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
		atomic.AddInt32(&q.waiter, 1)
		for {
			v, b := td.TryDequeue()
			if b {
				atomic.AddInt32(&q.waiter, -1)
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
	atomic.AddInt32(&q.waiter, 1)
	// 延迟创建 timer: 仅在需要 Wait 时才创建, 避免慢路径首次 TryDequeue 命中时的 timer 创建开销
	// 使用剩余时间(deadline-now)保证在 deadline 触发, 比原方案(完整timeout)更精确
	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
		atomic.AddInt32(&q.waiter, -1)
		q.mu.Unlock()
	}()
	for {
		v, b := td.TryDequeue()
		if b {
			return v, true
		}
		now := time.Now()
		if !now.Before(deadline) {
			var r T
			return r, false
		}
		if timer == nil {
			timer = time.AfterFunc(deadline.Sub(now), func() { q.cond.Broadcast() })
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
