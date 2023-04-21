package seq

import (
    "sync"
    "sync/atomic"
)

//======转换========

// MapKParallel 每个元素转换为any,使用 Sync() 保证消费不竞争
// order.1 顺序保证方式,规则如下:
// 0:不保证任务启动顺序,不保证消费顺序,会消费竞争
// 1:尽量保持顺序,优先保证并发数,异步任务完成时,会直接消费,会消费竞争,可以使用 Sync() 保证消费不竞争
// 2:异步任务与消费端解偶,在保证顺序的前提下,优先保证并发数,不会消费竞争
// 3:保持异步与消费同步,以消费为准,不消费完成不会开始下一个异步任务,不会消费竞争
//
// order.2 最大并发数,根据第一个参数决定逻辑
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
        return func(c func(any, V)) {
            var currentIndex int32 = 1
            var id int32
            var fns []*BiTuple[int32, func()]
            lock := &sync.Mutex{}
            l := sync.NewCond(lock)
            p := NewParallel(sl)
            fn := func() {
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
            t(func(k K, v V) {
                var id = atomic.AddInt32(&id, 1)
                p.Add(func() {
                    a := f(k, v)
                    if o == 1 {
                        c(a, v)
                    } else if o == 2 {
                        lock.Lock()
                        defer lock.Unlock()
                        if atomic.LoadInt32(&currentIndex) != id {
                            fns = append(fns, &BiTuple[int32, func()]{id, func() { c(a, v) }})
                        } else {
                            c(a, v)
                            atomic.AddInt32(&currentIndex, 1)
                            fn()
                        }
                    } else {
                        l.L.Lock()
                        defer l.L.Unlock()
                        for atomic.LoadInt32(&currentIndex) != id {
                            l.Wait()
                        }
                        defer l.Broadcast()
                        c(a, v)
                        atomic.AddInt32(&currentIndex, 1)
                    }
                })
            })
            p.Wait()
            fn()
            fns = nil
        }
    } else {
        return t.Parallel(sl).MapK(f)
    }
}

// MapVParallel 每个元素转换为any,使用 Sync() 保证消费不竞争
// order.1 顺序保证方式,规则如下:
// 0:不保证任务启动顺序,不保证消费顺序,会消费竞争
// 1:尽量保持顺序,优先保证并发数,异步任务完成时,会直接消费,会消费竞争,可以使用 Sync() 保证消费不竞争
// 2:异步任务与消费端解偶,在保证顺序的前提下,优先保证并发数,不会消费竞争
// 3:保持异步与消费同步,以消费为准,不消费完成不会开始下一个异步任务,不会消费竞争
//
// order.2 最大并发数,根据第一个参数决定逻辑
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
        return func(c func(K, any)) {
            var currentIndex int32 = 1
            var id int32
            var fns []*BiTuple[int32, func()]
            lock := &sync.Mutex{}
            l := sync.NewCond(lock)
            p := NewParallel(sl)
            fn := func() {
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
            t(func(k K, v V) {
                var id = atomic.AddInt32(&id, 1)
                p.Add(func() {
                    a := f(k, v)
                    if o == 1 {
                        c(k, a)
                    } else if o == 2 {
                        lock.Lock()
                        defer lock.Unlock()
                        if atomic.LoadInt32(&currentIndex) != id {
                            fns = append(fns, &BiTuple[int32, func()]{id, func() { c(k, a) }})
                        } else {
                            c(k, a)
                            atomic.AddInt32(&currentIndex, 1)
                            fn()
                        }
                    } else {
                        l.L.Lock()
                        defer l.L.Unlock()
                        for atomic.LoadInt32(&currentIndex) != id {
                            l.Wait()
                        }
                        defer l.Broadcast()
                        c(k, a)
                        atomic.AddInt32(&currentIndex, 1)
                    }
                })
            })
            p.Wait()
            fn()
            fns = nil
        }
    } else {
        return t.Parallel(sl).MapV(f)
    }
}

// ExchangeKV 转换为 BiSeq[V, K]
func (t BiSeq[K, V]) ExchangeKV() BiSeq[V, K] {
    return func(c func(V, K)) { t(func(k K, v V) { c(v, k) }) }
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

// MapFlat 每个元素转换为BiSeq[any,any],并扁平化
func (t BiSeq[K, V]) MapFlat(f func(K, V) BiSeq[any, any]) BiSeq[any, any] {
    return func(c func(any, any)) {
        t(func(k K, v V) {
            s := f(k, v)
            s.ForEach(c)
        })
    }
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
