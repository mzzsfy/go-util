package seq

import (
    "context"
    "golang.org/x/sync/semaphore"
    "sort"
    "sync"
)

//======增强========

// OnEach 每个元素额外执行
func (t Seq[T]) OnEach(f func(T)) Seq[T] {
    return func(c func(T)) {
        t(func(t T) {
            f(t)
            c(t)
        })
    }
}

// OnBefore 指定位置前(包含),每个元素额外执行
func (t Seq[T]) OnBefore(i int, f func(T)) Seq[T] {
    return func(c func(T)) {
        x := 0
        t(func(t T) {
            if x < i {
                x++
                f(t)
            }
            c(t)
        })
    }
}

// OnAfter 指定位置后(包含),每个元素额外执行
func (t Seq[T]) OnAfter(i int, f func(T)) Seq[T] {
    return func(c func(T)) {
        x := 0
        t(func(t T) {
            if x >= i {
                f(t)
            } else {
                x++
            }
            c(t)
        })
    }
}

// Sync 串行执行
func (t Seq[T]) Sync() Seq[T] {
    lock := sync.Mutex{}
    return func(c func(T)) {
        t(func(t T) {
            lock.Lock()
            defer lock.Unlock()
            c(t)
        })
    }
}

// Parallel 对后续操作启用并行执行 使用 Sync() 保证消费不竞争
func (t Seq[T]) Parallel(concurrency ...int) Seq[T] {
    sl := 0
    if len(concurrency) > 0 {
        sl = concurrency[0]
    }
    return func(c func(T)) {
        s := semaphore.NewWeighted(int64(sl))
        t.Map(func(t T) any {
            lock := sync.Mutex{}
            lock.Lock()
            go func() {
                defer lock.Unlock()
                if sl > 0 {
                    s.Acquire(context.Background(), 1)
                    defer s.Release(1)
                }
                c(t)
            }()
            return &lock
        }).Cache()(func(t any) {
            lock := t.(sync.Locker)
            lock.Lock()
        })
    }
}

// Sort 排序
func (t Seq[T]) Sort(less func(T, T) bool) Seq[T] {
    var r []T
    t(func(t T) { r = append(r, t) })
    sort.Slice(r, func(i, j int) bool { return less(r[i], r[j]) })
    return FromSlice(r)
}

// Distinct 去重
func (t Seq[T]) Distinct(eq func(T, T) bool) Seq[T] {
    var r []T
    t(func(t T) {
        for _, v := range r {
            if eq(t, v) {
                return
            }
        }
        r = append(r, t)
    })
    return FromSlice(r)
}

// Cache 缓存Seq,使该Seq可以多次重复消费,并保证前面内容不会重复执行
func (t Seq[T]) Cache() Seq[T] {
    var r []T
    t(func(t T) { r = append(r, t) })
    return FromSlice(r)
}
