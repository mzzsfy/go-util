package concurrent

import (
    "runtime"
    "sync/atomic"
    "unsafe"
)

//链表+数组实现队列,无锁

type layer struct {
    id   uint64
    take uint32
    arr  []any
    next unsafe.Pointer
}

type lkArrQueue[T any] struct {
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
        l := newLayer(layerId, int(q.layerMask+1))
        l.arr[0] = v
        lp := unsafe.Pointer(l)
        for i := 0; ; i++ {
            tail := (*layer)(atomic.LoadPointer(&q.tail))
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
            tail := (*layer)(atomic.LoadPointer(&q.tail))
            if tail.id == layerId {
                tail.arr[lIdx] = v
                return
            }
            //tail已经被替换,从head开始找到当前的layer
            if tail.id > layerId {
                head := (*layer)(atomic.LoadPointer(&q.head))
                for {
                    if head.id == layerId {
                        head.arr[lIdx] = v
                        return
                    }
                    head = (*layer)(head.next)
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
            var t T
            return t, false
        }
        hp := atomic.LoadPointer(&q.head)
        next := (*layer)((*layer)(hp).next)
        if next == nil {
            var t T
            return t, false
        }
        layerId := cid / (q.layerMask + 1)
        if next.id == layerId {
            if atomic.CompareAndSwapUint64(&q.consumeId, cid-1, cid) {
                for i := 0; ; i++ {
                    idx := cid & q.layerMask
                    v := next.arr[idx]
                    if v == nil {
                        if i > 10 {
                            runtime.Gosched()
                        }
                        continue
                    }
                    atomic.AddUint32(&next.take, 1)
                    return v.(T), true
                }
            }
            continue
        }
        if next.id == layerId-1 {
            if atomic.CompareAndSwapUint64(&q.consumeId, cid-1, cid) {
                for i := 0; ; i++ {
                    var n *layer
                    for j := 0; ; j++ {
                        n = (*layer)(atomic.LoadPointer(&next.next))
                        if n != nil {
                            break
                        } else if j > 10 {
                            runtime.Gosched()
                        }
                    }
                    idx := cid & q.layerMask
                    v := n.arr[idx]
                    if v == nil {
                        if i > 10 {
                            runtime.Gosched()
                        }
                        continue
                    }
                    atomic.AddUint32(&n.take, 1)
                    //读取新的层了
                    if idx == 0 {
                        for {
                            if (*layer)(hp).take == uint32(q.layerMask+1) {
                                if atomic.CompareAndSwapPointer(&q.head, hp, unsafe.Pointer(next)) {
                                    break
                                }
                            } else {
                                runtime.Gosched()
                            }
                        }
                    }
                    return v.(T), true
                }
            }
        }
        runtime.Gosched()
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

func newLayer(id uint64, size int) *layer {
    return &layer{
        id:  id,
        arr: make([]any, size),
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
    l := newLayer(0, 0)
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
