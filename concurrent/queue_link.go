package concurrent

import (
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
    //poolLock = sync.RWMutex{}
    //nodePools = make(map[string]any)
    nodePool = &sync.Pool{New: func() any { return &node{} }}
)

func getNodePool[T any]() *sync.Pool {
    return nil
    //return nodePool
    //name := reflect.TypeOf((*T)(nil)).Elem().String()
    //poolLock.RLock()
    //if pool, ok := nodePools[name]; ok {
    //    poolLock.RUnlock()
    //    return pool.(*sync.Pool)
    //}
    //poolLock.RUnlock()
    //poolLock.Lock()
    //defer poolLock.Unlock()
    //p := &sync.Pool{New: func() any { return &node[T]{} }}
    //nodePools[name] = p
    //return p
}

var (
    hit   = Int64Adder{}
    noHit = Int64Adder{}
)

func newLinkedQueue[T any]() Queue[T] {
    start := unsafe.Pointer(&node{})
    q := &lkQueue[T]{head: start, tail: start}
    pool := getNodePool[T]()
    if pool != nil {
        q.newNode = func() *node {
            for {
                n := pool.Get().(*node)
                if n.next == nil && n.value == nil {
                    hit.IncrementSimple()
                    return n
                } else {
                    noHit.IncrementSimple()
                }
            }
        }
        q.reclaimNode = func(n *node) {
            n.next = nil
            pool.Put(n)
        }
    }
    return q
}
func (q *lkQueue[T]) Size() int {
    return int(atomic.LoadInt32(&q.size))
}

func (q *lkQueue[T]) Enqueue(v T) {
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
            if next == nil || next == tail {
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
    for {
        head = casLoad(&q.head)
        next = casLoad(&head.next)
        // next.value为当前值,所以next.value没值则队列为空
        if next == nil {
            var r T
            return r, false
        }
        if casSwap(&q.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
            v := next.value
            if q.reclaimNode != nil {
                q.reclaimNode(head)
            }
            next.value = nil
            atomic.AddInt32(&q.size, -1)
            return v.(T), true
        }
    }
}

func casLoad(p *unsafe.Pointer) (n *node) {
    return (*node)(atomic.LoadPointer(p))
}

func casSwap(p *unsafe.Pointer, old, new unsafe.Pointer) (ok bool) {
    return atomic.CompareAndSwapPointer(p, old, new)
}
