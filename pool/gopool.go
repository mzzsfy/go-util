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
    defaultGopool = NewGopool(uint32(runtime.NumCPU()*30), withName("defaultGopool"), withMaxSize(uint32(runtime.NumCPU()*1000)))
    workerPool    = NewObjpool[worker](func() *worker { return &worker{} }, func(i *worker) { i.p = nil })
    taskPool      = NewObjpool[task](func() *task { return &task{} }, func(i *task) { i.next = nil; i.fn = nil })
)

func Go(f func()) error {
    return defaultGopool.Go(f)
}
func CtxGo(ctx context.Context, f func()) error {
    return defaultGopool.CtxGo(ctx, f)
}

type Gopool struct {
    loadProbability float32
    panicHandler    func(any, context.Context)
    name            string
    coreSize        uint32
    maxSize         uint32
    works           uint32
    //剩余任务数量,当任务开始执行时就不被计数
    tasks     uint64
    shutDown  int32
    taskLock  sync.Mutex
    taskHead  *task
    taskTail  *task
    closeChan sync.Once
}

func (p *Gopool) Name() string {
    return p.name
}

// WorkerCount 获取当前工作中携程数量
func (p *Gopool) WorkerCount() uint {
    return uint(p.works)
}

// TaskCount 获取任务数量
func (p *Gopool) TaskCount() uint {
    return uint(p.tasks)
}

// Shutdown 关闭携程池,停止接受新任务,停止运行新任务
func (p *Gopool) Shutdown() bool {
    return atomic.CompareAndSwapInt32(&p.shutDown, 0, 1)
}

// Restart 重启携程池
func (p *Gopool) Restart() bool {
    r := atomic.CompareAndSwapInt32(&p.shutDown, 1, 0)
    if r {
        for {
            if atomic.LoadUint32(&p.works) < p.coreSize {
                p.newWorker()
            } else {
                if uint32(float32(atomic.LoadUint64(&p.tasks))*p.loadProbability) < atomic.LoadUint32(&p.works) {
                    p.newWorker()
                } else {
                    break
                }
            }
        }
    }
    return r
}

func (p *Gopool) Go(f func()) error {
    return p.CtxGo(context.Background(), f)
}

func (p *Gopool) CtxGo(ctx context.Context, f func()) error {
    if atomic.LoadInt32(&p.shutDown) == 1 {
        return errors.New("pool is shut down")
    }
    t := taskPool.Get()
    atomic.AddUint64(&p.tasks, 1)
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
    if atomic.LoadUint32(&p.works) < p.coreSize {
        p.newWorker()
    } else {
        //满足扩容条件,扩容
        if uint32(float32(atomic.LoadUint64(&p.tasks))*p.loadProbability) < atomic.LoadUint32(&p.works) {
            p.newWorker()
        }
    }
    return nil
}

func (p *Gopool) newWorker() {
    w := workerPool.Get()
    w.p = p
    //w.id = atomic.AddUint64(&workerIdGen, 1)
    go w.run()
}

type worker struct {
    //id uint64
    p *Gopool
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
        atomic.AddUint64(&w.p.tasks, ^uint64(0))
        //println("111", t.id)
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

func NewGopool(coreSize uint32, options ...Option) *Gopool {
    gopool := &Gopool{
        coreSize:        coreSize,
        loadProbability: 0.2,
    }
    for _, option := range options {
        option(gopool)
    }
    return gopool
}

type Option func(gopool *Gopool)

func WithLoadProbability(probability float32) Option {
    if probability <= 0 || probability >= 1 {
        panic("LoadProbability must be between 0 and 1")
    }
    return func(gopool *Gopool) {
        gopool.loadProbability = probability
    }
}
func WithPanicHandler(handler func(any, context.Context)) Option {
    return func(gopool *Gopool) {
        gopool.panicHandler = handler
    }
}
func withMaxSize(size uint32) Option {
    return func(gopool *Gopool) {
        if size != 0 && gopool.coreSize > size {
            panic("maxSize must be more than coreSize")
        }
        gopool.maxSize = size
    }
}

func withName(name string) Option {
    return func(gopool *Gopool) {
        gopool.name = name
    }
}
