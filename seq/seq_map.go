package seq

import (
    "sync"
    "sync/atomic"
)

//======转换,添加或修改内部元素========

// MapParallel 每个元素转换为any,使用 Sync() 保证消费不竞争
// order 是否保持顺序,1尽量保持顺序(可能消费竞争),大于1强制保持顺序(约等于加锁)
// order 第二个参数,并发数
func (t Seq[T]) MapParallel(f func(T) any, order ...int) Seq[any] {
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
        return func(c func(any)) {
            var currentIndex int32 = 1
            var id int32
            t(func(t T) {
                var id = atomic.AddInt32(&id, 1)
                p.Add(func() {
                    a := f(t)
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
                    c(a)
                })
            })
            p.Wait()
        }
    } else {
        return t.Parallel(sl).Map(f)
    }
}

// Map 每个元素转换为any
func (t Seq[T]) Map(f func(T) any) Seq[any] {
    return func(c func(any)) { t(func(t T) { c(f(t)) }) }
}

// MapFlat 每个元素转换为Seq,并扁平化
func (t Seq[T]) MapFlat(f func(T) Seq[any]) Seq[any] {
    return func(c func(any)) { t(func(t T) { f(t).ForEach(c) }) }
}

// MapString 每个元素转换为字符串
func (t Seq[T]) MapString(f func(T) string) Seq[string] {
    return func(c func(string)) { t(func(t T) { c(f(t)) }) }
}

// MapInt 每个元素转换为int
func (t Seq[T]) MapInt(f func(T) int) Seq[int] {
    return func(c func(int)) { t(func(t T) { c(f(t)) }) }
}

// MapFloat32 每个元素转换为float32
func (t Seq[T]) MapFloat32(f func(T) float32) Seq[float32] {
    return func(c func(float32)) { t(func(t T) { c(f(t)) }) }
}

// MapFloat64 每个元素转换为float64
func (t Seq[T]) MapFloat64(f func(T) float64) Seq[float64] {
    return func(c func(float64)) { t(func(t T) { c(f(t)) }) }
}

// MapSliceN 每n个元素合并为[]T,由于golang泛型问题,不能使用Seq[[]T],使用 CastAny 转换为Seq[[]T]
func (t Seq[T]) MapSliceN(n int) Seq[any] {
    return t.MapSliceBy(func(t T, ts []T) bool { return len(ts) == n })
}

// MapSliceAnyN 每n个元素合并为[]T,由于golang泛型问题,不能使用[]Seq[T]
func (t Seq[T]) MapSliceAnyN(n int) Seq[[]any] {
    return t.MapSliceAnyBy(func(t T, ts []T) bool { return len(ts) == n })
}

// MapSliceBy 自定义元素合并为[]T,由于golang泛型问题,不能使用[]Seq[T] CastAny 转换为Seq[[]T]
func (t Seq[T]) MapSliceBy(f func(T, []T) bool) Seq[any] {
    return func(c func(any)) {
        var ts []T
        t(func(t T) {
            ts = append(ts, t)
            if f(t, ts) {
                c(ts)
                ts = nil
            }
        })
        if len(ts) > 0 {
            c(ts)
        }
    }
}

// MapSliceAnyBy 自定义元素合并为[]T,由于golang泛型问题,不能使用[]Seq[T]
func (t Seq[T]) MapSliceAnyBy(f func(T, []T) bool) Seq[[]any] {
    return func(c func([]any)) {
        var ts []T
        t(func(t T) {
            ts = append(ts, t)
            if f(t, ts) {
                c(FromSlice(ts).Map(AnyT[T]).ToSlice())
                ts = nil
            }
        })
        if len(ts) > 0 {
            c(FromSlice(ts).Map(AnyT[T]).ToSlice())
        }
    }
}

// Join 合并多个Seq
func (t Seq[T]) Join(seqs ...Seq[T]) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        for _, seq := range seqs {
            seq(func(t T) { c(t) })
        }
    }
}

// JoinF 合并Seq
func (t Seq[T]) JoinF(seq Seq[any], cast func(any) T) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        seq(func(t any) { c(cast(t)) })
    }
}

// Add 添加单个元素
func (t Seq[T]) Add(ts ...T) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        for _, e := range ts {
            c(e)
        }
    }
}

// AddF 添加单个元素
func (t Seq[T]) AddF(cast func(any) T, ts ...any) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        for _, e := range ts {
            c(cast(e))
        }
    }
}
