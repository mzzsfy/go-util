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
}

func (p *parallel) Add(fn func()) {
    p.cond.L.Lock()
    defer p.cond.L.Unlock()
    for atomic.LoadInt32(&p.running) >= p.concurrent {
        p.cond.Wait()
    }
    atomic.AddInt32(&p.running, 1)
    go func() {
        defer func() {
            p.cond.L.Lock()
            defer p.cond.L.Unlock()
            atomic.AddInt32(&p.running, -1)
            p.cond.Broadcast()
        }()
        fn()
    }()
}

func (p *parallel) Wait() {
    p.cond.L.Lock()
    defer p.cond.L.Unlock()
    for atomic.LoadInt32(&p.running) > 0 {
        p.cond.Wait()
    }
}

func NewAsyncParallel(concurrent int) Parallel {
    if concurrent <= 0 {
        panic("concurrent must > 0")
    }
    return &asyncParallel{
        concurrent: int32(concurrent),
    }
}

type asyncParallel struct {
    concurrent int32
    running    int32
    wg         sync.WaitGroup
    fns        []func()
    fnLock     sync.Mutex
    doFns      []func()
    doLock     sync.Mutex
    cacheFns   []func()
    cacheLock  sync.Mutex
}

func (p *asyncParallel) do() {
    defer func() {
        atomic.AddInt32(&p.running, -1)
    }()
    for {
        p.doLock.Lock()
        if len(p.doFns) == 0 {
            if len(p.fns) == 0 {
                p.doLock.Unlock()
                return
            }
            p.doFns = p.fns
            p.fns = nil
        }
        fn := p.doFns[0]
        p.doFns = p.doFns[1:]
        p.doLock.Unlock()
        fn()
    }
}
func (p *asyncParallel) Add(fn func()) {
    p.wg.Add(1)
    nfn := func() {
        defer p.wg.Done()
        fn()
    }
    //无竞争
    if p.fnLock.TryLock() {
        defer func() {
            p.cacheLock.Lock()
            if len(p.cacheFns) == 0 {
                p.cacheLock.Unlock()
                return
            }
            p.fns = append(p.fns, p.cacheFns...)
            p.cacheFns = nil
            p.cacheLock.Unlock()
            if len(p.doFns) == 0 {
                p.doFns = p.fns
                p.fns = nil
            }
            l := int(p.concurrent - atomic.LoadInt32(&p.running))
            for i := 0; i < l; i++ {
                atomic.AddInt32(&p.running, 1)
                go p.do()
            }
            p.fnLock.Unlock()
        }()
    } else {
        p.cacheLock.Lock()
        p.cacheFns = append(p.cacheFns, nfn)
        p.cacheLock.Unlock()
        return
    }
    p.fns = append(p.fns, nfn)
    if atomic.LoadInt32(&p.running) >= p.concurrent {
        return
    }
    atomic.AddInt32(&p.running, 1)
    go p.do()
}

func (p *asyncParallel) Wait() {
    p.wg.Wait()
}
