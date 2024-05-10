package concurrent

import (
    "runtime"
    "sync/atomic"
    "time"
    "unsafe"
)

//链表+数组实现队列,无锁

//type ceil[T any] struct {
//0,未写入,1,已写入,2,已取出
//status int32
//v      T
//}

type layer[T any] struct {
    id   uint64
    arr  []unsafe.Pointer
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
    if lIdx <= 4 {
        //更新tail
        q.newTail(layerId)
    }
    //获取tail
    for i := 0; ; i++ {
        tail := (*layer[T])(atomic.LoadPointer(&q.tail))
        if tail.id == layerId {
            atomic.StorePointer(&tail.arr[lIdx], unsafe.Pointer(&v))
            return
        }
        if tail.id > layerId {
            //tail已经被后面的层替换,从head开始执行插入
            q.enqueueFromHead(layerId, lIdx, &v)
        } else
        //等待tail更新
        {
            if i > 100 {
                time.Sleep(1 * time.Millisecond)
            } else if i > 10 {
                runtime.Gosched()
            }
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
        hp := atomic.LoadPointer(&q.head)
        np := (*layer[T])(hp).next
        next := (*layer[T])(np)
        //说明下一层还没插入
        if next == nil {
            continue
        }
        layerId := cid / (q.layerMask + 1)
        //当前层
        if next.id == layerId {
            if atomic.CompareAndSwapUint64(&q.consumeId, cid-1, cid) {
                return q.dequeueLayer(cid&q.layerMask, next), true
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
                t := q.dequeueLayer(cid&q.layerMask, l)
                q.replaceHead(hp, np)
                return t, true
            }
            continue
        } else
        if j > 100 {
            time.Sleep(1 * time.Millisecond)
        } else if j > 10 {
            runtime.Gosched()
        }
    }
}

func (q *lkArrQueue[T]) newTail(newLayerId uint64) {
    for i := 0; ; i++ {
        tp := atomic.LoadPointer(&q.tail)
        tail := (*layer[T])(tp)
        if tail.id >= newLayerId {
            return
        }
        if tail.id < newLayerId-1 {
            if i > 100 {
                time.Sleep(1 * time.Millisecond)
            } else if i > 10 {
                runtime.Gosched()
            }
            continue
        }
        //这里已经确定 tail 的 id 为 newLayerId-1
        nlp := unsafe.Pointer(newLayer[T](newLayerId, q.layerMask+1))
        if atomic.CompareAndSwapPointer(&tail.next, nil, nlp) {
            if !atomic.CompareAndSwapPointer(&q.tail, tp, nlp) {
                panic("替换tail失败")
            }
            break
        }
    }
}

//从head开始找到当前的layer,并插入
func (q *lkArrQueue[T]) enqueueFromHead(layerId uint64, lIdx uint64, v *T) {
    head := (*layer[T])(atomic.LoadPointer(&q.head))
    for i := 0; ; {
        if head.id == layerId {
            atomic.StorePointer(&head.arr[lIdx], unsafe.Pointer(v))
            return
        }
        head1 := (*layer[T])(atomic.LoadPointer(&head.next))
        if head1 == nil {
            i++
            if i > 2000 {
                panic("head")
            } else if i > 1000 {
                time.Sleep(1 * time.Millisecond)
            } else if i > 10 {
                runtime.Gosched()
            }
            head = (*layer[T])(atomic.LoadPointer(&q.head))
            if head.id > layerId {
                panic("head.id 太大")
                //return
            }
        } else {
            head = head1
        }
    }
}

func (q *lkArrQueue[T]) dequeueLayer(idx uint64, next *layer[T]) T {
    for i := 0; ; i++ {
        v := (*T)(atomic.LoadPointer(&next.arr[idx]))
        if v == nil {
            if i > 100 {
                time.Sleep(1 * time.Millisecond)
            } else if i > 10 {
                runtime.Gosched()
            }
            continue
        }
        atomic.AddUint32(&next.take, 1)
        return *v
    }
}

func (q *lkArrQueue[T]) replaceHead(oldHead, newHead unsafe.Pointer) {
    if atomic.LoadUint32(&(*layer[T])(newHead).take) == uint32(q.layerMask+1) {
        for i := 0; ; i++ {
            if atomic.LoadPointer(&q.head) != oldHead ||
                atomic.CompareAndSwapPointer(&q.head, oldHead, newHead) {
                break
            }
            if i > 100 {
                time.Sleep(1 * time.Millisecond)
            } else if i > 10 {
                runtime.Gosched()
            }
        }
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

func newLayer[T any](id, size uint64) *layer[T] {
    return &layer[T]{
        id:  id,
        arr: make([]unsafe.Pointer, size),
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
