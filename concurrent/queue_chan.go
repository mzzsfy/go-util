package concurrent

import (
	"time"
)

// chan包装队列,固定容量,MPMC安全
// 直接利用channel的阻塞语义,零分配,性能接近原生channel
type chanQueue[T any] struct {
	ch chan T
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
	return &chanQueue[T]{ch: make(chan T, buffer)}
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
		select {
		case v := <-q.ch:
			return v, true
		case <-time.After(timeout[0]):
			var zero T
			return zero, false
		}
	}
	v, ok := <-q.ch
	return v, ok
}
