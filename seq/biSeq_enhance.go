package seq

import (
    "sort"
    "sync"
)

//======增强========

// OnEach 每个元素执行f
func (t BiSeq[K, V]) OnEach(f func(K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) {
            f(k, v)
            c(k, v)
        })
    }
}

// OnEachN 每n个元素额外执行一次
func (t BiSeq[K, V]) OnEachN(step int, f func(k K, v V), skip ...int) BiSeq[K, V] {
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
            t.ForEach(func(k K, v V) { p.Add(func() { c(k, v) }) })
            p.Wait()
        } else {
            wg := sync.WaitGroup{}
            t.ForEach(func(k K, v V) {
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
    t(func(k K, v V) { r = append(r, BiTuple[K, V]{k, v}) })
    sort.Slice(r, func(i, j int) bool { return less(r[i].K, r[i].V, r[j].K, r[j].V) })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.K, v.V)
        }
    })
}

// SortK 根据K排序
func (t BiSeq[K, V]) SortK(less func(K, K) bool) BiSeq[K, V] {
    var r []BiTuple[K, V]
    t(func(k K, v V) { r = append(r, BiTuple[K, V]{k, v}) })
    sort.Slice(r, func(i, j int) bool { return less(r[i].K, r[j].K) })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.K, v.V)
        }
    })
}

// SortV 根据V排序
func (t BiSeq[K, V]) SortV(less func(V, V) bool) BiSeq[K, V] {
    var r []BiTuple[K, V]
    t(func(k K, v V) { r = append(r, BiTuple[K, V]{k, v}) })
    sort.Slice(r, func(i, j int) bool { return less(r[i].V, r[j].V) })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.K, v.V)
        }
    })
}
