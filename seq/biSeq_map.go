package seq

import (
    "sync"
    "sync/atomic"
)

//======转换========

// MapKParallel 每个元素转换为any,使用 Sync() 保证消费不竞争
// order 是否保持顺序,大于0保持顺序
// order 第二个参数,并发数
func (t BiSeq[K, V]) MapKParallel(f func(k K, v V) any, order ...int) BiSeq[any, V] {
    o := 0
    sl := 0
    if len(order) > 0 {
        o = order[0]
    }
    if len(order) > 1 {
        sl = order[1]
    }
    if o > 0 {
        p := NewParallel(sl)
        l := sync.NewCond(&sync.Mutex{})
        return func(c func(any, V)) {
            var currentIndex int32 = 1
            var id int32
            t(func(k K, v V) {
                var id = atomic.AddInt32(&id, 1)
                p.Add(func() {
                    a := f(k, v)
                    l.L.Lock()
                    for atomic.LoadInt32(&currentIndex) != id {
                        l.Wait()
                    }
                    atomic.AddInt32(&currentIndex, 1)
                    defer l.Broadcast()
                    if o > 1 {
                        defer l.L.Unlock()
                    } else {
                        l.L.Unlock()
                    }
                    c(a, v)
                })
            })
            p.Wait()
        }
    } else {
        return t.Parallel(sl).MapK(f)
    }
}

// MapVParallel 每个元素转换为any,使用 Sync() 保证消费不竞争
// order 是否保持顺序,1尽量保持顺序(可能消费竞争),大于1强制保持顺序(约等于加锁)
// order 第二个参数,并发数
func (t BiSeq[K, V]) MapVParallel(f func(k K, v V) any, order ...int) BiSeq[K, any] {
    o := 0
    sl := 0
    if len(order) > 0 {
        o = order[0]
    }
    if len(order) > 1 {
        sl = order[1]
    }
    if o > 0 {
        p := NewParallel(sl)
        l := sync.NewCond(&sync.Mutex{})
        return func(c func(K, any)) {
            var currentIndex int32 = 1
            var id int32
            t(func(k K, v V) {
                var id = atomic.AddInt32(&id, 1)
                p.Add(func() {
                    a := f(k, v)
                    l.L.Lock()
                    for atomic.LoadInt32(&currentIndex) != id {
                        l.Wait()
                    }
                    atomic.AddInt32(&currentIndex, 1)
                    defer l.Broadcast()
                    if o > 1 {
                        defer l.L.Unlock()
                    } else {
                        l.L.Unlock()
                    }
                    c(k, a)
                })
            })
            p.Wait()
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

// SeqBy 转换为Seq[any],自定义转换
func (t BiSeq[K, V]) SeqBy(f func(K, V) any) Seq[any] {
    return func(c func(any)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// SeqKBy 转换为只保留K的Seq,并自定义转换
func (t BiSeq[K, V]) SeqKBy(f func(K, V) K) Seq[K] {
    return func(c func(K)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// SeqVBy 转换为只保留V的Seq,并自定义转换
func (t BiSeq[K, V]) SeqVBy(f func(K, V) V) Seq[V] {
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

// MapSliceN 每n个元素合并为[]T,由于golang泛型问题,不能使用[]BiTuple[K,V],使用 BiCastAny 进行恢复泛型[]BiTuple[K,V]
func (t BiSeq[K, V]) MapSliceN(n int) Seq[any] {
    return t.MapSliceBy(func(k K, v V, ts any) bool { return len(ts.([]BiTuple[K, V])) == n })
}

// MapSliceAnyN 每n个元素合并为[]T,由于golang泛型问题,不能使用[]BiTuple[K,V]
func (t BiSeq[K, V]) MapSliceAnyN(n int) Seq[[]BiTuple[any, any]] {
    return t.MapSliceAnyBy(func(k K, v V, ts []BiTuple[any, any]) bool { return len(ts) == n })
}

//MapSliceBy 自定义元素合并为[]T,由于golang泛型问题,不能使用[]BiTuple[K,V],使用 BiCastAny 进行恢复泛型[]BiTuple[K,V]
func (t BiSeq[K, V]) MapSliceBy(f func(K, V, any) bool) Seq[any] {
    return func(c func(any)) {
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

//MapSliceAnyBy 自定义元素合并为[]T,由于golang泛型问题,不能使用[]BiTuple[K,V]
func (t BiSeq[K, V]) MapSliceAnyBy(f func(K, V, []BiTuple[any, any]) bool) Seq[[]BiTuple[any, any]] {
    return func(c func([]BiTuple[any, any])) {
        var ts []BiTuple[any, any]
        t(func(k K, v V) {
            ts = append(ts, BiTuple[any, any]{k, v})
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

// JoinBy 合并Seq
func (t BiSeq[K, V]) JoinBy(seq BiSeq[any, any], cast func(any, any) (K, V)) BiSeq[K, V] {
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

// AddTuple 添加元素
func (t BiSeq[K, V]) AddTuple(vs ...BiTuple[K, V]) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) { c(k, v) })
        for _, v := range vs {
            c(v.K, v.V)
        }
    }
}

// AddBy 添加元素
func (t BiSeq[K, V]) AddBy(cast func(any, any) (K, V), es ...any) BiSeq[K, V] {
    if len(es)%2 != 0 {
        panic("添加的元素个数必须为偶数")
    }
    return func(c func(K, V)) {
        t(func(k K, v V) { c(k, v) })
        FromIntSeq(0, len(es), 2)(func(i int) { c(cast(es[i], es[i+1])) })
    }
}
