package concurrent

import (
    "runtime"
    "sync"
    "sync/atomic"
    "unsafe"
)

//链表+数组实现队列,无锁,待解决:id溢出问题

const layerSize = 64

type ceil struct {
    set bool
    p   any
}

const (
    changeHead uint32 = 1 << iota
    changeNext
)

type layer struct {
    status uint32

    waitRead sync.WaitGroup
    waitNext sync.WaitGroup

    lid       uint64
    reclaim   uint32
    reclaimId []uint64

    arr  [layerSize]ceil
    _    [cpuCacheKillerPaddingLength]byte
    next unsafe.Pointer
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
    for i := 0; ; i++ {
        p := unsafe.Pointer(q.tail)
        tail := (*layer)(atomic.LoadPointer(&p))
        tid := tail.id()
        if tid == layerId {
            q.SetV(tail, lIdx, v)
            return
        }
        if tid > layerId {
            //tail已经被后面的层替换,从head开始执行插入
            q.enqueueFromHead(layerId, lIdx, v)
            return
        }
        q.newTail(layerId)
        tail.waitNext.Wait()
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
        hp := unsafe.Pointer(q.head)
        l := (*layer)(atomic.LoadPointer(&hp))
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
    hp := unsafe.Pointer(q.head)
    l := (*layer)(atomic.LoadPointer(&hp))
    if l.id()+4 < layerId {
        q.updateHead()
    }
    for {
        if l.id() == layerId {
            return q.dequeueLayer(idx, l)
        } else {
            for i := 0; ; i++ {
                p := unsafe.Pointer(l.next)
                next1 := (*layer)(atomic.LoadPointer(&p))
                if next1 == nil {
                    l.waitNext.Wait()
                } else {
                    l = next1
                    break
                }
            }
            id := l.id()
            if id > layerId {
                l = (*layer)(atomic.LoadPointer(&hp))
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
        if atomic.LoadPointer(&head.next) != nil {
            stat := atomic.LoadUint32(&head.status)
            if stat&(changeHead) != 0 {
                return
            }
            if atomic.LoadPointer(&head.next) == nil {
                q.unsetLayerBit(head, changeHead)
                return
            }
            if atomic.CompareAndSwapUint32(&head.status, stat, stat|changeHead) {
                head.waitRead.Wait()
                if atomic.CompareAndSwapPointer(&q.head, hp, head.next) {
                    q.unsetLayerBit(head, changeHead)
                    reclaimLayer(head)
                } else {
                    panic("替换head失败")
                }
            } else {
                continue
            }
        } else {
            return
        }
    }
}

func (q *lkArrQueue[T]) newTail(newLayerId uint64) {
    for i := 0; ; i++ {
        tp := atomic.LoadPointer(&q.tail)
        tail := (*layer)(tp)
        tid := tail.id()
        if tid >= newLayerId {
            return
        }
        if tid+1 < newLayerId {
            tail.waitNext.Wait()
            continue
        }
        //这里已经确定 tail 的 lid 为 newLayerId-1
        stat := atomic.LoadUint32(&tail.status)
        if stat&(changeNext) != 0 {
            return
        }
        if !atomic.CompareAndSwapUint32(&tail.status, stat, stat|changeNext) {
            continue
        }
        if atomic.LoadPointer(&tail.next) != nil {
            q.unsetLayerBit(tail, changeNext)
            return
        }
        l := newLayer(newLayerId)
        nlp := unsafe.Pointer(l)
        if atomic.CompareAndSwapPointer(&tail.next, nil, nlp) {
            atomic.CompareAndSwapPointer(&q.tail, tp, nlp)
            tail.waitNext.Done()
            q.unsetLayerBit(tail, changeNext)
            break
        } else {
            panic("替换next失败")
        }
    }
}

func (q *lkArrQueue[T]) unsetLayerBit(tail *layer, bit uint32) {
    for {
        stat := atomic.LoadUint32(&tail.status)
        if atomic.CompareAndSwapUint32(&tail.status, stat, stat&^bit) {
            break
        }
    }
}

//从head开始找到当前的layer,并插入
func (q *lkArrQueue[T]) enqueueFromHead(layerId uint64, lIdx uint64, v T) {
    head := (*layer)(atomic.LoadPointer(&q.head))
    for ; ; {
        hid := head.id()
        if hid == layerId {
            q.SetV(head, lIdx, v)
            return
        }
        if hid > layerId {
            panic("head.id() > layerId")
        }
        head = (*layer)(atomic.LoadPointer(&head.next))
    }
}

func (q *lkArrQueue[T]) dequeueLayer(idx uint64, l *layer) T {
    return q.GetV(l, idx)
}

func (q *lkArrQueue[T]) GetV(l *layer, idx uint64) T {
    defer l.waitRead.Done()
    c := &l.arr[idx]
    for i := 1; ; i++ {
        if c.set {
            if i > 10 {
                runtime.Gosched()
            }
        } else {
            c.set = false
            c.p = nil
            break
        }
    }
    return (c.p).(T)
}

func (q *lkArrQueue[T]) SetV(l *layer, idx uint64, v T) {
    c := &l.arr[idx]
    c.p = v
    c.set = true
}

func (q *lkArrQueue[T]) Size() int {
    pid := atomic.LoadUint64(&q.produceId)
    cid := atomic.LoadUint64(&q.consumeId)
    return int(pid - cid)
}

var layerPool = sync.Pool{New: func() any { return &layer{} }}

func newLayer(id uint64) *layer {
    l := layerPool.Get().(*layer)
    atomic.StoreUint64(&l.lid, id)
    l.waitRead.Add(layerSize)
    l.waitNext.Add(1)
    return l
}
func reclaimLayer(l *layer) {
    for i := 0; i < layerSize; i++ {
        l.arr[i].p = nil
    }

    atomic.AddUint32(&l.reclaim, 1)
    l.reclaimId = append(l.reclaimId, l.lid)
    if atomic.LoadUint32(&l.status) != 0 {
        panic("layer.status != 0")
    }
    atomic.StorePointer(&l.next, nil)
    layerPool.Put(l)
}

func newLinkArrQueue[T any]() Queue[T] {
    l := newLayer(0)
    id := uint64(layerSize - 1)
    l.waitRead.Add(-layerSize)
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
