package seq

import (
    "sync"
    "sync/atomic"
)

// mapParallelCore 并行转换的有序调度核心,order>0 时使用,order 语义见 MapParallel
func mapParallelCore[T, R any](t Seq[T], syncFn func(T) R, o int, sl int) Seq[R] {
    return func(c func(R)) {
        var currentIndex int32 = 1
        var id int32
        // 待执行的有序回调,map查找O(1)
        fns := make(map[int32]func())
        lock := &sync.Mutex{}
        l := sync.NewCond(lock)
        p := NewParallel(sl)
        // 按序消费已完成的任务
        fn := func() {
            lock.Lock()
            defer lock.Unlock()
            for {
                f, ok := fns[atomic.LoadInt32(&currentIndex)]
                if !ok {
                    break
                }
                delete(fns, currentIndex)
                f()
                atomic.AddInt32(&currentIndex, 1)
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
                        fns[id] = func() { c(a) }
                    } else {
                        c(a)
                        atomic.AddInt32(&currentIndex, 1)
                        if len(fns) > 0 {
                            DefaultParallelFunc(fn)
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
    }
}

// biMapVParallelCore 并行转换的有序调度核心,order>0 时使用
func biMapVParallelCore[K, V, R any](t BiSeq[K, V], f func(K, V) R, o int, sl int) BiSeq[K, R] {
    return func(c func(K, R)) {
        var currentIndex int32 = 1
        var id int32
        fns := make(map[int32]func())
        lock := &sync.Mutex{}
        l := sync.NewCond(lock)
        p := NewParallel(sl)
        fn := func() {
            lock.Lock()
            defer lock.Unlock()
            for {
                idx := atomic.LoadInt32(&currentIndex)
                g, ok := fns[idx]
                if !ok {
                    break
                }
                delete(fns, idx)
                g()
                atomic.AddInt32(&currentIndex, 1)
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
                        fns[id] = func() { c(k, a) }
                    } else {
                        c(k, a)
                        atomic.AddInt32(&currentIndex, 1)
                        if len(fns) > 0 {
                            DefaultParallelFunc(fn)
                        }
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
    }
}

// biMapVCore 常规顺序映射核心
func biMapVCore[K, V, R any](t BiSeq[K, V], f func(K, V) R) BiSeq[K, R] {
    return func(c func(K, R)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// mapParallelCustomizeCore 并行转换自定义调度核心
func mapParallelCustomizeCore[T, R any](t Seq[T], asyncFn func(T, func(R))) Seq[R] {
    return func(c func(R)) {
        wg := sync.WaitGroup{}
        var err any
        t(func(t T) {
            wg.Add(1)
            asyncFn(t, func(r R) {
                defer func() {
                    if a := recover(); a != nil {
                        err = a
                    }
                    wg.Done()
                }()
                c(r)
            })
            if err != nil {
                panic(err)
            }
        })
        wg.Wait()
    }
}
