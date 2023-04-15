package seq

import (
    "context"
    "golang.org/x/sync/semaphore"
    "sync"
    "sync/atomic"
)

//======转换========

// MapKParallel 每个元素转换为any,使用 Sync() 保证消费不竞争
// order 是否保持顺序,大于0保持顺序
// order 第二个参数,并发数
func (t BiSeq[K, V]) MapKParallel(f func(k K, v V) any, order ...int) BiSeq[any, V] {
    o := false
    sl := 0
    if len(order) > 0 {
        o = order[0] > 0
    }
    if len(order) > 1 {
        sl = order[1]
    }
    if o {
        l := sync.NewCond(&sync.Mutex{})
        return func(c func(k any, v V)) {
            var currentIndex int32 = 1
            var currentIndex1 int32 = 1
            var id int32
            s := semaphore.NewWeighted(int64(sl))
            t.MapK(func(k K, v V) any {
                lock := sync.Mutex{}
                lock.Lock()
                var id = atomic.AddInt32(&id, 1)
                go func() {
                    if sl > 0 {
                        //限制并发时,保证携程启动顺序
                        l.L.Lock()
                        for atomic.LoadInt32(&currentIndex1) != id {
                            l.Wait()
                        }
                        atomic.AddInt32(&currentIndex1, 1)
                        l.L.Unlock()
                        l.Broadcast()
                        s.Acquire(context.Background(), 1)
                    }
                    defer lock.Unlock()
                    a := f(k, v)
                    if sl > 0 {
                        s.Release(1)
                    }
                    l.L.Lock()
                    for atomic.LoadInt32(&currentIndex) != id {
                        l.Wait()
                    }
                    c(a, v)
                    atomic.AddInt32(&currentIndex, 1)
                    l.L.Unlock()
                    l.Broadcast()
                }()
                return &lock
            }).SeqK().Cache()(func(t any) {
                lock := t.(sync.Locker)
                lock.Lock()
            })
        }
    } else {
        return t.Parallel(sl).MapK(f)
    }
}

// MapVParallel 每个元素转换为any,使用 Sync() 保证消费不竞争
// order 是否保持顺序,大于0保持顺序
// order 第二个参数,并发数
func (t BiSeq[K, V]) MapVParallel(f func(k K, v V) any, order ...int) BiSeq[K, any] {
    o := false
    sl := 0
    if len(order) > 0 {
        o = order[0] > 0
    }
    if len(order) > 1 {
        sl = order[1]
    }
    if o {
        l := sync.NewCond(&sync.Mutex{})
        return func(c func(k K, v any)) {
            var currentIndex int32 = 1
            var currentIndex1 int32 = 1
            var id int32
            s := semaphore.NewWeighted(int64(sl))
            t.MapK(func(k K, v V) any {
                lock := sync.Mutex{}
                lock.Lock()
                var id = atomic.AddInt32(&id, 1)
                go func() {
                    if sl > 0 {
                        //限制并发时,保证携程启动顺序
                        l.L.Lock()
                        for atomic.LoadInt32(&currentIndex1) != id {
                            l.Wait()
                        }
                        atomic.AddInt32(&currentIndex1, 1)
                        l.L.Unlock()
                        l.Broadcast()
                        s.Acquire(context.Background(), 1)
                    }
                    defer lock.Unlock()
                    a := f(k, v)
                    if sl > 0 {
                        s.Release(1)
                    }
                    l.L.Lock()
                    for atomic.LoadInt32(&currentIndex) != id {
                        l.Wait()
                    }
                    c(k, a)
                    atomic.AddInt32(&currentIndex, 1)
                    l.L.Unlock()
                    l.Broadcast()
                }()
                return &lock
            }).SeqK().Cache()(func(t any) {
                lock := t.(sync.Locker)
                lock.Lock()
            })
        }
    } else {
        return t.Parallel(sl).MapV(f)
    }
}

// Map 每个元素自定义转换为any,用于连续转换操作,使用 BiCastAny 进行恢复泛型
func (t BiSeq[K, V]) Map(f func(K, V) (any, any)) BiSeq[any, any] {
    return func(c func(any, any)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// MapK 每个元素自定义转换K为any,用于连续转换操作,使用 BiCastAny 进行恢复泛型
func (t BiSeq[K, V]) MapK(f func(K, V) any) BiSeq[any, V] {
    return func(c func(any, V)) { t(func(k K, v V) { c(f(k, v), v) }) }
}

// MapV 每个元素自定义转换V为any,用于连续转换操作,使用 BiCastAny 进行恢复泛型
func (t BiSeq[K, V]) MapV(f func(K, V) any) BiSeq[K, any] {
    return func(c func(K, any)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// SeqK 转换为只保留K的Seq
func (t BiSeq[K, V]) SeqK() Seq[K] {
    return func(c func(K)) { t(func(k K, v V) { c(k) }) }
}

// SeqV 转换为只保留V的Seq
func (t BiSeq[K, V]) SeqV() Seq[V] {
    return func(c func(V)) { t(func(k K, v V) { c(v) }) }
}

// SeqF 转换为Seq[any],自定义转换
func (t BiSeq[K, V]) SeqF(f func(K, V) any) Seq[any] {
    return func(c func(any)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// SeqKF 转换为只保留K的Seq,并自定义转换
func (t BiSeq[K, V]) SeqKF(f func(K, V) K) Seq[K] {
    return func(c func(K)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// SeqVF 转换为只保留V的Seq,并自定义转换
func (t BiSeq[K, V]) SeqVF(f func(K, V) V) Seq[V] {
    return func(c func(V)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// MapFlat 每个元素转换为BiSeq[any,any],并扁平化
func (t BiSeq[K, V]) MapFlat(f func(K, V) BiSeq[any, any]) BiSeq[any, any] {
    return func(c func(any, any)) {
        t(func(k K, v V) {
            s := f(k, v)
            s.ForEach(c)
        })
    }
}

// MapSliceN 每n个元素合并为[]T
func (t BiSeq[K, V]) MapSliceN(n int) Seq[[]BiTuple[K, V]] {
    return func(c func([]BiTuple[K, V])) {
        var ts []BiTuple[K, V]
        t(func(k K, v V) {
            ts = append(ts, BiTuple[K, V]{k, v})
            if len(ts) == n {
                c(ts)
                ts = nil
            }
        })
        if len(ts) > 0 {
            c(ts)
        }
    }
}

// MapSliceF 自定义元素合并为[]T
func (t BiSeq[K, V]) MapSliceF(f func(K, V, []BiTuple[K, V]) bool) Seq[[]BiTuple[K, V]] {
    return func(c func([]BiTuple[K, V])) {
        var ts []BiTuple[K, V]
        t(func(k K, v V) {
            ts = append(ts, BiTuple[K, V]{k, v})
            if f(k, v, ts) {
                c(ts)
                ts = nil
            }
        })
        if len(ts) > 0 {
            c(ts)
        }
    }
}

// StringMap 转换为Seq[string]
func (t BiSeq[K, V]) StringMap(f func(K, V) string) Seq[string] {
    return func(c func(string)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// Join 合并多个Seq
func (t BiSeq[K, V]) Join(seqs ...BiSeq[K, V]) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) { c(k, v) })
        for _, seq := range seqs {
            seq(func(k K, v V) { c(k, v) })
        }
    }
}

// JoinF 合并Seq
func (t BiSeq[K, V]) JoinF(seq BiSeq[any, any], cast func(any, any) (K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) { c(k, v) })
        seq(func(k any, v any) { c(cast(k, v)) })
    }
}

// Add 添加元素
func (t BiSeq[K, V]) Add(k K, v V) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) { c(k, v) })
        c(k, v)
    }
}

// AddF 添加元素
func (t BiSeq[K, V]) AddF(cast func(any, any) (K, V), es ...any) BiSeq[K, V] {
    if len(es)%2 != 0 {
        panic("添加的元素个数必须为偶数")
    }
    return func(c func(K, V)) {
        t(func(k K, v V) { c(k, v) })
        FromIntSeq(0, len(es), 2).ForEach(func(i int) { c(cast(es[i], es[i+1])) })
    }
}
