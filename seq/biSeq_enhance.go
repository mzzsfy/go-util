package seq

import (
    "context"
    "golang.org/x/sync/semaphore"
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
    return func(c func(K, V)) {
        lock := sync.Mutex{}
        t(func(k K, v V) {
            lock.Lock()
            defer lock.Unlock()
            c(k, v)
        })
    }
}

// Parallel 对后续操作启用并行执行
func (t BiSeq[K, V]) Parallel(concurrency ...int) BiSeq[K, V] {
    sl := 0
    if len(concurrency) > 0 {
        sl = concurrency[0]
    }
    return func(c func(K, V)) {
        var b BiSeq[any, any]
        if sl > 0 {
            s := semaphore.NewWeighted(int64(sl))
            b = t.Map(func(k K, v V) (any, any) {
                lock := sync.Mutex{}
                lock.Lock()
                go func() {
                    defer lock.Unlock()
                    s.Acquire(context.Background(), 1)
                    defer s.Release(1)
                    c(k, v)
                }()
                return &lock, nil
            })
        } else {
            b = t.Map(func(k K, v V) (any, any) {
                lock := sync.Mutex{}
                lock.Lock()
                go func() {
                    defer lock.Unlock()
                    c(k, v)
                }()
                return &lock, nil
            })
        }
        b.Cache()(func(t, _ any) {
            lock := t.(sync.Locker)
            lock.Lock()
        })
    }
}

// Sort 排序
func (t BiSeq[K, V]) Sort(less func(K, V, K, V) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) { r = append(r, biTuple[K, V]{k, v}) })
    sort.Slice(r, func(i, j int) bool { return less(r[i].k, r[i].v, r[j].k, r[j].v) })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// SortK 根据K排序
func (t BiSeq[K, V]) SortK(less func(K, K) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) { r = append(r, biTuple[K, V]{k, v}) })
    sort.Slice(r, func(i, j int) bool { return less(r[i].k, r[j].k) })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// SortV 根据V排序
func (t BiSeq[K, V]) SortV(less func(V, V) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) { r = append(r, biTuple[K, V]{k, v}) })
    sort.Slice(r, func(i, j int) bool { return less(r[i].v, r[j].v) })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// Distinct 去重
func (t BiSeq[K, V]) Distinct(eq func(K, V, K, V) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) {
        for _, x := range r {
            if eq(k, v, x.k, x.v) {
                return
            }
        }
        r = append(r, biTuple[K, V]{k, v})
    })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// DistinctK 使用K去重
func (t BiSeq[K, V]) DistinctK(eq func(K, K) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) {
        for _, x := range r {
            if eq(k, x.k) {
                return
            }
        }
        r = append(r, biTuple[K, V]{k, v})
    })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// DistinctV 使用V去重
func (t BiSeq[K, V]) DistinctV(eq func(V, V) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) {
        for _, x := range r {
            if eq(v, x.v) {
                return
            }
        }
        r = append(r, biTuple[K, V]{k, v})
    })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}
