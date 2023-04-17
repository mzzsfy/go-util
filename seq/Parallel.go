package seq

import (
    "sync"
    "sync/atomic"
)

type Parallel interface {
    // Add 添加一个函数到并行执行队列中
    Add(fn func())
    // Wait 等待所有函数执行完成
    Wait()
}

type parallel struct {
    fns     chan func()
    release chan struct{}
    once    sync.Once
    runner  int32
    running int32
}

//todo
func (p *parallel) init() {
    p.once.Do(func() {
        for fn := range p.fns {
            <-p.release
            if atomic.AddInt32(&p.running, 1) >= p.runner {
                atomic.AddInt32(&p.running, -1)
                fn()
            } else {
                go fn()
            }
        }
        p.once = sync.Once{}
    })
}
func (p *parallel) Add(fn func()) {

}

func (p *parallel) Wait() {
    //TODO implement me
    panic("implement me")
}
