package pool

import (
    "context"
    "errors"
    "github.com/mzzsfy/go-util/concurrent"
    "runtime"
    "sync"
    "sync/atomic"
)

var (
    //workerIdGen   uint64
    //taskIdGen     uint64
    defaultGoPool = NewGopool(WithName("defaultGoPool"))
    taskPool      = NewObjectPool[task](func() *task { return &task{} }, func(i *task) { i.ctx = nil; i.fn = nil })
)

func Go(f func()) error {
    return defaultGoPool.Go(f)
}
func CtxGo(ctx context.Context, f func()) error {
    return defaultGoPool.CtxGo(ctx, f)
}

type GoPool struct {
    panicHandler func(any, context.Context)
    name         string
    works        int32
    shutDown     int32
    taskQueue    concurrent.Queue[*task]
    closeChan    sync.Once
}

func (p *GoPool) Name() string {
    return p.name
}

// WorkerCount 获取当前工作中携程数量
func (p *GoPool) WorkerCount() uint64 {
    return uint64(atomic.LoadInt32(&p.works))
}

// TaskCount 获取队列任务数量
func (p *GoPool) TaskCount() uint64 {
    return uint64(p.taskQueue.Size())
}

// Shutdown 关闭携程池,停止接受新任务,停止运行新任务
func (p *GoPool) Shutdown() bool {
    return atomic.CompareAndSwapInt32(&p.shutDown, 0, 1)
}

// Restart 重启携程池
func (p *GoPool) Restart() bool {
    r := atomic.CompareAndSwapInt32(&p.shutDown, 1, 0)
    if r {
        for {
            t, b := p.taskQueue.Dequeue()
            if b {
                go p.goRun(t)
            } else {
                return r
            }
        }
    }
    return r
}

func (p *GoPool) Go(f func()) error {
    return p.CtxGo(context.Background(), f)
}

func (p *GoPool) CtxGo(ctx context.Context, f func()) error {
    if atomic.LoadInt32(&p.shutDown) == 1 {
        return errors.New("pool is shut down")
    }
    t := taskPool.Get()
    //t.id = atomic.AddUint64(&taskIdGen, 1)
    t.fn = f
    t.ctx = ctx
    p.newWorker(t)
    return nil
}

func (p *GoPool) newWorker(t *task) bool {
    if p.taskQueue.Size() > 10 || int(atomic.LoadInt32(&p.works)) < 10 {
        go p.goRun(t)
        return true
    } else {
        p.taskQueue.Enqueue(t)
        return false
    }
}

func (p *GoPool) goRun(t *task) {
    atomic.AddInt32(&p.works, 1)
    defer atomic.AddInt32(&p.works, -1)
    for {
        if !p.goRun1(t) {
            break
        }
    }
}
func (p *GoPool) goRun1(t *task) (Continue bool) {
    defer func() {
        if t != nil && p.panicHandler != nil {
            if a := recover(); a != nil {
                Continue = true
                p.panicHandler(a, t.ctx)
            }
        }
    }()
    if t != nil {
        t.fn()
        taskPool.Put(t)
        t = nil
    }
    var (
        ok   = false
        fail = 0
    )
    for {
        if atomic.LoadInt32(&p.shutDown) == 1 {
            return
        }
        t, ok = p.taskQueue.Dequeue()
        if !ok {
            if fail > 10 || (fail > 0 && p.works > 1024) {
                return
            }
            fail++
            runtime.Gosched()
            continue
        }
        t.fn()
        taskPool.Put(t)
        t = nil
        fail = 0
    }
}

type task struct {
    //id   uint64
    fn  func()
    ctx context.Context
}

// NewGopool 创建一个携程池,不推荐使用(性能有待优化)
func NewGopool(options ...Option) *GoPool {
    gopool := &GoPool{
        taskQueue: concurrent.NewQueue(concurrent.WithTypeLink[*task]()),
    }
    for _, option := range options {
        option(gopool)
    }
    return gopool
}

type Option func(gopool *GoPool)

func WithPanicHandler(handler func(any, context.Context)) Option {
    return func(gopool *GoPool) {
        gopool.panicHandler = handler
    }
}

func WithName(name string) Option {
    return func(gopool *GoPool) {
        gopool.name = name
    }
}
