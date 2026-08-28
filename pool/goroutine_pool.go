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
	defaultMaxWorkers  = 1024
	defaultIdleTimeout = 30 * time.Second
)

var (
	// ErrPoolClosed 提交任务到已关闭协程池时返回此错误
	ErrPoolClosed = errors.New("pool is shut down")

	defaultGoPool = NewGoPool(WithName("defaultGoPool"))
	taskPool      = NewObjectPool(func() *task { return &task{} }, func(i *task) { i.ctx = nil; i.fn = nil })
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
//
// 关闭语义为排空: Shutdown 前已接受(返回 nil)的任务全部执行完, Shutdown 后提交返回错误。
// 提交侧(CtxGo/dispatch)与关闭侧(Shutdown)通过 mu 串行化"检查 shutDown -> 入队"与
// "CAS shutDown"两个临界区, 保证两者必有且仅有一个生效, 不存在已接受但丢失的任务。
type GoPool struct {
	panicHandler func(any, context.Context)
	name         string
	works        int32
	shutDown     int32
	maxWorkers   int32
	idleTimeout  time.Duration
	taskQueue    concurrent.BlockQueue[*task]
	// mu 串行化三处临界区: dispatch 的"shutDown 检查+入队+补 worker"、
	// worker 超时退出的"最终检查队列+注销 works"、Shutdown/Restart 的"CAS shutDown"
	wg sync.WaitGroup
	mu sync.Mutex
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
// 返回前所有已接受任务均已执行, 队列为空
func (p *GoPool) Shutdown() bool {
	// CAS 与 dispatch 的"检查+入队"在 mu 下互斥:
	// dispatch 持锁时 Shutdown 无法 CAS; CAS 成功后 dispatch 必读到 shutDown==1 而拒绝
	p.mu.Lock()
	if !atomic.CompareAndSwapInt32(&p.shutDown, 0, 1) {
		p.mu.Unlock()
		return false
	}
	p.mu.Unlock()
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
	// 快速失败: 池运行中无需等待 worker 退出
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
	case <-ctx.Done():
		// 超时: goroutine 仍会等待 wg.Wait() 完成, 无法中断
		// 设计限制: sync.WaitGroup.Wait 不支持 context 取消
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if atomic.LoadInt32(&p.shutDown) != 1 {
		return false
	}
	// 清理队列残留
	p.drainQueue()
	return atomic.CompareAndSwapInt32(&p.shutDown, 1, 0)
}

func (p *GoPool) Go(f func()) error {
	return p.CtxGo(context.Background(), f)
}

func (p *GoPool) CtxGo(ctx context.Context, f func()) error {
	// 无锁快速失败, 真正的接受决策在 dispatch 锁内
	if atomic.LoadInt32(&p.shutDown) == 1 {
		return ErrPoolClosed
	}
	t := taskPool.Get()
	t.fn = f
	t.ctx = ctx
	if !p.dispatch(t) {
		return ErrPoolClosed
	}
	return nil
}

// dispatch 分发任务: 检查关闭状态、入队并按需创建新 worker
// 返回 false 表示池已关闭, 任务未被接受(已回收), 调用方应返回 ErrPoolClosed
// "检查 shutDown -> 入队"整体持锁, 与 Shutdown 的 CAS 互斥, 保证已接受任务不丢失
func (p *GoPool) dispatch(t *task) bool {
	p.mu.Lock()
	if atomic.LoadInt32(&p.shutDown) == 1 {
		p.mu.Unlock()
		taskPool.Put(t)
		return false
	}
	p.taskQueue.Enqueue(t)
	// 有 worker 在 BlockQueue 上阻塞等待时, Enqueue 已通过 Signal 唤醒它, 无需创建新 worker
	// waiter > 0 意味着 worker 在 cond.Wait 中必然消费该任务;
	// 即将超时退出的 worker 不在 waiter 计数内, 由 goRun 的退出前兜底检查保证不丢任务
	if p.taskQueue.WaiterCount() > 0 {
		p.mu.Unlock()
		return true
	}
	// maxWorkers 构造后不可变, 直接读避免 atomic 开销
	maxW := p.maxWorkers
	for {
		w := atomic.LoadInt32(&p.works)
		if w >= maxW {
			break
		}
		if atomic.CompareAndSwapInt32(&p.works, w, w+1) {
			p.wg.Add(1)
			go p.goRun(nil)
			break
		}
	}
	p.mu.Unlock()
	return true
}

// goRun worker 主循环, 空闲时阻塞在 BlockQueue 上
func (p *GoPool) goRun(t *task) {
	defer p.wg.Done()
	retired := false
	defer func() {
		if !retired {
			atomic.AddInt32(&p.works, -1)
		}
	}()
	for {
		// 先执行手头任务
		if t != nil {
			p.executeTask(t)
			t = nil
		}
		// 阻塞等待下一个任务
		t, _ = p.taskQueue.DequeueBlock(p.idleTimeout)
		if t == shutdownSentinel {
			// 收到关闭哨兵: 立即退出, 队列残留任务由 Shutdown 的 drainQueue 兜底执行
			atomic.AddInt32(&p.works, -1)
			retired = true
			return
		}
		if t == nil {
			// 空闲超时: 与 dispatch 互斥地做最终检查, 消除"dispatch 读到 waiter/works>0
			// 不补 worker, 而 worker 实际正在退出"的 check-then-act 窗口
			p.mu.Lock()
			t, _ = p.taskQueue.Dequeue()
			if t == nil || t == shutdownSentinel {
				atomic.AddInt32(&p.works, -1)
				p.mu.Unlock()
				if t == shutdownSentinel {
					// 误取了其他 worker 的退出信号, 归还以保证其也能退出
					p.taskQueue.Enqueue(shutdownSentinel)
				}
				retired = true
				return
			}
			p.mu.Unlock()
			// 拿到竞态期入队的任务, 继续执行
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

// NewGoPool 创建一个协程池
func NewGoPool(options ...Option) *GoPool {
	gopool := &GoPool{
		maxWorkers:  defaultMaxWorkers,
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

// WithMaxWorkers 设置最大 worker 数量, 默认 1024
func WithMaxWorkers(n int) Option {
	return func(gopool *GoPool) {
		if n > 0 {
			gopool.maxWorkers = int32(n)
		}
	}
}
