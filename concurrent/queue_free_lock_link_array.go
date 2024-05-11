package concurrent

import (
    "math"
    "runtime"
    "sync"
    "sync/atomic"
    "unsafe"
)

//链表+数组实现队列,无锁,待解决:id溢出问题

const layerSize = 64

type ceil struct {
    p    unsafe.Pointer
    _    [cpuCacheKillerPaddingLength]byte
    take uint32
}

type layer struct {
    lid       uint64
    reclaim   uint32
    reclaimId []uint64

    arr  [layerSize]ceil
    _    [cpuCacheKillerPaddingLength]byte
    next unsafe.Pointer
    take uint32
}

func (l *layer) id() uint64 {
    return atomic.LoadUint64(&l.lid)
}

type lkArrQueue[T any] struct {
    zero      T
    layerMask uint64 //层掩码

    produceId uint64
    _         [cpuCacheKillerPaddingLength]byte
    consumeId uint64
    _         [cpuCacheKillerPaddingLength]byte

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
    //if lIdx <= 4 {
    //更新tail
    q.newTail(layerId)
    //}
    //获取tail
    for i := 0; ; i++ {
        tail := (*layer)(atomic.LoadPointer(&q.tail))
        if tail.id() == layerId {
            q.SetV(tail, lIdx, &v)
            return
        }
        if tail.id() > layerId {
            //tail已经被后面的层替换,从head开始执行插入
            q.enqueueFromHead(layerId, lIdx, &v)
            return
        }
        if i > 10 {
            runtime.Gosched()
        }
    }
}

func (q *lkArrQueue[T]) Dequeue() (T, bool) {
    for j := 0; ; j++ {
        pid := atomic.LoadUint64(&q.produceId)
        cid := atomic.LoadUint64(&q.consumeId) + 1
        //队列为空
        if pid < cid {
            return q.zero, false
        }
        l := (*layer)(atomic.LoadPointer(&q.head))
        //说明下一层还没插入
        if l == nil {
            if j > 10 {
                runtime.Gosched()
            }
            continue
        }
        layerId := cid / (q.layerMask + 1)
        lid := l.id()
        //当前层
        if lid == layerId {
            if atomic.CompareAndSwapUint64(&q.consumeId, cid-1, cid) {
                return q.dequeueLayer(cid&q.layerMask, l), true
            }
            if j > 10 {
                runtime.Gosched()
            }
            continue
        } else
        if lid < layerId {
            if atomic.CompareAndSwapUint64(&q.consumeId, cid-1, cid) {
                if lid+4 < layerId {
                    for {
                        id := (*layer)(atomic.LoadPointer(&q.head)).id()
                        if id+4 >= layerId {
                            break
                        }
                        q.updateHead()
                        runtime.Gosched()
                    }
                }
                return q.dequeueDepth(layerId, cid&q.layerMask), true
            }
        }
        if j > 10 {
            runtime.Gosched()
        }
    }
}

//从head中找到当前层,并取出值
func (q *lkArrQueue[T]) dequeueDepth(layerId, idx uint64) T {
    defer q.updateHead()
    l := (*layer)(atomic.LoadPointer(&q.head))
    for {
        if l.id() == layerId {
            return q.dequeueLayer(idx, l)
        } else {
            for i := 0; ; i++ {
                next1 := (*layer)(atomic.LoadPointer(&l.next))
                if next1 == nil {
                    if l.id() == 0 && atomic.LoadUint32(&l.take) == 0 {
                        l = (*layer)(atomic.LoadPointer(&q.head))
                        if l.id() == 0 && atomic.LoadUint32(&l.take) == 0 {
                            panic("读取了回收的layer")
                        }
                        continue
                    }
                    if i > 10 {
                        runtime.Gosched()
                    }
                } else {
                    l = next1
                    break
                }
            }
            id := l.id()
            if id > layerId {
                l = (*layer)(atomic.LoadPointer(&q.head))
                if l.id() > layerId {
                    panic("layer.id() > layerId")
                }
            }
        }
    }
}

func (q *lkArrQueue[T]) updateHead() {
    for {
        hp := atomic.LoadPointer(&q.head)
        head := (*layer)(hp)
        if atomic.LoadPointer(&head.next) == nil {
            break
        }
        take := atomic.LoadUint32(&head.take)
        if take != uint32(q.layerMask+1) {
            break
        }
        if !atomic.CompareAndSwapUint32(&head.take, take, math.MaxUint64-10000) {
            break
        }
        if atomic.CompareAndSwapPointer(&q.head, hp, head.next) {
            reclaimLayer(head)
        } else {
            panic("替换head失败")
        }
    }
}

func (q *lkArrQueue[T]) newTail(newLayerId uint64) {
    for i := 0; ; i++ {
        tp := atomic.LoadPointer(&q.tail)
        tail := (*layer)(tp)
        if tail.id() >= newLayerId {
            return
        }
        if tail.id()+1 < newLayerId {
            if i > 10 {
                runtime.Gosched()
            }
            continue
        }
        //这里已经确定 tail 的 lid 为 newLayerId-1
        l := newLayer(newLayerId)
        nlp := unsafe.Pointer(l)
        if atomic.CompareAndSwapPointer(&tail.next, nil, nlp) {
            atomic.CompareAndSwapPointer(&q.tail, tp, nlp)
            break
        } else {
            reclaimLayer(l)
        }
    }
}

//从head开始找到当前的layer,并插入
func (q *lkArrQueue[T]) enqueueFromHead(layerId uint64, lIdx uint64, v *T) {
    head := (*layer)(atomic.LoadPointer(&q.head))
    for ; ; {
        if head.id() == layerId {
            q.SetV(head, lIdx, v)
            return
        }
        head = (*layer)(atomic.LoadPointer(&head.next))
    }
}

func (q *lkArrQueue[T]) dequeueLayer(idx uint64, l *layer) T {
    for i := 0; ; i++ {
        if atomic.LoadUint32(&l.arr[idx].take) != 0 {
            panic("take!=0")
        }
        v := q.GetV(l, idx)
        if v == nil {
            if i > 10 {
                runtime.Gosched()
            }
            continue
        }
        atomic.AddUint32(&l.take, 1)
        return *v
    }
}

func (q *lkArrQueue[T]) GetV(l *layer, idx uint64) *T {
    t := (*T)(atomic.LoadPointer(&l.arr[idx].p))
    if t != nil {
        if atomic.AddUint32(&l.arr[idx].take, 1) > 1 {
            panic("take>1")
        }
    }
    return t
}

func (q *lkArrQueue[T]) SetV(l *layer, idx uint64, v *T) {
    atomic.StorePointer(&l.arr[idx].p, unsafe.Pointer(v))
}

func (q *lkArrQueue[T]) Size() int {
    pid := atomic.LoadUint64(&q.produceId)
    cid := atomic.LoadUint64(&q.consumeId)
    return int(pid - cid)
}

var layerPool = sync.Pool{
    New: func() any { return &layer{} },
}

func newLayer(id uint64) *layer {
    l := layerPool.Get().(*layer)
    atomic.StoreUint64(&l.lid, id)
    for i := 0; i < layerSize; i++ {
        if atomic.LoadPointer(&l.arr[i].p) != nil {
            panic("layer.arr[i].p != nil")
        }
        if atomic.LoadUint32(&l.arr[i].take) != 0 {
            panic("layer.arr[i].take != 0")
        }
    }
    return l
    //return &layer{
    //    lid: id,
    //}
}
func reclaimLayer(l *layer) {
    for i := 0; i < layerSize; i++ {
        atomic.StorePointer(&l.arr[i].p, nil)
        atomic.StoreUint32(&l.arr[i].take, 0)
    }

    atomic.AddUint32(&l.reclaim, 1)
    l.reclaimId = append(l.reclaimId, l.lid)

    atomic.StoreUint32(&l.take, 0)
    atomic.StorePointer(&l.next, nil)
    layerPool.Put(l)
}

func newLinkArrQueue[T any]() Queue[T] {
    l := newLayer(0)
    id := uint64(layerSize - 1)
    l.take = uint32(layerSize)
    return &lkArrQueue[T]{
        consumeId: id,
        produceId: id,
        layerMask: uint64(layerSize - 1),
        head:      unsafe.Pointer(l),
        tail:      unsafe.Pointer(l),
    }
}

func WithTypeArrayLink[T any]() Opt[T] {
    return func(opt *opt[T]) {
        opt.Type = func() Queue[T] { return newLinkArrQueue[T]() }
    }
}
