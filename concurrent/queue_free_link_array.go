package concurrent

import (
    "runtime"
    "sync/atomic"
    "unsafe"
)

//链表+数组实现队列,无锁

type layer struct {
    id   uint64
    arr  []any
    next unsafe.Pointer
}

type lkArrQueue[T any] struct {
    //id初始值为 layerSize - 2, 这样初始状态队列获取到的第一个layerId为1,索引为0
    produceId uint64
    _         [cpuCacheKillerPaddingLength]byte
    consumeId uint64
    _         [cpuCacheKillerPaddingLength]byte

    head      unsafe.Pointer
    _         [cpuCacheKillerPaddingLength]byte
    tail      unsafe.Pointer
    _         [cpuCacheKillerPaddingLength]byte
    layerSize int    //层大小,必须是2的幂
    layerMask uint64 //层掩码
}

func (q *lkArrQueue[T]) Enqueue(v T) {
    //获取pid
    pid := atomic.AddUint64(&q.produceId, 1)
    //获取层中的索引
    lIdx := pid & q.layerMask
    //根据pid获取layerId
    layerId := pid - lIdx
    //新建layer
    if lIdx == 0 {
        l := newLayer(layerId, q.layerSize)
        lp := unsafe.Pointer(l)
        for i := 0; ; i++ {
            tail := (*layer)(atomic.LoadPointer(&q.tail))
            if tail.id != layerId-1 {
                if i > 10 {
                    runtime.Gosched()
                }
                continue
            }
            tail.next = lp
            if atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), lp) {
                break
            }
        }
        return
    } else {
        //获取tail
        tail := (*layer)(atomic.LoadPointer(&q.tail))
        for i := 0; ; i++ {
            if tail.id != layerId {
                if i > 10 {
                    runtime.Gosched()
                }
                continue
            }
            tail.arr[lIdx] = v
        }
    }
}

func (q *lkArrQueue[T]) Dequeue() (T, bool) {
    //todo 需要判断当前队列是否为空
    pid := atomic.AddUint64(&q.produceId, 1)
    lIdx := pid & q.layerMask
    layerId := pid - lIdx
    for i := 0; ; i++ {
        head := (*layer)(atomic.LoadPointer(&q.head))
        if head.id != layerId {
            if i > 10 {
                runtime.Gosched()
            }
            continue
        }
        return head.arr[lIdx].(T), true
    }
}

func newLayer(id uint64, size int) *layer {
    return &layer{
        id:  id,
        arr: make([]any, size),
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
