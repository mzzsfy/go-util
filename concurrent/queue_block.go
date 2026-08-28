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
	bq := &blockQueue[T]{
		Queue: queue,
		mu:    mu,
		cond:  sync.NewCond(mu),
	}
	bq.timerPool.New = func() any {
		return time.AfterFunc(maxTimeoutDuration, bq.cond.Broadcast)
	}
	return bq
}

// maxTimeoutDuration 池化 timer 的初始挂载时长, 仅为占位, 使用前必然 Reset
const maxTimeoutDuration = time.Duration(1<<62 - 1)

// 自旋档位: 纯自旋覆盖短时唤醒窗口避免 park; Windows 低精度调度下
// park 唤醒延迟可达毫秒级, 档位过低会使高频场景整体吞吐悬崖式退化
const (
	spinPureMax      = 300
	spinTotalMax     = 3000
	spinGoschedEvery = 64
)

// spinTryDequeue 自旋尝试出队: 前段纯自旋, 后段间歇让出
func spinTryDequeue[T any](td TryDequeuer[T]) (T, bool) {
	for i := 0; i < spinTotalMax; i++ {
		if v, ok := td.TryDequeue(); ok {
			return v, true
		}
		if i >= spinPureMax && i%spinGoschedEvery == 0 {
			runtime.Gosched()
		}
	}
	var zero T
	return zero, false
}

type blockQueue[T any] struct {
	Queue[T]
	_      [6]int64
	waiter int32
	mu     *sync.Mutex
	cond   *sync.Cond
	// timerPool 复用超时 timer, 回调固定为广播, Reset 换触发时长
	timerPool sync.Pool
}

func (q *blockQueue[T]) getTimeoutTimer() *time.Timer {
	t := q.timerPool.Get()
	if t == nil {
		return time.AfterFunc(maxTimeoutDuration, q.cond.Broadcast)
	}
	return t.(*time.Timer)
}

func (q *blockQueue[T]) putTimeoutTimer(t *time.Timer) {
	// AfterFunc 型 timer 回调仅为广播, Stop 失败(已触发)亦无副作用
	t.Stop()
	q.timerPool.Put(t)
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
		if v, ok := spinTryDequeue(td); ok {
			return v, true
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
	if v, ok := spinTryDequeue(td); ok {
		return v, true
	}
	q.mu.Lock()
	atomic.AddInt32(&q.waiter, 1)
	timer := q.getTimeoutTimer()
	defer func() {
		q.putTimeoutTimer(timer)
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
		timer.Reset(deadline.Sub(now))
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
