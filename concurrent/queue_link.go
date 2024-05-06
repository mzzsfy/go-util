package concurrent

import (
    "runtime"
    "sync"
    "sync/atomic"
    "unsafe"
)

//cas 链表
type lkQueue[T any] struct {
    newNode     func() *node
    reclaimNode func(n *node)

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

var (
    hit   = Int64Adder{}
    noHit = Int64Adder{}
)

func newLinkedQueue[T any]() Queue[T] {
    start := unsafe.Pointer(&node{})
    q := &lkQueue[T]{head: start, tail: start}
    if nodePool != nil {
        q.newNode = func() *node {
            for {
                n := nodePool.Get().(*node)
                if n.next == nil && n.value == nil {
                    hit.IncrementSimple()
                    return n
                } else {
                    noHit.IncrementSimple()
                }
            }
        }
        q.reclaimNode = func(n *node) {
            atomic.StorePointer(&n.next, nil)
            nodePool.Put(n)
        }
    }
    return q
}
func (q *lkQueue[T]) Size() int {
    return int(atomic.LoadInt32(&q.size))
}

func (q *lkQueue[T]) Enqueue(v T) {
    if any(v) == nil {
        return
    }
    var n *node
    if q.newNode == nil {
        n = &node{}
    } else {
        n = q.newNode()
    }
    n.value = v
    var tail *node
    for {
        // 1.读取当前tail
        tail = casLoad(&q.tail)
        // 设置tail.next为当前值
        if casSwap(&tail.next, nil, unsafe.Pointer(n)) {
            //无竞争

            atomic.AddInt32(&q.size, 1)
            //尝试将new node设置为tail
            //若有竞争,则head.next正确,tail不正确,不影响遍历,影响插入,Enqueue重试即可
            casSwap(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(n))
            return
        } else {
            //有竞争

            next := casLoad(&tail.next)
            //可能已经被消费,所以next为nil
            if next == nil || next.value == nil {
                continue
            }
            //尝试将tail.next设置为tail,然后再次尝试插入tail
            casSwap(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
        }
    }
}

func (q *lkQueue[T]) Dequeue() (T, bool) {
    var (
        head *node
        next *node
    )
    i := 0
    for {
        hp := atomic.LoadPointer(&q.head)
        head = (*node)(hp)
        np := atomic.LoadPointer(&head.next)
        next = (*node)(np)
        // next.value为当前值
        if next == nil {
            //可能是插入未完成
            if q.size == 0 {
                var r T
                return r, false
            } else {
                i++
                runtime.Gosched()
                if i > 100 {
                    //tail := casLoad(&q.tail)
                    //fmt.Println("1", hp, head, tail)
                    //time.Sleep(300 * time.Millisecond)
                    //panic("")
                }
                continue
            }
        }
        //if next.value == nil {
        //    tail := casLoad(&q.tail)
        //    fmt.Println("2", hp, np, head, tail)
        //    time.Sleep(300 * time.Millisecond)
        //    panic("")
        //}
        if casSwap(&q.head, hp, np) {
            atomic.AddInt32(&q.size, -1)
            v := next.value.(T)
            next.value = nil
            if q.reclaimNode != nil {
                q.reclaimNode(head)
            }
            return v, true
        }
    }
}

func casLoad(p *unsafe.Pointer) (n *node) {
    return (*node)(atomic.LoadPointer(p))
}

func casSwap(p *unsafe.Pointer, old, new unsafe.Pointer) (ok bool) {
    return atomic.CompareAndSwapPointer(p, old, new)
}
