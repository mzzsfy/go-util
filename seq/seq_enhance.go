package seq

import (
    "sort"
    "sync"
)

//======增强,不改变内容========

// Stoppable 调用后可以 panic(&Stop) 来主动停止迭代
func (t Seq[T]) Stoppable() Seq[T] {
    return func(c func(T)) {
        defer func() {
            a := recover()
            if a != nil && a != &Stop {
                panic(a)
            }
        }()
        t(func(t T) { c(t) })
    }
}

// Catch defer recover 的简单封装
func (t Seq[T]) Catch(f func(any)) Seq[T] {
    return func(c func(T)) {
        defer func() {
            a := recover()
            if a != nil && a != &Stop {
                f(a)
            }
        }()
        t(func(t T) { c(t) })
    }
}

// Finally defer 的简单封装
func (t Seq[T]) Finally(f func()) Seq[T] {
    return func(c func(T)) {
        defer f()
        t(func(t T) { c(t) })
    }
}

// CatchWithValue defer recover 的简单封装,保留最后一次调用的值
func (t Seq[T]) CatchWithValue(f func(T, any)) Seq[T] {
    return func(c func(T)) {
        var last T
        defer func() {
            a := recover()
            if a != nil && a != &Stop {
                f(last, a)
            }
        }()
        t(func(t T) {
            last = t
            c(t)
        })
    }
}

// OnEach 每个元素额外执行
func (t Seq[T]) OnEach(f func(T)) Seq[T] {
    return func(c func(T)) {
        t(func(t T) {
            f(t)
            c(t)
        })
    }
}

// OnEachN 每n个元素额外执行一次
func (t Seq[T]) OnEachN(step int, f func(T), skip ...int) Seq[T] {
    if step <= 0 {
        panic("step must > 0")
    }
    return func(c func(T)) {
        x := 0
        if len(skip) > 0 {
            x = -skip[0]
        }
        t(func(t T) {
            x++
            if x > 0 && x%step == 0 {
                f(t)
            }
            c(t)
        })
    }
}

// OnEachNX 每n个元素额外执行一次,当结束时,如果剩余元素不足n个,额外执行一次
func (t Seq[T]) OnEachNX(step int, f func(T), skip ...int) Seq[T] {
    if step <= 0 {
        panic("step must > 0")
    }
    return func(c func(T)) {
        x := 0
        if len(skip) > 0 {
            x = -skip[0]
        }
        var last *T
        t(func(t T) {
            x++
            last = &t
            if x > 0 && x%step == 0 {
                f(t)
            }
            c(t)
        })
        if x%step != 0 {
            f(*last)
        }
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

// OnFirst 执行前额外执行
func (t Seq[T]) OnFirst(f func(T)) Seq[T] {
    return func(c func(T)) {
        x := 0
        t(func(t T) {
            if x == 0 {
                x++
                f(t)
            }
            c(t)
        })
    }
}

// OnLast 执行完成后额外执行
func (t Seq[T]) OnLast(f func(*T)) Seq[T] {
    return func(c func(T)) {
        var last *T
        t(func(t T) {
            last = &t
            c(t)
        })
        f(last)
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
func (t Seq[T]) Parallel(concurrent ...int) Seq[T] {
    sl := 0
    if len(concurrent) > 0 {
        sl = concurrent[0]
    }
    return func(c func(T)) {
        if sl > 0 {
            p := NewParallel(sl)
            t(func(t T) { p.Add(func() { c(t) }) })
            p.Wait()
        } else {
            wg := sync.WaitGroup{}
            t(func(t T) {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    c(t)
                }()
            })
            wg.Wait()
        }
    }
}

// Sort 排序
func (t Seq[T]) Sort(less func(T, T) bool) Seq[T] {
    var r []T
    once := sync.Once{}
    fn := func() {
        t(func(t T) { r = append(r, t) })
        sort.Slice(r, func(i, j int) bool { return less(r[i], r[j]) })
    }
    return func(t func(T)) {
        once.Do(fn)
        for _, v := range r {
            t(v)
        }
    }
}

// Cache 缓存Seq,使该Seq可以多次重复消费,并保证前面内容不会重复执行
func (t Seq[T]) Cache() Seq[T] {
    var r []T
    once := sync.Once{}
    fn := func() {
        t(func(t T) { r = append(r, t) })
    }
    return func(t func(T)) {
        once.Do(fn)
        for _, v := range r {
            t(v)
        }
    }
}

// Repeat 重复该Seq n次,如果不传递,则无限重复
func (t Seq[T]) Repeat(n ...int) Seq[T] {
    return func(f func(T)) {
        if len(n) == 0 {
            for {
                t(f)
            }
        } else {
            l := n[0]
            for i := 0; i < l; i++ {
                t(f)
            }
        }
    }
}
