package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mzzsfy/go-util/concurrent"
)

const (
	defaultMaxWorks    = 1024
	defaultIdleTimeout = 30 * time.Second
)

var (
	// ErrPoolClosed 提交任务到已关闭协程池时返回此错误
	ErrPoolClosed = errors.New("pool is shut down")

	defaultGoPool = NewGopool(WithName("defaultGoPool"))
	taskPool      = NewObjectPool[task](func() *task { return &task{} }, func(i *task) { i.ctx = nil; i.fn = nil })
	// shutdownSentinel 用于唤醒阻塞 worker 的哨兵任务, nil 无法入队(lkQueue 丢弃 nil)
	shutdownSentinel = &task{fn: func() {}, ctx: context.Background()}
)

func Go(f func()) error {
	return defaultGoPool.Go(f)
}

func CtxGo(ctx context.Context, f func()) error {
	return defaultGoPool.CtxGo(ctx, f)
}

// GoPool 弹性协程池
// 空闲 worker 阻塞等待任务而非自旋, 超时后自动退出
type GoPool struct {
	panicHandler func(any, context.Context)
	name         string
	works        int32
	shutDown     int32
	maxWorks     int32
	idleTimeout  time.Duration
	taskQueue    concurrent.BlockQueue[*task]
	wg           sync.WaitGroup
}

func (p *GoPool) Name() string {
	return p.name
}

// WorkerCount 获取当前工作中协程数量
func (p *GoPool) WorkerCount() uint64 {
	return uint64(atomic.LoadInt32(&p.works))
}

// TaskCount 获取队列任务数量
func (p *GoPool) TaskCount() uint64 {
	return uint64(p.taskQueue.Size())
}

// drainQueue 清空队列中残留任务并执行, 返回已处理任务数
// 用于 Shutdown 后处理 dispatch 竞态入队的任务, 保证已接受的任务不被丢失
func (p *GoPool) drainQueue() int {
	count := 0
	for {
		t, ok := p.taskQueue.Dequeue()
		if !ok || t == nil {
			break
		}
		if t != shutdownSentinel {
			// 执行而非仅回收, 确保竞态入队的任务不会丢失
			p.executeTask(t)
			count++
		}
	}
	return count
}

// Shutdown 优雅关闭协程池: 停止接受新任务, 等待已有 worker 执行完队列中剩余任务
func (p *GoPool) Shutdown() bool {
	if !atomic.CompareAndSwapInt32(&p.shutDown, 0, 1) {
		return false
	}
	// 唤醒所有等待中的 worker, 让它们检查 shutdown 标记并参与排空队列
	works := atomic.LoadInt32(&p.works)
	for i := int32(0); i < works; i++ {
		p.taskQueue.Enqueue(shutdownSentinel)
	}
	// 等待所有 worker 退出
	p.wg.Wait()
	// 循环排空残留任务: dispatch 竞态可能导致任务在所有 worker 退出后入队
	// 已接受(CtxGo 返回 nil)的任务必须被执行, 循环直到队列稳定为空
	for p.drainQueue() > 0 {
	}
	return true
}

// Restart 重启协程池, 等待所有旧 worker 退出后再返回
// 注意: 超时返回 false 时, 后台 goroutine 仍会等待 wg.Wait() 完成
// 这是 sync.WaitGroup 的设计限制, 无法中断等待
// 调用方应确保 Shutdown 真正完成后再调用 Restart, 或接受超时后 goroutine 继续运行
func (p *GoPool) Restart() bool {
	if atomic.LoadInt32(&p.shutDown) != 1 {
		return false
	}
	// 等待所有 worker 退出(依赖 Shutdown 的 wg.Wait 保证)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		// 清理队列残留
		p.drainQueue()
		return atomic.CompareAndSwapInt32(&p.shutDown, 1, 0)
	case <-ctx.Done():
		// 超时: goroutine 仍会等待 wg.Wait() 完成, 无法中断
		// 设计限制: sync.WaitGroup.Wait 不支持 context 取消
		return false
	}
}

func (p *GoPool) Go(f func()) error {
	return p.CtxGo(context.Background(), f)
}

func (p *GoPool) CtxGo(ctx context.Context, f func()) error {
	if atomic.LoadInt32(&p.shutDown) == 1 {
		return ErrPoolClosed
	}
	t := taskPool.Get()
	t.fn = f
	t.ctx = ctx
	p.dispatch(t)
	return nil
}

// dispatch 分发任务: 入队并按需创建新 worker
// 任务入队后优先由 BlockQueue 唤醒空闲 worker, 同时尝试补充 worker 数量到上限内
func (p *GoPool) dispatch(t *task) {
	p.taskQueue.Enqueue(t)
	// 入队后二次检查: 若此时已 shutdown, 补发 sentinel 确保 worker 能消费残存任务
	// 场景: CtxGo 检查 shutDown==0 后被调度走, Shutdown 设 shutDown=1 并入 sentinel 退出 workers,
	// CtxGo 恢复后入队的任务可能无人处理, 补发 sentinel 唤醒 worker 来消费
	if atomic.LoadInt32(&p.shutDown) == 1 {
		p.taskQueue.Enqueue(shutdownSentinel)
		return
	}
	// CAS 创建新 worker, 缓存 maxWorks 减少热路径 atomic load
	maxW := atomic.LoadInt32(&p.maxWorks)
	for {
		w := atomic.LoadInt32(&p.works)
		if w >= maxW {
			return
		}
		if atomic.CompareAndSwapInt32(&p.works, w, w+1) {
			p.wg.Add(1)
			go p.goRun(nil)
			return
		}
	}
}

// goRun worker 主循环, 空闲时阻塞在 BlockQueue 上
func (p *GoPool) goRun(t *task) {
	defer p.wg.Done()
	defer atomic.AddInt32(&p.works, -1)
	for {
		// 先执行手头任务
		if t != nil {
			p.executeTask(t)
			t = nil
		}
		// 阻塞等待下一个任务
		t, _ = p.taskQueue.DequeueBlock(p.idleTimeout)
		// 空闲超时或收到关闭哨兵后退出
		if t == nil || t == shutdownSentinel {
			return
		}
	}
}

// executeTask 执行单个任务, 处理 panic
func (p *GoPool) executeTask(t *task) {
	defer func() {
		if a := recover(); a != nil {
			if p.panicHandler != nil {
				p.panicHandler(a, t.ctx)
			}
		}
		if t != shutdownSentinel {
			taskPool.Put(t)
		}
	}()
	t.fn()
}

type task struct {
	fn  func()
	ctx context.Context
}

// NewGopool 创建一个协程池
func NewGopool(options ...Option) *GoPool {
	gopool := &GoPool{
		maxWorks:    defaultMaxWorks,
		idleTimeout: defaultIdleTimeout,
		taskQueue: concurrent.BlockQueueWrapper(
			concurrent.NewQueue(concurrent.WithTypeSegment[*task]()),
		),
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

// WithIdleTimeout 设置 worker 空闲超时退出时间, 默认 30s
func WithIdleTimeout(d time.Duration) Option {
	return func(gopool *GoPool) {
		if d > 0 {
			gopool.idleTimeout = d
		}
	}
}

// WithMaxWorks 设置最大 worker 数量, 默认 1024
func WithMaxWorks(n int) Option {
	return func(gopool *GoPool) {
		if n > 0 {
			gopool.maxWorks = int32(n)
		}
	}
}
