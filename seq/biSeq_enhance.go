package seq

import (
    "sort"
    "sync"
)

//======增强========

// Stoppable 调用后可以 panic(&Stop) 来主动停止迭代
func (t BiSeq[K, V]) Stoppable() BiSeq[K, V] {
    return func(c func(K, V)) {
        defer func() {
            a := recover()
            if a != nil && a != &Stop {
                panic(a)
            }
        }()
        t(func(k K, v V) { c(k, v) })
    }
}

// Catch defer recover 的简单封装
func (t BiSeq[K, V]) Catch(f func(any)) BiSeq[K, V] {
    return func(c func(K, V)) {
        defer func() {
            a := recover()
            if a != nil && a != &Stop {
                f(a)
            }
        }()
        t(func(k K, v V) { c(k, v) })
    }
}

// CatchWithValue defer recover 的简单封装,保留最后一次调用的值
func (t BiSeq[K, V]) CatchWithValue(f func(K, V, any)) BiSeq[K, V] {
    return func(c func(K, V)) {
        var lastK K
        var lastV V
        defer func() {
            a := recover()
            if a != nil && a != &Stop {
                f(lastK, lastV, a)
            }
        }()
        t(func(k K, v V) {
            lastK = k
            lastV = v
            c(k, v)
        })
    }
}

// OnEach 每个元素额外在前面执行一次
func (t BiSeq[K, V]) OnEach(f func(K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) {
            f(k, v)
            c(k, v)
        })
    }
}

// OnEachAfter 每个元素额外在后面执行一次
func (t BiSeq[K, V]) OnEachAfter(f func(K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) {
            c(k, v)
            f(k, v)
        })
    }
}

// OnEachN 每n个元素额外执行一次
func (t BiSeq[K, V]) OnEachN(step int, f func(k K, v V), skip ...int) BiSeq[K, V] {
    if step <= 0 {
        panic("step must > 0")
    }
    return func(c func(k K, v V)) {
        x := 0
        if len(skip) > 0 {
            x = -skip[0]
        }
        t(func(k K, v V) {
            x++
            if x > 0 && x%step == 0 {
                f(k, v)
            }
            c(k, v)
        })
    }
}

// OnEachNX 每n个元素额外执行一次,当结束时,如果剩余元素不足n个,额外执行一次
func (t BiSeq[K, V]) OnEachNX(step int, f func(k K, v V), skip ...int) BiSeq[K, V] {
    if step <= 0 {
        panic("step must > 0")
    }
    return func(c func(k K, v V)) {
        x := 0
        if len(skip) > 0 {
            x = -skip[0]
        }
        var lastK *K
        var lastV *V
        t(func(k K, v V) {
            x++
            lastK = &k
            lastV = &v
            if x > 0 && x%step == 0 {
                f(k, v)
            }
            c(k, v)
        })
        if x%step != 0 {
            f(*lastK, *lastV)
        }
    }
}

// OnBefore 指定位置前(包含),每个元素额外执行
func (t BiSeq[K, V]) OnBefore(i int, f func(K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        x := 0
        t(func(k K, v V) {
            if x < i {
                x++
                f(k, v)
            }
            c(k, v)
        })
    }
}

// OnAfter 指定位置后(包含),每个元素额外执行
func (t BiSeq[K, V]) OnAfter(i int, f func(K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        x := 0
        t(func(k K, v V) {
            if x >= i {
                f(k, v)
            } else {
                x++
            }
            c(k, v)
        })
    }
}

// OnFirst 执行前额外执行
func (t BiSeq[K, V]) OnFirst(f func(K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        x := 0
        t(func(k K, v V) {
            if x == 0 {
                x++
                f(k, v)
            }
            c(k, v)
        })
    }
}

// OnLast 执行完成后额外执行
func (t BiSeq[K, V]) OnLast(f func(*K, *V)) BiSeq[K, V] {
    return func(x func(K, V)) {
        var lastK *K
        var lastV *V
        t(func(k K, v V) {
            lastK = &k
            lastV = &v
            x(k, v)
        })
        f(lastK, lastV)
    }
}

// Cache 缓存Seq,使该Seq可以多次消费,并保证前面内容不会重复执行
func (t BiSeq[K, V]) Cache() BiSeq[K, V] {
    var r []BiTuple[K, V]
    once := sync.Once{}
    fn := func() {
        t(func(k K, v V) { r = append(r, BiTuple[K, V]{k, v}) })
    }
    return func(k func(K, V)) {
        once.Do(fn)
        for _, v := range r {
            k(v.K, v.V)
        }
    }
}

// Sync 串行执行
func (t BiSeq[K, V]) Sync() BiSeq[K, V] {
    lock := sync.Mutex{}
    return func(c func(K, V)) {
        t(func(k K, v V) {
            lock.Lock()
            defer lock.Unlock()
            c(k, v)
        })
    }
}

// Parallel 对后续操作启用并行执行,使用 Sync() 保证消费不竞争
func (t BiSeq[K, V]) Parallel(concurrency ...int) BiSeq[K, V] {
    sl := 0
    if len(concurrency) > 0 {
        sl = concurrency[0]
    }
    return func(c func(k K, v V)) {
        if sl > 0 {
            p := NewParallel(sl)
            t(func(k K, v V) { p.Add(func() { c(k, v) }) })
            p.Wait()
        } else {
            wg := sync.WaitGroup{}
            t(func(k K, v V) {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    c(k, v)
                }()
            })
            wg.Wait()
        }
    }
}

// Sort 排序
func (t BiSeq[K, V]) Sort(less func(K, V, K, V) bool) BiSeq[K, V] {
    var r []BiTuple[K, V]
    once := sync.Once{}
    fn := func() {
        t(func(k K, v V) { r = append(r, BiTuple[K, V]{k, v}) })
        sort.Slice(r, func(i, j int) bool { return less(r[i].K, r[i].V, r[j].K, r[j].V) })
    }
    return BiFrom(func(k func(K, V)) {
        once.Do(fn)
        for _, v := range r {
            k(v.K, v.V)
        }
    })
}

// SortK 根据K排序
func (t BiSeq[K, V]) SortK(less func(K, K) bool) BiSeq[K, V] {
    var r []BiTuple[K, V]
    once := sync.Once{}
    fn := func() {
        t(func(k K, v V) { r = append(r, BiTuple[K, V]{k, v}) })
        sort.Slice(r, func(i, j int) bool { return less(r[i].K, r[j].K) })
    }
    return BiFrom(func(k func(K, V)) {
        once.Do(fn)
        for _, v := range r {
            k(v.K, v.V)
        }
    })
}

// SortV 根据V排序
func (t BiSeq[K, V]) SortV(less func(V, V) bool) BiSeq[K, V] {
    var r []BiTuple[K, V]
    once := sync.Once{}
    fn := func() {
        t(func(k K, v V) { r = append(r, BiTuple[K, V]{k, v}) })
        sort.Slice(r, func(i, j int) bool { return less(r[i].V, r[j].V) })
    }
    return BiFrom(func(k func(K, V)) {
        once.Do(fn)
        for _, v := range r {
            k(v.K, v.V)
        }
    })
}
