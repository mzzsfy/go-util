package seq

import (
    "sync"
    "sync/atomic"
)

//======转换,添加或修改内部元素========

// MapParallel 每个元素转换为any
// order.1 顺序保证方式,规则如下:
// 0:不保证任务启动顺序,不保证消费顺序,会消费竞争
// 1:尽量保持顺序,优先保证并发数,异步任务完成时,会直接消费,会消费竞争,可以使用 Sync() 保证消费不竞争
// 2:异步任务与消费端解偶,在保证顺序的前提下,优先保证并发数,不会消费竞争
// 3:保持异步与消费同步,以消费为准,不消费完成不会开始下一个异步任务,不会消费竞争
//
// order.2 最大并发数,根据第一个参数决定逻辑
func (t Seq[T]) MapParallel(syncFn func(T) any, order ...int) Seq[any] {
    o := 0
    sl := 0
    if len(order) > 0 {
        o = order[0]
    }
    if len(order) > 1 {
        sl = order[1]
    }
    if o > 0 {
        return func(c func(any)) {
            var currentIndex int32 = 1
            var id int32
            var fns []*BiTuple[int32, func()]
            lock := &sync.Mutex{}
            l := sync.NewCond(lock)
            p := NewParallel(sl)
            fn := func() {
                lock.Lock()
                defer lock.Unlock()
                for {
                    loaded := false
                    idx := atomic.LoadInt32(&currentIndex)
                    for i, b := range fns {
                        if b != nil && b.K == idx {
                            b.V()
                            atomic.AddInt32(&currentIndex, 1)
                            loaded = true
                            fns[i] = nil
                            break
                        }
                    }
                    if !loaded {
                        break
                    }
                }
            }
            t(func(t T) {
                var id = atomic.AddInt32(&id, 1)
                p.Add(func() {
                    a := syncFn(t)
                    if o == 1 {
                        c(a)
                    } else if o == 2 {
                        lock.Lock()
                        defer lock.Unlock()
                        if atomic.LoadInt32(&currentIndex) != id {
                            fns = append(fns, &BiTuple[int32, func()]{id, func() { c(a) }})
                        } else {
                            c(a)
                            atomic.AddInt32(&currentIndex, 1)
                            for _, f := range fns {
                                if f != nil {
                                    DefaultParallelFunc(fn)
                                    return
                                }
                            }
                        }
                    } else {
                        l.L.Lock()
                        defer l.L.Unlock()
                        for atomic.LoadInt32(&currentIndex) != id {
                            l.Wait()
                        }
                        defer l.Broadcast()
                        c(a)
                        atomic.AddInt32(&currentIndex, 1)
                    }
                })
            })
            p.Wait()
            fn()
            fns = nil
        }
    } else {
        return t.Parallel(sl).Map(syncFn)
    }
}

func (t Seq[T]) MapParallelCustomize(asyncFn func(T, func(any))) Seq[any] {
    return func(c func(any)) {
        wg := sync.WaitGroup{}
        t(func(t T) {
            wg.Add(1)
            asyncFn(t, func(r any) {
                defer wg.Done()
                c(r)
            })
        })
        wg.Wait()
    }
}

// Map 每个元素转换为any
func (t Seq[T]) Map(f func(T) any) Seq[any] {
    return func(c func(any)) { t(func(t T) { c(f(t)) }) }
}

// MapString 每个元素转换为 string
func (t Seq[T]) MapString(f func(T) string) Seq[string] {
    return func(c func(string)) { t(func(t T) { c(f(t)) }) }
}

// MapInt 每个元素转换为 int
func (t Seq[T]) MapInt(f func(T) int) Seq[int] {
    return func(c func(int)) { t(func(t T) { c(f(t)) }) }
}

// MapInt64 每个元素转换为 int64
func (t Seq[T]) MapInt64(f func(T) int64) Seq[int64] {
    return func(c func(int64)) { t(func(t T) { c(f(t)) }) }
}

// MapFloat32 每个元素转换为 float32
func (t Seq[T]) MapFloat32(f func(T) float32) Seq[float32] {
    return func(c func(float32)) { t(func(t T) { c(f(t)) }) }
}

// MapFloat64 每个元素转换为 float64
func (t Seq[T]) MapFloat64(f func(T) float64) Seq[float64] {
    return func(c func(float64)) { t(func(t T) { c(f(t)) }) }
}

// MapBytes 每个元素转换为 []byte
func (t Seq[T]) MapBytes(f func(T) []byte) Seq[[]byte] {
    return func(c func([]byte)) { t(func(t T) { c(f(t)) }) }
}

// MapFlat 每个元素转换为Seq,并扁平化
func (t Seq[T]) MapFlat(f func(T) Seq[any]) Seq[any] {
    return func(c func(any)) { t(func(t T) { f(t)(c) }) }
}

// MapFlatInt 扁平化
func (t Seq[T]) MapFlatInt(f func(T) Seq[int]) Seq[int] {
    return func(c func(int)) { t(func(t T) { f(t)(c) }) }
}

// MapFlatInt64 扁平化
func (t Seq[T]) MapFlatInt64(f func(T) Seq[int64]) Seq[int64] {
    return func(c func(int64)) { t(func(t T) { f(t)(c) }) }
}

// MapFlatString 扁平化
func (t Seq[T]) MapFlatString(f func(T) Seq[string]) Seq[string] {
    return func(c func(string)) { t(func(t T) { f(t)(c) }) }
}

// MapFlatFloat32 扁平化
func (t Seq[T]) MapFlatFloat32(f func(T) Seq[float32]) Seq[float32] {
    return func(c func(float32)) { t(func(t T) { f(t)(c) }) }
}

// MapFlatFloat64 扁平化
func (t Seq[T]) MapFlatFloat64(f func(T) Seq[float64]) Seq[float64] {
    return func(c func(float64)) { t(func(t T) { f(t)(c) }) }
}

// MapFlatBytes 扁平化
func (t Seq[T]) MapFlatBytes(f func(T) Seq[[]byte]) Seq[[]byte] {
    return func(c func([]byte)) { t(func(t T) { f(t)(c) }) }
}

// MapSliceN 每n个元素合并为[]T,由于golang泛型问题,不能使用Seq[[]T],使用 CastAny 转换为Seq[[]T]
func (t Seq[T]) MapSliceN(n int) Seq[any] {
    return t.MapSliceBy(func(t T, ts []T) bool { return len(ts) == n })
}

// MapSliceBy 自定义元素合并为[]T,由于golang泛型问题,不能返回[]Seq[T],使用 CastAny 转换为Seq[[]T]
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

// Add 直接添加元素
func (t Seq[T]) Add(ts ...T) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        for _, e := range ts {
            c(e)
        }
    }
}

// AddF 直接添加需要转换的元素
func (t Seq[T]) AddF(cast func(any) T, ts ...any) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        for _, e := range ts {
            c(cast(e))
        }
    }
}

// AddIf 满足条件才添加元素
func (t Seq[T]) AddIf(condition bool, ts ...T) Seq[T] {
    if !condition {
        return t
    }
    return t.Add(ts...)
}

// AddIfF 满足条件才添加元素
func (t Seq[T]) AddIfF(condition func(Seq[T]) bool, ts ...T) Seq[T] {
    if !condition(t) {
        return t
    }
    return t.Add(ts...)
}

// AddFIf 满足条件才添加需要转换的元素
func (t Seq[T]) AddFIf(condition bool, cast func(any) T, ts ...any) Seq[T] {
    if !condition {
        return t
    }
    return t.AddF(cast, ts...)
}

// AddFIfF 满足条件才添加需要转换的元素
func (t Seq[T]) AddFIfF(condition func(Seq[T]) bool, cast func(any) T, ts ...any) Seq[T] {
    if !condition(t) {
        return t
    }
    return t.AddF(cast, ts...)
}
