package seq

import (
    "sort"
    "sync"
)

//======增强,不改变内容========

// Stoppable 调用后可以使用 panic(&Stop) 来主动停止迭代,否则会导致panic
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

// RecoverErr defer recover 的简单封装,在发生panic时,会调用f函数,任何位置可用
func (t Seq[T]) RecoverErr(f func(any)) Seq[T] {
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

// Deprecated: 不要使用这个方法,方法名称有歧义,请使用 RecoverErr
func (t Seq[T]) Catch(f func(any)) Seq[T] {
    return t.RecoverErr(f)
}

// RecoverErrWithValue defer recover 的简单封装,保留最后一次调用的值
func (t Seq[T]) RecoverErrWithValue(f func(T, any)) Seq[T] {
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

// Deprecated: 不要使用这个方法,方法名称有歧义,请使用 RecoverErrWithValue
func (t Seq[T]) CatchWithValue(f func(T, any)) Seq[T] {
    return t.RecoverErrWithValue(f)
}

// Finally defer 的简单封装
func (t Seq[T]) Finally(f func()) Seq[T] {
    return func(c func(T)) {
        defer f()
        t(func(t T) { c(t) })
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

// OnEachF 每个元素根据f函数判断是否需要额外执行
func (t Seq[T]) OnEachF(step func(T) bool, f func(T), skip ...int) Seq[T] {
    return func(c func(T)) {
        x := 0
        if len(skip) > 0 {
            x = -skip[0]
        }
        t(func(t T) {
            x++
            if x > 0 && step(t) {
                f(t)
            }
            c(t)
        })
    }
}

// OnEachN 每n个元素额外执行
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
            t(func(t T) {
                p.Add(func() {
                    c(t)
                })
            })
            p.Wait()
        } else {
            wg := sync.WaitGroup{}
            var err any
            t(func(t T) {
                wg.Add(1)
                DefaultParallelFunc(func() {
                    defer func() {
                        if a := recover(); a != nil {
                            err = a
                        }
                        wg.Done()
                    }()
                    c(t)
                })
                if err != nil {
                    panic(err)
                }
            })
            wg.Wait()
        }
    }
}

// ParallelCustomize 自定义并行执行策略
func (t Seq[T]) ParallelCustomize(fn func(T, func())) Seq[T] {
    return func(c func(T)) {
        wg := sync.WaitGroup{}
        var err any
        t(func(t T) {
            wg.Add(1)
            fn(t, func() {
                defer func() {
                    if a := recover(); a != nil {
                        err = a
                    }
                    wg.Done()
                }()
                c(t)
            })
            if err != nil {
                panic(err)
            }
        })
        wg.Wait()
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

// SortCustomize 自定义排序
func (t Seq[T]) SortCustomize(sort func([]T)) Seq[T] {
    var r []T
    once := sync.Once{}
    fn := func() {
        t(func(t T) { r = append(r, t) })
        sort(r)
    }
    return func(t func(T)) {
        once.Do(fn)
        for _, v := range r {
            t(v)
        }
    }
}

// Reverse 逆序
func (t Seq[T]) Reverse() Seq[T] {
    var r []T
    once := sync.Once{}
    fn := func() {
        t(func(t T) { r = append(r, t) })
    }
    return func(t func(T)) {
        once.Do(fn)
        for i := len(r) - 1; i >= 0; i-- {
            t(r[i])
        }
    }
}

// Cache 缓存Seq,使该Seq可以多次重复消费,init为true时,会立刻触发消费行为
func (t Seq[T]) Cache(init ...bool) Seq[T] {
    var r []T
    once := sync.Once{}
    fn := func() {
        t(func(t T) { r = append(r, t) })
    }
    if len(init) > 0 && init[0] {
        once.Do(fn)
    }
    return func(t func(T)) {
        once.Do(fn)
        for _, v := range r {
            t(v)
        }
    }
}

// Repeat 重复该Seq n次,如果不传递n,则无限重复,当前seq如果比较重,建议使用cache缓存
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
