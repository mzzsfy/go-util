package pool

import (
    "context"
    "errors"
    "runtime"
    "sync"
    "sync/atomic"
)

var (
    //workerIdGen   uint64
    //taskIdGen     uint64
    defaultGoPool = NewGopool(WithName("defaultGoPool"), WithMaxSize(uint32(runtime.NumCPU()*1000)))
    workerPool    = NewObjectPool[worker](func() *worker { return &worker{} }, func(i *worker) { i.p = nil })
    taskPool      = NewObjectPool[task](func() *task { return &task{} }, func(i *task) { i.next = nil; i.fn = nil })
)

func Go(f func()) error {
    return defaultGoPool.Go(f)
}
func CtxGo(ctx context.Context, f func()) error {
    return defaultGoPool.CtxGo(ctx, f)
}

type GoPool struct {
    //coreSize以下的,每个goroutine负责一个任务,超过部分每1个任务由loadProbability(0.0~1.0)个goroutine负责
    loadProbability float32
    panicHandler    func(any, context.Context)
    name            string
    coreSize        uint32
    maxSize         uint32
    works           uint32
    //剩余任务数量,当任务开始执行时就不被计数
    tasks     uint32
    shutDown  int32
    taskLock  sync.Mutex
    taskHead  *task
    taskTail  *task
    closeChan sync.Once
}

func (p *GoPool) Name() string {
    return p.name
}

// WorkerCount 获取当前工作中携程数量
func (p *GoPool) WorkerCount() uint {
    return uint(atomic.LoadUint32(&p.works))
}

// TaskCount 获取任务数量
func (p *GoPool) TaskCount() uint {
    return uint(atomic.LoadUint32(&p.tasks))
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
            if !p.newWorker() {
                break
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
    atomic.AddUint32(&p.tasks, 1)
    //t.id = atomic.AddUint64(&taskIdGen, 1)
    t.fn = f
    t.ctx = ctx

    p.taskLock.Lock()
    if p.taskHead == nil {
        p.taskHead = t
        p.taskTail = t
    } else {
        p.taskTail.next = t
        p.taskTail = t
    }
    p.taskLock.Unlock()

    //少于coreSize,直接扩容
    p.newWorker()
    return nil
}

func (p *GoPool) newWorker() bool {
    if p.loadProbability == 0 || atomic.LoadUint32(&p.works) == 0 || atomic.LoadUint32(&p.works) < p.coreSize {
        p.runWorker()
    } else {
        //满足扩容条件,扩容
        if uint32(float32(atomic.LoadUint32(&p.tasks))*p.loadProbability) > (atomic.LoadUint32(&p.works)-p.coreSize) && (p.maxSize == 0 || atomic.LoadUint32(&p.works) < p.maxSize) {
            p.runWorker()
        } else {
            return false
        }
    }
    return true
}

func (p *GoPool) runWorker() {
    w := workerPool.Get()
    w.p = p
    //w.id = atomic.AddUint64(&workerIdGen, 1)
    go w.run()
}

type worker struct {
    //id uint64
    p *GoPool
}

func (w *worker) run() {
    atomic.AddUint32(&w.p.works, 1)
    defer func() {
        atomic.AddUint32(&w.p.works, ^uint32(0))
        workerPool.Put(w)
    }()
    for {
        if !w.run1() {
            break
        }
    }
}
func (w *worker) run1() (Continue bool) {
    var t *task
    defer func() {
        if t != nil && w.p.panicHandler != nil {
            if a := recover(); a != nil {
                Continue = true
                w.p.panicHandler(a, t.ctx)
            }
        }
    }()
    for {
        if atomic.LoadInt32(&w.p.shutDown) == 1 {
            return
        }
        if w.p.taskHead != nil {
            w.p.taskLock.Lock()
            t = w.p.taskHead
            if t != nil {
                w.p.taskHead = t.next
            }
            w.p.taskLock.Unlock()
        }
        if t == nil {
            return
        }
        atomic.AddUint32(&w.p.tasks, ^uint32(0))
        t.fn()
        taskPool.Put(t)
        t = nil
    }
}

type task struct {
    //id   uint64
    fn   func()
    ctx  context.Context
    next *task
}

// NewGopool 创建一个携程池,不推荐使用(性能有待优化)
// Deprecated
func NewGopool(options ...Option) *GoPool {
    gopool := &GoPool{
        coreSize:        uint32(runtime.NumCPU() * 100),
        loadProbability: float32(0.3),
    }
    for _, option := range options {
        option(gopool)
    }
    return gopool
}

type Option func(gopool *GoPool)

func WithCoreSize(size uint32) Option {
    return func(gopool *GoPool) {
        gopool.maxSize = size
    }
}
func WithMaxSize(size uint32) Option {
    return func(gopool *GoPool) {
        if size != 0 && gopool.coreSize > size {
            panic("maxSize must be more than coreSize")
        }
        gopool.maxSize = size
    }
}
func WithLoadProbability(probability float32) Option {
    if probability < 0 || probability > 1 {
        panic("LoadProbability must be between 0 and 1")
    }
    if probability == 1 {
        probability = 0
    }
    if probability == 0 {
        return func(gopool *GoPool) {
            gopool.loadProbability = probability
        }
    }
    return func(gopool *GoPool) {
        gopool.loadProbability = probability
    }
}
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
