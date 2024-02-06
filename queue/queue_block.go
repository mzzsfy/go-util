package queue

import (
    "runtime"
    "sync/atomic"
    "time"
)

func BlockQueueWrapper[T any](queue Queue[T]) BlockQueue[T] {
    if q, ok := queue.(BlockQueue[T]); ok {
        return q
    }
    return &blockQueue[T]{Queue: queue, notify: newLinkedQueue[func(*blockQueue[T]) bool]()}
}

// 阻塞队列,使用了chan方案,高并发下性能不好
type blockQueue[T any] struct {
    Queue[T]
    _      [6]int64
    notify Queue[func(*blockQueue[T]) bool]
}

func (q *blockQueue[T]) Enqueue(v T) {
    q.Queue.Enqueue(v)
    var n func(*blockQueue[T]) bool
    var b bool
    for {
        n, b = q.notify.Dequeue()
        if !b || (n)(q) {
            break
        }
    }
}

func (q *blockQueue[T]) DequeueBlock(timeout ...time.Duration) (T, bool) {
    if len(timeout) > 0 {
        return q.DequeueBlockByTimeout(timeout[0])
    } else {
        var t T
        var b bool
        for {
            t, b = q.DequeueBlockByTimeout(time.Minute)
            if b {
                return t, b
            }
        }
    }
}

func (q *blockQueue[T]) DequeueBlockByTimeout(timeout time.Duration) (T, bool) {
    t, b := q.Dequeue()
    if b {
        return t, b
    }
    if timeout <= 0 {
        return t, false
    }
    ch := time.After(timeout)
    for i := 0; i < 5; i++ {
        if q.Size() != 0 {
            break
        }
        t, b = q.Dequeue()
        if b {
            return t, b
        }
        runtime.Gosched()
    }
    c := make(chan struct{}, 1)
    i := int32(0)
    defer atomic.AddInt32(&i, 1)
    for {
        q.notify.Enqueue(func(_ *blockQueue[T]) bool {
            if atomic.LoadInt32(&i) > 0 {
                return false
            } else {
                c <- struct{}{}
            }
            return atomic.LoadInt32(&i) == 0
        })
        select {
        case <-ch:
            return t, false
        case <-c:
            t, b = q.Dequeue()
            if b {
                return t, b
            }
        }
    }
}
