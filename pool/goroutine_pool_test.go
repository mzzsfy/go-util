package pool

import (
    "context"
    "errors"
    "math/rand"
    "runtime"
    "sync"
    "sync/atomic"
    "testing"
    "time"

    "github.com/mzzsfy/go-util/unsafe"
)

const sleepTime = time.Millisecond * 3

func getSleepTime() time.Duration {
    return sleepTime * time.Duration(rand.Intn(100))
}

// testGopoolGoParams 合并原 TestGopool_Go / TestGopool_Go1 的重复逻辑
func testGopoolGoParams(t *testing.T, n int) {
    t.Helper()
    x := int32(n)
    maxGoroutine := 0
    lock := sync.Mutex{}
    wg := sync.WaitGroup{}
    wg.Add(n)
    pool := NewGopool()
    for i := 0; i < n; i++ {
        pool.Go(func() {
            time.Sleep(getSleepTime())
            goroutine := runtime.NumGoroutine()
            lock.Lock()
            if goroutine > maxGoroutine {
                maxGoroutine = goroutine
            }
            lock.Unlock()
            wg.Done()
            atomic.AddInt32(&x, -1)
        })
    }
    wg.Wait()
    pool.Shutdown()
    if x != 0 {
        t.Fatalf("x != 0, got %d", x)
    }
    if count := defaultGoPool.TaskCount(); count != 0 {
        t.Fatalf("count != 0, got %d", count)
    }
    t.Logf("n=%d maxGoroutine=%d", n, maxGoroutine)
}

func TestGopool_Go(t *testing.T) {
    t.Parallel()
    t.Log("start 200000")
    testGopoolGoParams(t, 200000)
}

func TestGopool_Go1(t *testing.T) {
    t.Parallel()
    t.Log("start 20000")
    testGopoolGoParams(t, 20000)
}

func Benchmark_Go(b *testing.B) {
    n := b.N
    x := int32(n)
    wg := sync.WaitGroup{}
    wg.Add(n)
    var goid int64
    go func() {
        atomic.StoreInt64(&goid, unsafe.GoID())
    }()
    for i := 0; i < n; i++ {
        go func() {
            time.Sleep(getSleepTime())
            atomic.AddInt32(&x, -1)
            wg.Done()
        }()
    }
    wg.Wait()
    if b.N > 600000 {
        time.Sleep(time.Millisecond * 50)
        go func() {
            b.Log("goid start", atomic.LoadInt64(&goid), "end", unsafe.GoID(), "new", unsafe.GoID()-atomic.LoadInt64(&goid))
        }()
    }
    if x != 0 {
        b.Fatal("x != 0", x)
    }
    if defaultGoPool.TaskCount() != 0 {
        b.Fatal("count != 0", defaultGoPool.TaskCount())
    }
    time.Sleep(time.Millisecond * 50)
}

func BenchmarkGopool_Go(b *testing.B) {
    wg := sync.WaitGroup{}
    n := b.N
    x := int32(n)
    wg.Add(n)
    gopool := NewGopool()
    var goid int64
    go func() {
        atomic.StoreInt64(&goid, unsafe.GoID())
    }()
    for i := 0; i < n; i++ {
        gopool.Go(func() {
            time.Sleep(getSleepTime())
            atomic.AddInt32(&x, -1)
            wg.Done()
        })
    }
    wg.Wait()
    if b.N > 600000 {
        time.Sleep(time.Millisecond * 50)
        go func() {
            b.Log("goid start", atomic.LoadInt64(&goid), "end", unsafe.GoID(), "new", unsafe.GoID()-atomic.LoadInt64(&goid))
        }()
    }
    if x != 0 {
        b.Fatal("x != 0", x)
    }
    if defaultGoPool.TaskCount() != 0 {
        b.Fatal("count != 0", defaultGoPool.TaskCount())
    }
    time.Sleep(time.Millisecond * 50)
}

func TestGoPool_Shutdown_NoTaskDrop(t *testing.T) {
    t.Parallel()
    const n = 100
    var executed int32

    pool := NewGopool()

    // 入队 n 个任务, 每个任务将 executed +1
    var wg sync.WaitGroup
    wg.Add(n)
    for i := 0; i < n; i++ {
        pool.Go(func() {
            atomic.AddInt32(&executed, 1)
            wg.Done()
        })
    }

    // 立即 shutdown, 但所有任务应执行完
    pool.Shutdown()

    // Shutdown 会等 worker 退出, 但 wg 确保任务全部执行
    wg.Wait()

    if executed != n {
        t.Fatalf("expected %d tasks executed, got %d", n, executed)
    }
}

func TestGoPoolShutdownRestart(t *testing.T) {
    t.Parallel()

    pool := NewGopool(WithIdleTimeout(10 * time.Millisecond))

    const firstBatch = 16
    var firstExecuted int32
    var firstWG sync.WaitGroup
    firstWG.Add(firstBatch)
    for i := 0; i < firstBatch; i++ {
        err := pool.Go(func() {
            atomic.AddInt32(&firstExecuted, 1)
            firstWG.Done()
        })
        if err != nil {
            t.Fatalf("submit first batch failed: %v", err)
        }
    }

    firstWG.Wait()
    if !pool.Shutdown() {
        t.Fatal("shutdown should succeed")
    }
    if firstExecuted != firstBatch {
        t.Fatalf("expected first batch %d, got %d", firstBatch, firstExecuted)
    }

    if !pool.Restart() {
        t.Fatal("restart should succeed")
    }

    const secondBatch = 16
    var secondExecuted int32
    var secondWG sync.WaitGroup
    secondWG.Add(secondBatch)
    for i := 0; i < secondBatch; i++ {
        err := pool.Go(func() {
            atomic.AddInt32(&secondExecuted, 1)
            secondWG.Done()
        })
        if err != nil {
            t.Fatalf("submit second batch failed: %v", err)
        }
    }

    secondWG.Wait()
    if secondExecuted != secondBatch {
        t.Fatalf("expected second batch %d, got %d", secondBatch, secondExecuted)
    }
    if !pool.Shutdown() {
        t.Fatal("second shutdown should succeed")
    }
}

func TestGoPoolConcurrentShutdownRestart(t *testing.T) {
    t.Parallel()

    pool := NewGopool(WithIdleTimeout(10 * time.Millisecond))

    // 并发调 Shutdown, 只有一个人能 CAS 成功, 其余直接返回 false
    var shutdownWG sync.WaitGroup
    for i := 0; i < 8; i++ {
        shutdownWG.Add(1)
        go func() {
            defer shutdownWG.Done()
            pool.Shutdown()
        }()
    }
    shutdownWG.Wait()

    // 并发调 Restart, 只有一个人能 CAS 成功
    var restartWG sync.WaitGroup
    for i := 0; i < 8; i++ {
        restartWG.Add(1)
        go func() {
            defer restartWG.Done()
            pool.Restart()
        }()
    }
    restartWG.Wait()

    // 确保池处于可用状态
    for !pool.Restart() {
        if atomic.LoadInt32(&pool.shutDown) == 0 {
            break
        }
        runtime.Gosched()
    }

    // 验证 Restart 后可以正常提交和执行任务
    var executed int32
    var verifyWG sync.WaitGroup
    verifyWG.Add(1)
    if err := pool.Go(func() {
        atomic.AddInt32(&executed, 1)
        verifyWG.Done()
    }); err != nil {
        t.Fatalf("submit verify task failed: %v", err)
    }
    verifyWG.Wait()
    if executed != 1 {
        t.Fatalf("expected verify task executed once, got %d", executed)
    }
    pool.Shutdown()
}

// TestGoPool_CtxGo_Shutdown_Race 验证并发提交与 Shutdown 之间的竞态安全性:
// 1. 不应产生 data race
// 2. 所有成功提交的任务都应执行
// 3. Shutdown 应在合理时间内完成
func TestGoPool_CtxGo_Shutdown_Race(t *testing.T) {
    t.Parallel()

    pool := NewGopool(WithIdleTimeout(50 * time.Millisecond))

    const totalTasks = 2000
    var executed int32
    var submitWg sync.WaitGroup
    submitWg.Add(totalTasks)

    // 并发提交任务, 同时触发 Shutdown
    shutdownDone := make(chan struct{})
    go func() {
        // 等待所有任务提交完成后再 shutdown, 避免时序敏感
        time.Sleep(10 * time.Millisecond)
        pool.Shutdown()
        close(shutdownDone)
    }()

    // 记录成功提交的任务数
    var submitted int32
    for i := 0; i < totalTasks; i++ {
        err := pool.Go(func() {
            atomic.AddInt32(&executed, 1)
            submitWg.Done()
        })
        if err == nil {
            atomic.AddInt32(&submitted, 1)
        } else {
            // shutdown 后提交失败, 需要手动 Done 以避免 Wait 永远阻塞
            submitWg.Done()
        }
    }

    // 等待所有成功提交的任务执行完毕
    submitWg.Wait()

    // 验证: 成功提交的任务全部被执行
    got := atomic.LoadInt32(&executed)
    sub := atomic.LoadInt32(&submitted)
    if got != sub {
        t.Fatalf("executed %d != submitted %d, 部分任务丢失", got, sub)
    }

    // 验证 shutdown 确实完成了
    select {
    case <-shutdownDone:
    case <-time.After(10 * time.Second):
        t.Fatal("Shutdown 未在预期时间内完成")
    }
}

// TestGoPool_Shutdown_DispatchRace 验证 dispatch/Shutdown 竞态不会丢失任务
// 构造场景: 大量 goroutine 在 CtxGo 检查 shutDown==0 后被调度走, Shutdown 完成,
// 然后这些 goroutine 恢复并调用 dispatch 入队任务。验证这些任务仍被执行。
func TestGoPool_Shutdown_DispatchRace(t *testing.T) {
    t.Parallel()

    for i := 0; i < 10; i++ {
        pool := NewGopool(WithIdleTimeout(10 * time.Millisecond))

        const n = 100
        var executed int32
        var wg sync.WaitGroup
        wg.Add(n)

        // 并发提交任务
        for j := 0; j < n; j++ {
            go func() {
                // 故意制造竞态: 先检查, 再延迟, 再提交
                // 增加被调度走的几率
                runtime.Gosched()
                err := pool.Go(func() {
                    atomic.AddInt32(&executed, 1)
                    wg.Done()
                })
                if err != nil {
                    // 提交失败, 手动 Done
                    wg.Done()
                }
            }()
        }

        // 立即 shutdown, 制造竞态窗口
        pool.Shutdown()

        // 等待所有任务完成或确认失败
        done := make(chan struct{})
        go func() {
            wg.Wait()
            close(done)
        }()

        select {
        case <-done:
        case <-time.After(5 * time.Second):
            t.Fatal("等待任务完成超时")
        }

        // 验证: 所有成功提交的任务都被执行
        // 由于竞态, 可能不是所有 n 个都成功提交, 但已提交的必须执行
        got := atomic.LoadInt32(&executed)
        if got < 0 {
            t.Fatalf("executed 为负数: %d", got)
        }
        // 主要验证: 没有 panic, 没有死锁
    }
}

// TestGoPool_ErrPoolClosed 验证 shutdown 后提交返回 sentinel error, 且 errors.Is 可匹配
func TestGoPool_ErrPoolClosed(t *testing.T) {
    t.Parallel()

    p := NewGopool()
    p.Shutdown()

    err := p.Go(func() {})
    if err == nil {
        t.Fatal("expected error after shutdown")
    }
    if !errors.Is(err, ErrPoolClosed) {
        t.Fatalf("expected ErrPoolClosed, got %v", err)
    }
    if err.Error() != "pool is shut down" {
        t.Fatalf("unexpected error message: %q", err.Error())
    }
}

// TestGoPool_PanicHandler 验证 WithPanicHandler 能捕获 panic 并获取 panic 值和 context
func TestGoPool_PanicHandler(t *testing.T) {
    t.Parallel()

    var (
        panicVal any
        panicCtx context.Context
        panicWG  sync.WaitGroup
    )
    panicWG.Add(1)

    pool := NewGopool(WithIdleTimeout(10*time.Millisecond), WithPanicHandler(func(v any, ctx context.Context) {
        panicVal = v
        panicCtx = ctx
        panicWG.Done()
    }))

    // 提交一个会 panic 的任务
    err := pool.Go(func() {
        panic("test-panic")
    })
    if err != nil {
        t.Fatalf("submit panic task failed: %v", err)
    }

    // 等待 panic handler 被调用
    panicWG.Wait()

    // 验证 handler 收到了正确的 panic 值
    if panicVal != "test-panic" {
        t.Fatalf("expected panic value 'test-panic', got %v", panicVal)
    }
    // 验证 handler 收到了非 nil 的 context
    if panicCtx == nil {
        t.Fatal("expected non-nil context in panic handler")
    }

    // 验证 panic 后协程池仍然可以正常工作
    var executed int32
    var normalWG sync.WaitGroup
    normalWG.Add(1)
    err = pool.Go(func() {
        atomic.AddInt32(&executed, 1)
        normalWG.Done()
    })
    if err != nil {
        t.Fatalf("submit normal task after panic failed: %v", err)
    }
    normalWG.Wait()
    if executed != 1 {
        t.Fatalf("expected 1 task executed after panic, got %d", executed)
    }

    pool.Shutdown()
}

// TestGoPool_PackageLevelFunctions 验证包级别 Go() 和 CtxGo() 函数正常工作
func TestGoPool_PackageLevelFunctions(t *testing.T) {
    t.Parallel()

    var (
        goExecuted    int32
        ctxGoExecuted int32
        wg            sync.WaitGroup
    )
    wg.Add(2)

    // 使用包级别 Go() 提交任务
    if err := Go(func() {
        atomic.AddInt32(&goExecuted, 1)
        wg.Done()
    }); err != nil {
        t.Fatalf("package-level Go() failed: %v", err)
    }

    // 使用包级别 CtxGo() 提交任务
    if err := CtxGo(context.Background(), func() {
        atomic.AddInt32(&ctxGoExecuted, 1)
        wg.Done()
    }); err != nil {
        t.Fatalf("package-level CtxGo() failed: %v", err)
    }

    wg.Wait()

    if goExecuted != 1 {
        t.Fatalf("expected Go() executed 1, got %d", goExecuted)
    }
    if ctxGoExecuted != 1 {
        t.Fatalf("expected CtxGo() executed 1, got %d", ctxGoExecuted)
    }
}

// TestGoPool_WorkerCountAndName 验证 Name() 和 WorkerCount() 访问器
func TestGoPool_WorkerCountAndName(t *testing.T) {
    t.Parallel()

    pool := NewGopool(WithName("test-pool"), WithIdleTimeout(10*time.Millisecond))

    // 验证 Name() 返回正确值
    if pool.Name() != "test-pool" {
        t.Fatalf("expected name 'test-pool', got %q", pool.Name())
    }

    // 提交足够的任务来确保 worker 被创建
    const n = 16
    var wg sync.WaitGroup
    wg.Add(n)
    // 用一个阻塞通道来保持 worker 存活
    block := make(chan struct{})
    for i := 0; i < n; i++ {
        pool.Go(func() {
            <-block
            wg.Done()
        })
    }

    // 短暂等待 worker 启动
    time.Sleep(50 * time.Millisecond)

    // 验证 WorkerCount() > 0
    wc := pool.WorkerCount()
    if wc == 0 {
        t.Fatal("expected WorkerCount() > 0 after submitting tasks")
    }

    // 释放所有 worker
    close(block)
    wg.Wait()
    pool.Shutdown()
}

// TestGoPool_RestartReturnsFalseWhenRunning 验证 Restart 在池运行中时返回 false
func TestGoPool_RestartReturnsFalseWhenRunning(t *testing.T) {
    t.Parallel()

    pool := NewGopool()

    // 池正在运行, Restart 应返回 false
    if pool.Restart() {
        t.Fatal("Restart() should return false when pool is running")
    }

    // 关闭池
    if !pool.Shutdown() {
        t.Fatal("Shutdown() should succeed")
    }

    // 池已关闭, Restart 应返回 true
    if !pool.Restart() {
        t.Fatal("Restart() should return true after shutdown")
    }

    // 池已重启, 再次 Restart 应返回 false
    if pool.Restart() {
        t.Fatal("Restart() should return false when pool is already restarted")
    }

    // 清理: 关闭池
    pool.Shutdown()
}

// TestGoPool_WithMaxWorks 验证 WithMaxWorks 限制最大 worker 数量
func TestGoPool_WithMaxWorks(t *testing.T) {
    t.Parallel()

    const maxWorkers = 5
    pool := NewGopool(WithMaxWorks(maxWorkers), WithIdleTimeout(10*time.Millisecond))

    // 用阻塞通道保持 worker 存活, 提交足够多的任务让 worker 尝试创建
    const totalTasks = 50
    block := make(chan struct{})
    var wg sync.WaitGroup
    wg.Add(totalTasks)
    for i := 0; i < totalTasks; i++ {
        err := pool.Go(func() {
            <-block
            wg.Done()
        })
        if err != nil {
            t.Fatalf("提交任务 %d 失败: %v", i, err)
        }
    }

    // 短暂等待 worker 启动
    time.Sleep(100 * time.Millisecond)

    // WorkerCount 不应超过 maxWorkers
    wc := pool.WorkerCount()
    if wc > uint64(maxWorkers) {
        t.Fatalf("WorkerCount 应 <= %d, got %d", maxWorkers, wc)
    }

    // 释放所有任务
    close(block)
    wg.Wait()
    pool.Shutdown()
}

// TestGoPool_WithMaxWorks_AllTasksDone 验证限制 worker 数量后所有任务仍能完成
func TestGoPool_WithMaxWorks_AllTasksDone(t *testing.T) {
    t.Parallel()

    pool := NewGopool(WithMaxWorks(3), WithIdleTimeout(10*time.Millisecond))

    const totalTasks = 100
    var executed int32
    var wg sync.WaitGroup
    wg.Add(totalTasks)
    for i := 0; i < totalTasks; i++ {
        err := pool.Go(func() {
            atomic.AddInt32(&executed, 1)
            wg.Done()
        })
        if err != nil {
            t.Fatalf("提交任务 %d 失败: %v", i, err)
        }
    }

    wg.Wait()
    if got := atomic.LoadInt32(&executed); got != totalTasks {
        t.Fatalf("应执行 %d 个任务, 实际 %d", totalTasks, got)
    }
    pool.Shutdown()
}

// TestGoPool_Shutdown_CompletesQueuedTasks 验证 Shutdown 后所有已入队的排队任务都被执行完
// 构造方式: 让少量 worker 阻塞 -> 提交大量排队任务 -> 调 Shutdown -> 解除阻塞 -> 检查全部完成
func TestGoPool_Shutdown_CompletesQueuedTasks(t *testing.T) {
    t.Parallel()

    pool := NewGopool(WithMaxWorks(2), WithIdleTimeout(30*time.Second))

    var executed int32
    const totalTasks = 100
    // 用阻塞通道让 2 个 worker 占满
    block := make(chan struct{})

    // 先提交 2 个阻塞任务占满 worker
    for i := 0; i < 2; i++ {
        err := pool.Go(func() {
            <-block
            atomic.AddInt32(&executed, 1)
        })
        if err != nil {
            t.Fatalf("提交阻塞任务失败: %v", err)
        }
    }

    // 等待 worker 启动
    time.Sleep(20 * time.Millisecond)

    // 提交排队任务
    var wg sync.WaitGroup
    wg.Add(totalTasks - 2)
    for i := 0; i < totalTasks-2; i++ {
        err := pool.Go(func() {
            atomic.AddInt32(&executed, 1)
            wg.Done()
        })
        if err != nil {
            // 如果提交失败(已被 shutdown), 手动 Done 避免死锁
            wg.Done()
        }
    }

    // 异步调用 Shutdown, 此时 worker 仍被阻塞, sentinel 入队但无人消费
    shutdownDone := make(chan struct{})
    go func() {
        pool.Shutdown()
        close(shutdownDone)
    }()

    // 短暂等待确保 Shutdown 的 CAS 已执行
    time.Sleep(10 * time.Millisecond)

    // 解除阻塞, 让 worker 开始处理排队任务
    close(block)

    // 验证 Shutdown 在合理时间内完成
    select {
    case <-shutdownDone:
        // ok
    case <-time.After(10 * time.Second):
        t.Fatal("Shutdown 未在预期时间内完成")
    }

    // 等待所有排队任务完成
    wg.Wait()

    // 所有任务应被成功执行
    if got := atomic.LoadInt32(&executed); got != totalTasks {
        t.Fatalf("应执行 %d 个任务, 实际 %d", totalTasks, got)
    }
}
