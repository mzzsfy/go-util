package concurrent

import (
    "sync/atomic"
    "time"
)

func WithTypeDelay[T any](delay time.Duration) Opt[T] {
    return func(opt *opt[T]) {
        opt.Type = func() Queue[T] {
            return newDelayQueue[T](delay)
        }
    }
}
func newDelayQueue[T any](delayTime time.Duration) Queue[T] {
    return &delayQueue[T]{
        in:    newLinkedQueue[delay[T]](),
        out:   newLinkedQueue[T](),
        delay: delayTime,
    }
}

// 延时队列
type delay[T any] struct {
    t     time.Time
    value T
}

type delayQueue[T any] struct {
    in      Queue[delay[T]]
    out     Queue[T]
    delay   time.Duration
    running int32
}

func (q *delayQueue[T]) Size() int {
    return q.in.Size() + q.out.Size()
}

func (q *delayQueue[T]) start() {
    if atomic.CompareAndSwapInt32(&q.running, 0, 1) {
        go func() {
            defer atomic.StoreInt32(&q.running, 0)
            for {
                v, b := q.in.Dequeue()
                if !b {
                    return
                }
                if time.Now().Sub(v.t) < q.delay {
                    time.Sleep(q.delay - time.Now().Sub(v.t))
                }
                q.out.Enqueue(v.value)
            }
        }()
    }
}

func (q *delayQueue[T]) Enqueue(v T) {
    q.in.Enqueue(delay[T]{t: time.Now(), value: v})
    q.start()
}

func (q *delayQueue[T]) Dequeue() (T, bool) {
    return q.out.Dequeue()
}
