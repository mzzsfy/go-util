package concurrent

import (
    "time"
)

type Queue[T any] interface {
    Enqueue(v T)
    Dequeue() (T, bool)
    Size() int
}
type BlockQueue[T any] interface {
    Queue[T]
    DequeueBlock(timeout ...time.Duration) (T, bool)
}

type opt[T any] struct {
    Type func() Queue[T]
    opt  []func(Queue[T]) Queue[T]
}

type Opt[T any] func(*opt[T])

func NewQueue[T any](opts ...Opt[T]) Queue[T] {
    if len(opts) == 0 {
        return newLinkedQueue[T]()
    }
    opt := &opt[T]{
        Type: newLinkedQueue[T],
    }
    for _, o := range opts {
        o(opt)
    }
    r := opt.Type()
    for _, f := range opt.opt {
        r = f(r)
    }
    return r
}
