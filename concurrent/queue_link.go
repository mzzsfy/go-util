package concurrent

import (
    "reflect"
    "sync"
    "sync/atomic"
    "unsafe"
)

//cas 链表
type lkQueue[T any] struct {
    newNode     func() *node[T]
    reclaimNode func(n *node[T])

    head unsafe.Pointer
    _    [7]int64
    tail unsafe.Pointer
    _    [7]int64
    size int32
}
type node[T any] struct {
    value T
    next  unsafe.Pointer
}

func WithTypeLink[T any]() Opt[T] {
    return func(opt opt[T]) {
        opt.Type = newLinkedQueue[T]
    }
}

var (
    poolLock  = sync.RWMutex{}
    nodePools = make(map[string]any)
)

func getNodePool[T any]() *sync.Pool {
    name := reflect.TypeOf((*T)(nil)).Elem().String()
    poolLock.RLock()
    if pool, ok := nodePools[name]; ok {
        poolLock.RUnlock()
        return pool.(*sync.Pool)
    }
    poolLock.RUnlock()
    poolLock.Lock()
    defer poolLock.Unlock()
    p := &sync.Pool{
        New: func() any {
            return &node[T]{}
        },
    }
    nodePools[name] = p
    return p
}

func newLinkedQueue[T any]() Queue[T] {
    n := unsafe.Pointer(&node[T]{})
    q := &lkQueue[T]{head: n, tail: n}
    pool := getNodePool[T]()
    if pool != nil {
        q.newNode = func() *node[T] {
            return pool.Get().(*node[T])
        }
        var t T
        q.reclaimNode = func(n *node[T]) {
            n.value = t
            pool.Put(n)
        }
    }
    return q
}
func (q *lkQueue[T]) Size() int {
    return int(atomic.LoadInt32(&q.size))
}

func (q *lkQueue[T]) Enqueue(v T) {
    var n *node[T]
    if q.newNode == nil {
        n = &node[T]{value: v}
    } else {
        n = q.newNode()
        n.value = v
        n.next = nil
    }
    var tail *node[T]
    for {
        // 1.读取当前tail
        tail = q.load(&q.tail)
        //无竞争
        if q.casTT(&tail.next, nil, n) {
            // 设置tail为当前值
            //若无竞争,则head.next和插入都正常,不影响遍历
            //若有竞争,则head.next正确,tail不正确,不影响遍历,影响插入
            q.casTT(&q.tail, tail, n)
            atomic.AddInt32(&q.size, 1)
            return
        } else /*有竞争*/ {
            //尝试将next设置为tail
            q.casTP(&q.tail, tail, tail.next)
        }
        //总结,每个循环中
        //尝试使用n替换tail.next(如果nil)
        //否则,尝试使用tail.next替换tail,也就是移动一步
    }
}

func (q *lkQueue[T]) Dequeue() (T, bool) {
    var (
        head *node[T]
        next *node[T]
        v    T
    )
    for {
        head = q.load(&q.head)
        next = q.load(&head.next)
        // next.value为当前值,所以next.value没值则队列为空
        if next == nil {
            return v, false
        }
        //确保head 与 next 读取数据一致
        if head == q.load(&q.head) {
            v = next.value
            if q.casTT(&q.head, head, next) {
                atomic.AddInt32(&q.size, -1)
                if q.reclaimNode != nil {
                    q.reclaimNode(head)
                }
                return v, true
            }
        }
    }
}

func (q *lkQueue[T]) load(p *unsafe.Pointer) (n *node[T]) {
    return (*node[T])(atomic.LoadPointer(p))
}
func (q *lkQueue[T]) casTT(p *unsafe.Pointer, old, new *node[T]) (ok bool) {
    return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
func (q *lkQueue[T]) casPT(p *unsafe.Pointer, old unsafe.Pointer, new *node[T]) (ok bool) {
    return atomic.CompareAndSwapPointer(p, old, unsafe.Pointer(new))
}
func (q *lkQueue[T]) casTP(p *unsafe.Pointer, old *node[T], new unsafe.Pointer) (ok bool) {
    return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), new)
}
