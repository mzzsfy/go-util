package seq

import (
    "sync"
    "sync/atomic"
)

type Parallel interface {
    // Add 添加一个函数到并行执行队列中,当并发数达到上限时, 会阻塞等待
    Add(fn func())
    // Wait 等待所有函数执行完成
    Wait()
}

func NewParallel(concurrent int) Parallel {
    if concurrent <= 0 {
        panic("concurrent must > 0")
    }
    return &parallel{
        concurrent: int32(concurrent),
        cond:       sync.Cond{L: &sync.Mutex{}},
    }
}

type parallel struct {
    concurrent int32
    running    int32
    cond       sync.Cond
    // ParallelFunc 自定义携程任务运行模式
    ParallelFunc func(fn func())
    error        any
}

func (p *parallel) Add(fn func()) {
    p.cond.L.Lock()
    defer p.cond.L.Unlock()
    for atomic.LoadInt32(&p.running) >= p.concurrent {
        p.cond.Wait()
    }
    if p.error != nil {
        panic(p.error)
    }
    atomic.AddInt32(&p.running, 1)
    pf := DefaultParallelFunc
    if p.ParallelFunc != nil {
        pf = p.ParallelFunc
    }
    pf(func() {
        defer func() {
            if r := recover(); r != nil {
                p.error = r
            }
            p.cond.L.Lock()
            defer p.cond.L.Unlock()
            atomic.AddInt32(&p.running, -1)
            p.cond.Broadcast()
        }()
        fn()
    })
}

func (p *parallel) Wait() {
    p.cond.L.Lock()
    defer p.cond.L.Unlock()
    for atomic.LoadInt32(&p.running) > 0 {
        p.cond.Wait()
    }
    if p.error != nil {
        panic(p.error)
    }
}
