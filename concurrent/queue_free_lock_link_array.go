package concurrent

import (
    "fmt"
    "runtime"
    "sync"
    "sync/atomic"
    "unsafe"
)

//链表+数组实现队列,无锁,待解决:id溢出问题

const layerSize = 2 << 6

const (
    layerStatusChangeHead uint32 = 1 << (iota + 16)
    layerStatusChangeNext
)

var waitReclaimLayer = make(chan *layer, layerSize)

type layer struct {
    status uint32
    _      [cpuCacheKillerPaddingLength]byte

    waitRead sync.WaitGroup
    //waitNext sync.WaitGroup

    lid uint64
    //reclaim   uint32
    //reclaimId []uint64

    //set  uint32
    num  int32
    arr  [layerSize]unsafe.Pointer
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
    //_        [cpuCacheKillerPaddingLength]byte
    //waitFull unsafe.Pointer
}

func (q *lkArrQueue[T]) Enqueue(v T) {
    //获取pid
    pid := atomic.AddUint64(&q.produceId, 1)
    //获取层中的索引
    lIdx := pid & q.layerMask
    //根据pid获取layerId
    layerId := pid / (q.layerMask + 1)
    for i := 0; ; i++ {
        tail := (*layer)(atomic.LoadPointer(&q.tail))
        tid := tail.id()
        if tid == layerId {
            q.SetV(tail, lIdx, &v)
            return
        }
        //t1 := (*layer)(atomic.LoadPointer(&q.waitFull))
        //if t1.id() == layerId {
        //    q.SetV(t1, lIdx, &v)
        //    return
        //}
        //if t1.id() > layerId {
        //    q.enqueueFromWait(layerId, lIdx, &v)
        //    return
        //}
        if tid > layerId {
            //tail已经被后面的层替换, 从head开始执行插入
            q.enqueueFromHead(layerId, lIdx, &v)
            return
        }
        if tid < layerId {
            q.newTail(layerId)
        }
        if i > 100 {
            runtime.Gosched()
        }
        //atomic.AddUint32(&tail.status, 1)
        //tail.waitNext.Wait()
        //atomic.AddUint32(&tail.status, ^uint32(0))
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
        //if lid < layerId {
        //    if atomic.CompareAndSwapUint64(&q.consumeId, cid-1, cid) {
        //        return q.dequeueDepth(layerId, cid&q.layerMask), true
        //    }
        //}
        if j > 10 {
            runtime.Gosched()
        }
        q.updateHead(layerId)
    }
}

//从head中找到当前层,并取出值
func (q *lkArrQueue[T]) dequeueDepth(layerId, idx uint64) T {
    l := (*layer)(atomic.LoadPointer(&q.head))
    if l.id()+1 < layerId && atomic.LoadUint32(&l.status)&layerStatusChangeHead != 0 {
        //l.waitNext.Wait()
        q.updateHead(layerId)
    }
    return q.dequeueDepth1(layerId, idx)
}

func (q *lkArrQueue[T]) dequeueDepth1(layerId uint64, idx uint64) T {
    l := (*layer)(atomic.LoadPointer(&q.head))
    for {
        if l.id() == layerId {
            return q.dequeueLayer(idx, l)
        } else {
            next1 := (*layer)(atomic.LoadPointer(&l.next))
            if next1 == nil {
                atomic.AddUint32(&l.status, 1)
                //l.waitNext.Wait()
                next1 = (*layer)(atomic.LoadPointer(&l.next))
                atomic.AddUint32(&l.status, ^uint32(0))
                if next1 == nil {
                    l = (*layer)(atomic.LoadPointer(&q.head))
                    continue
                }
            } else {
                l = next1
            }
            if l.id() > layerId {
                l = (*layer)(atomic.LoadPointer(&q.head))
                if l.id() > layerId {
                    panic(fmt.Sprintf("layer.id(%d) > layerId(%d)", l.id(), layerId))
                }
            }
        }
    }
}

func (q *lkArrQueue[T]) updateHead(layerId uint64) {
    for {
        hp := atomic.LoadPointer(&q.head)
        head := (*layer)(hp)
        if head.id() < layerId && atomic.LoadPointer(&head.next) != nil {
            stat := atomic.LoadUint32(&head.status)
            if stat&(layerStatusChangeHead) != 0 || !atomic.CompareAndSwapUint32(&head.status, stat, stat|layerStatusChangeHead) {
                return
            }
            if atomic.LoadPointer(&head.next) == nil {
                q.unsetLayerBit(head, layerStatusChangeHead)
                return
            }

            head.waitRead.Wait()
            if atomic.CompareAndSwapPointer(&q.head, hp, head.next) {
                q.unsetLayerBit(head, layerStatusChangeHead)
                reclaimLayer(head)
            } else {
                q.unsetLayerBit(head, layerStatusChangeHead)
                return
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
            atomic.AddUint32(&tail.status, 1)
            //tail.waitNext.Wait()
            atomic.AddUint32(&tail.status, ^uint32(0))
            continue
        }
        //这里已经确定 tail 的 lid 为 newLayerId-1
        stat := atomic.LoadUint32(&tail.status)
        if stat&(layerStatusChangeNext) != 0 {
            return
        }
        if !atomic.CompareAndSwapUint32(&tail.status, stat, stat|layerStatusChangeNext) {
            continue
        }
        if atomic.LoadPointer(&tail.next) != nil {
            q.unsetLayerBit(tail, layerStatusChangeNext)
            return
        }
        l := newLayer(newLayerId)
        nlp := unsafe.Pointer(l)
        if atomic.CompareAndSwapPointer(&tail.next, nil, nlp) {
            atomic.CompareAndSwapPointer(&q.tail, tp, nlp)
            //tail.waitNext.Done()
            q.unsetLayerBit(tail, layerStatusChangeNext)
            break
        } else {
            panic("替换next失败")
        }
    }
}

func (q *lkArrQueue[T]) setLayerBit(tail *layer, bit uint32) {
    for {
        stat := atomic.LoadUint32(&tail.status)
        if atomic.CompareAndSwapUint32(&tail.status, stat, stat|bit) {
            break
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
func (q *lkArrQueue[T]) enqueueFromHead(layerId uint64, lIdx uint64, v *T) {
    head := (*layer)(atomic.LoadPointer(&q.head))
    for ; ; {
        hid := head.id()
        if hid == layerId {
            q.SetV(head, lIdx, v)
            return
        }
        if hid > layerId {
            head = (*layer)(atomic.LoadPointer(&q.head))
            if head.id() > layerId {
                panic("head.id() > layerId")
            }
            continue
        }
        head = (*layer)(atomic.LoadPointer(&head.next))
        if head == nil {
            head = (*layer)(atomic.LoadPointer(&q.head))
        }
    }
}

//func (q *lkArrQueue[T]) enqueueFromWait(layerId uint64, lIdx uint64, v *T) {
//    l := (*layer)(atomic.LoadPointer(&q.waitFull))
//    for {
//        hid := l.id()
//        if hid == layerId {
//            q.SetV(l, lIdx, v)
//            return
//        }
//        if hid > layerId {
//            q.enqueueFromHead(layerId, lIdx, v)
//            return
//        }
//        l = (*layer)(atomic.LoadPointer(&l.next))
//        if l == nil {
//            l = (*layer)(atomic.LoadPointer(&q.waitFull))
//        }
//    }
//}
//
//func (q *lkArrQueue[T]) updateWait(l *layer) {
//    for {
//        next := l.next
//        if next == nil || atomic.LoadInt32(&l.num) != layerSize || !atomic.CompareAndSwapPointer(&q.waitFull, unsafe.Pointer(l), next) {
//            break
//        }
//        l = (*layer)(atomic.LoadPointer(&q.waitFull))
//    }
//}

func (q *lkArrQueue[T]) dequeueLayer(idx uint64, l *layer) T {
    return q.GetV(l, idx)
}

func (q *lkArrQueue[T]) GetV(l *layer, idx uint64) T {
    for i := 1; ; i++ {
        p := atomic.LoadPointer(&l.arr[idx])
        if p != nil {
            t := (*T)(p)
            atomic.StorePointer(&l.arr[idx], nil)
            //atomic.AddInt32(&l.num, -1)
            l.waitRead.Done()
            return *t
        } else if i > 10 {
            runtime.Gosched()
        }
    }
}

func (q *lkArrQueue[T]) SetV(l *layer, idx uint64, v *T) {
    atomic.StorePointer(&l.arr[idx], unsafe.Pointer(v))
    atomic.AddInt32(&l.num, 1)
    //q.updateWait(l)
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
    //l.waitNext.Add(1)
    return l
}
func reclaimLayer(l *layer) {
    if atomic.LoadUint32(&l.status) != 0 {
        runtime.Gosched()
        if atomic.LoadUint32(&l.status) != 0 {
            select {
            case waitReclaimLayer <- l:
            default:
            }
            return
        }
    }
out:
    for {
        select {
        case l1 := <-waitReclaimLayer:
            if atomic.LoadUint32(&l1.status) == 0 {
                atomic.StoreUint64(&l1.lid, 0)
                atomic.StoreInt32(&l1.num, 0)
                atomic.StorePointer(&l1.next, nil)
                layerPool.Put(l1)
            } else {
                println("l1.status != 0,discard")
                break out
            }
        default:
            break out
        }
    }
    atomic.StoreUint64(&l.lid, 0)
    atomic.StoreInt32(&l.num, 0)
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
        //waitFull:  unsafe.Pointer(l),
    }
}

func WithTypeArrayLink[T any]() Opt[T] {
    return func(opt *opt[T]) {
        opt.Type = func() Queue[T] { return newLinkArrQueue[T]() }
    }
}
