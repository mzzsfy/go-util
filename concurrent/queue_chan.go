package concurrent

import (
	"sync"
	"time"
)

// chan包装队列,固定容量,MPMC安全
// 直接利用channel的阻塞语义,零分配,性能接近原生channel
type chanQueue[T any] struct {
	ch chan T
	// timerPool 复用超时 timer, 避免每次 DequeueBlock 分配
	timerPool sync.Pool
}

// WithTypeChan 使用channel包装队列,固定容量,零分配
// buffer为缓冲区大小,队列满时Enqueue会阻塞
func WithTypeChan[T any](buffer int) Opt[T] {
	return func(opt *opt[T]) {
		opt.Type = func() Queue[T] {
			return newChanQueue[T](buffer)
		}
	}
}

func newChanQueue[T any](buffer int) Queue[T] {
	if buffer < 1 {
		buffer = 1
	}
	q := &chanQueue[T]{ch: make(chan T, buffer)}
	q.timerPool.New = func() any { return time.NewTimer(maxTimeoutDuration) }
	return q
}

func (q *chanQueue[T]) Enqueue(v T) {
	q.ch <- v
}

func (q *chanQueue[T]) Dequeue() (T, bool) {
	select {
	case v := <-q.ch:
		return v, true
	default:
		var zero T
		return zero, false
	}
}

func (q *chanQueue[T]) TryDequeue() (T, bool) {
	select {
	case v := <-q.ch:
		return v, true
	default:
		var zero T
		return zero, false
	}
}

func (q *chanQueue[T]) Size() int {
	return len(q.ch)
}

func (q *chanQueue[T]) DequeueBlock(timeout ...time.Duration) (T, bool) {
	// 快速路径: 非阻塞尝试,避免timer分配
	select {
	case v := <-q.ch:
		return v, true
	default:
	}
	if len(timeout) > 0 {
		tm := q.timerPool.Get().(*time.Timer)
		// 排空残留信号, 归池前 channel 必为空
		if !tm.Stop() {
			select {
			case <-tm.C:
			default:
			}
		}
		tm.Reset(timeout[0])
		var v T
		var ok bool
		select {
		case v = <-q.ch:
			ok = true
		case <-tm.C:
		}
		tm.Stop()
		select {
		case <-tm.C:
		default:
		}
		q.timerPool.Put(tm)
		return v, ok
	}
	v, ok := <-q.ch
	return v, ok
}
