package concurrent

import (
    "runtime"
    "sync/atomic"
    "unsafe"
)

//cas 无锁链表
type lkQueue[T any] struct {
    head unsafe.Pointer
    _    [cpuCacheKillerPaddingLength]byte
    tail unsafe.Pointer
    _    [cpuCacheKillerPaddingLength]byte
    size int32
}
type node struct {
    value any
    next  unsafe.Pointer
}

func WithTypeLink[T any]() Opt[T] {
    return func(opt *opt[T]) {
        opt.Type = newLinkedQueue[T]
    }
}

func newLinkedQueue[T any]() Queue[T] {
    var start = unsafe.Pointer(newNode())
    q := &lkQueue[T]{head: start, tail: start}
    return q
}
func (q *lkQueue[T]) Size() int {
    return int(atomic.LoadInt32(&q.size))
}
func newNode() *node {
    return &node{}
}
func reclaimNode(_ *node) {
}

func (q *lkQueue[T]) Enqueue(v T) {
    if any(v) == nil {
        return
    }
    n := newNode()
    n.value = v
    for {
        // 1.读取当前tail
        tail := casLoad(&q.tail)
        // 设置tail.next为当前值
        if atomic.CompareAndSwapPointer(&tail.next, nil, unsafe.Pointer(n)) {
            //无竞争
            atomic.AddInt32(&q.size, 1)
            //尝试将new node设置为tail
            //若有竞争,则head.next正确,tail不正确,不影响遍历,影响插入,Enqueue重试即可
            atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(n))
            return
        } else {
            //有竞争
            next := casLoad(&tail.next)
            //可能已经被消费,所以next为nil
            if next == nil {
                continue
            }
            //尝试将tail.next设置为tail,然后再次尝试插入tail
            atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
        }
    }
}

func (q *lkQueue[T]) Dequeue() (T, bool) {
    for {
        hp := atomic.LoadPointer(&q.head)
        head := (*node)(hp)
        np := atomic.LoadPointer(&head.next)
        next := (*node)(np)
        // next.value为当前值
        if next == nil {
            //可能是插入未完成
            if q.size < 128 {
                if atomic.LoadInt32(&q.size) == 0 {
                    var r T
                    return r, false
                }
            }
            runtime.Gosched()
            continue
        }
        if atomic.CompareAndSwapPointer(&q.head, hp, np) {
            atomic.AddInt32(&q.size, -1)
            ////如果queue为空,且消费比插入还快,则尝试更新tail
            //if atomic.CompareAndSwapPointer(&q.tail, hp, np) {
            //    println("111")
            //}
            v := next.value.(T)
            reclaimNode(head)
            return v, true
        }
    }
}

func casLoad(p *unsafe.Pointer) (n *node) {
    return (*node)(atomic.LoadPointer(p))
}
