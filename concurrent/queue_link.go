package concurrent

import (
    "runtime"
    "sync"
    "sync/atomic"
    "unsafe"
)

//cas 无锁链表
type lkQueue[T any] struct {
    useLock bool
    elock   sync.Mutex
    dlock   sync.Mutex

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
    return func(opt opt[T]) {
        opt.Type = newLinkedQueue[T]
    }
}

var (
    nodePool *sync.Pool
    //nodePool = &sync.Pool{New: func() any { return &node{} }}
)

func newLinkedQueue[T any]() Queue[T] {
    var start unsafe.Pointer
    if nodePool == nil {
        start = unsafe.Pointer(&node{})
    } else {
        start = unsafe.Pointer(nodePool.Get().(*node))
    }
    q := &lkQueue[T]{head: start, tail: start}
    return q
}
func (q *lkQueue[T]) Size() int {
    return int(atomic.LoadInt32(&q.size))
}
func (q *lkQueue[T]) newNode() *node {
    if nodePool == nil {
        return &node{}
    }
    return nodePool.Get().(*node)
}
func (q *lkQueue[T]) reclaimNode(n *node) {
    if nodePool == nil {
        return
    }
    n.value = nil
    atomic.StorePointer(&n.next, nil)
    nodePool.Put(n)
}

var (
    add1 = Int64Adder{}
    add2 = Int64Adder{}
    add3 = Int64Adder{}
    add4 = Int64Adder{}
)

func (q *lkQueue[T]) Enqueue(v T) {
    if any(v) == nil {
        return
    }
    n := q.newNode()
    n.value = v
    if q.useLock {
        q.elock.Lock()
        np := unsafe.Pointer(n)
        (*node)(q.tail).next = np
        q.tail = np
        q.elock.Unlock()
    } else {
        var tail *node
        for {
            // 1.读取当前tail
            tail = casLoad(&q.tail)
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
                    add1.IncrementSimple()
                    continue
                }
                //尝试将tail.next设置为tail,然后再次尝试插入tail
                atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
                add2.IncrementSimple()
            }
        }
    }
}

func (q *lkQueue[T]) Dequeue() (T, bool) {
    if q.useLock {
        q.dlock.Lock()
        np := (*node)(q.head).next
        if np == nil {
            q.dlock.Unlock()
            var r T
            return r, false
        }
        next := (*node)(np)
        v := next.value.(T)
        q.reclaimNode(next)
        q.dlock.Unlock()
        return v, true
    } else {
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
                add3.IncrementSimple()
                continue
            }
            if atomic.CompareAndSwapPointer(&q.head, hp, np) {
                atomic.AddInt32(&q.size, -1)
                ////如果queue为空,且消费比插入还快,则尝试更新tail
                //if atomic.CompareAndSwapPointer(&q.tail, hp, np) {
                //    println("111")
                //}
                v := next.value.(T)
                q.reclaimNode(head)
                return v, true
            } else {
                add4.IncrementSimple()
            }
        }
    }
}

func casLoad(p *unsafe.Pointer) (n *node) {
    return (*node)(atomic.LoadPointer(p))
}
