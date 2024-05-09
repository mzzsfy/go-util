package concurrent

import (
    "runtime"
    "sync/atomic"
    "unsafe"
)

//链表+数组实现队列,无锁

type layer[T any] struct {
    id  uint64
    arr []struct {
        set bool
        v   T
    }
    _    [cpuCacheKillerPaddingLength]byte
    next unsafe.Pointer
    take uint32
}

type lkArrQueue[T any] struct {
    zero      T
    layerMask uint64 //层掩码

    produceId uint64
    _         [cpuCacheKillerPaddingLength]byte
    consumeId uint64
    _         [cpuCacheKillerPaddingLength]byte

    //head.next为当前层
    head unsafe.Pointer
    _    [cpuCacheKillerPaddingLength]byte
    tail unsafe.Pointer
}

func (q *lkArrQueue[T]) Enqueue(v T) {
    //获取pid
    pid := atomic.AddUint64(&q.produceId, 1)
    //获取层中的索引
    lIdx := pid & q.layerMask
    //根据pid获取layerId
    layerId := pid / (q.layerMask + 1)
    //新建layer
    if lIdx == 0 {
        l := newLayer[T](layerId, int(q.layerMask+1))
        l.arr[0].v = v
        l.arr[0].set = true
        lp := unsafe.Pointer(l)
        for i := 0; ; i++ {
            tail := (*layer[T])(atomic.LoadPointer(&q.tail))
            if tail.id != layerId-1 {
                if i > 10 {
                    runtime.Gosched()
                }
                continue
            }
            atomic.StorePointer(&tail.next, lp)
            if atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), lp) {
                break
            }
        }
        return
    } else {
        //获取tail
        for i := 0; ; i++ {
            tail := (*layer[T])(atomic.LoadPointer(&q.tail))
            if tail.id == layerId {
                tail.arr[lIdx].v = v
                tail.arr[lIdx].set = true
                return
            }
            //tail已经被替换,从head开始找到当前的layer
            if tail.id > layerId {
                head := (*layer[T])(atomic.LoadPointer(&q.head))
                for {
                    if head.id == layerId {
                        head.arr[lIdx] = struct {
                            set bool
                            v   T
                        }{
                            set: true,
                            v:   v,
                        }
                        return
                    }
                    head = (*layer[T])(head.next)
                }
            } else {
                if i > 10 {
                    runtime.Gosched()
                }
            }
        }
    }
}

func (q *lkArrQueue[T]) Dequeue() (T, bool) {
    for {
        pid := atomic.LoadUint64(&q.produceId)
        cid := atomic.LoadUint64(&q.consumeId) + 1
        if pid <= cid {
            return q.zero, false
        }
        hp := atomic.LoadPointer(&q.head)
        np := (*layer[T])(hp).next
        next := (*layer[T])(np)
        if next == nil {
            continue
        }
        layerId := cid / (q.layerMask + 1)
        //当前层
        if next.id == layerId {
            if atomic.CompareAndSwapUint64(&q.consumeId, cid-1, cid) {
                t, b := q.getValue(cid&q.layerMask, next)
                if b {
                    return t, b
                }
            }
            continue
        } else
        //下一层 
        if next.id == layerId-1 {
            if atomic.CompareAndSwapUint64(&q.consumeId, cid-1, cid) {
                var l *layer[T]
                for i := 0; ; i++ {
                    l = (*layer[T])(next.next)
                    if l != nil {
                        break
                    } else if i > 10 {
                        runtime.Gosched()
                    }
                }
                if t, ok := q.getValue(cid&q.layerMask, l); ok {
                    q.fixHead(atomic.LoadUint32(&l.take) > uint32(q.layerMask-4), hp, unsafe.Pointer(next))
                    return t, true
                }
            }
            continue
        } else
        //2层 
        if next.id == layerId-2 {
            if atomic.CompareAndSwapUint64(&q.consumeId, cid-1, cid) {
                var l *layer[T]
                for i := 0; ; i++ {
                    l = (*layer[T])(next.next)
                    if l != nil {
                        break
                    } else if i > 10 {
                        runtime.Gosched()
                    }
                }
                var ll *layer[T]
                for i := 0; ; i++ {
                    ll = (*layer[T])(l.next)
                    if ll != nil {
                        break
                    } else if i > 10 {
                        runtime.Gosched()
                    }
                }
                if t, ok := q.getValue(cid&q.layerMask, ll); ok {
                    return t, true
                }
            }
        }
        runtime.Gosched()
    }
}

func (q *lkArrQueue[T]) getValue(idx uint64, next *layer[T]) (T, bool) {
    for i := 0; ; i++ {
        v := next.arr[idx]
        if !v.set {
            if i > 10 {
                runtime.Gosched()
            }
            continue
        }
        atomic.AddUint32(&next.take, 1)
        return v.v, true
    }
}

func (q *lkArrQueue[T]) fixHead(forced bool, oldHead, newHead unsafe.Pointer) {
    if forced {
        for {
            if atomic.CompareAndSwapPointer(&q.head, oldHead, newHead) {
                break
            } else {
                if atomic.LoadPointer(&q.head) != oldHead {
                    break
                }
                runtime.Gosched()
            }
        }
    } else {
        atomic.CompareAndSwapPointer(&q.head, oldHead, newHead)
    }
}

func (q *lkArrQueue[T]) Size() int {
    pid := atomic.LoadUint64(&q.produceId)
    cid := atomic.LoadUint64(&q.consumeId)
    if pid >= cid {
        return int(pid - cid)
    } else {
        //id溢出,兼容一下
        return int(pid + (1 << 63) - cid)
    }
}

func newLayer[T any](id uint64, size int) *layer[T] {
    return &layer[T]{
        id: id,
        arr: make([]struct {
            set bool
            v   T
        }, size),
    }
}

func newLinkArrQueue[T any](layerSize int) Queue[T] {
    if layerSize <= 3 {
        layerSize = 4
    }
    layerSize |= layerSize >> 1
    layerSize |= layerSize >> 2
    layerSize |= layerSize >> 4
    layerSize |= layerSize >> 8
    layerSize |= layerSize >> 16
    layerSize += 1
    l := newLayer[T](0, 0)
    id := uint64(layerSize) - 1
    l.take = uint32(layerSize)
    return &lkArrQueue[T]{
        consumeId: id,
        produceId: id,
        layerMask: uint64(layerSize - 1),
        head:      unsafe.Pointer(l),
        tail:      unsafe.Pointer(l),
    }
}

func WithTypeArrayLink[T any](layerSize int) Opt[T] {
    return func(opt *opt[T]) {
        opt.Type = func() Queue[T] { return newLinkArrQueue[T](layerSize) }
    }
}
