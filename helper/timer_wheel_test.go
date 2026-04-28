package helper

import (
    "fmt"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

// ==================== 测试辅助工具 ====================

// panicRecord 记录一次 panic 事件
type panicRecord struct {
    Task
    Recovered any
}

// execTracker 统一的执行追踪器，替代 mockExecutor/newCountingExecutor/newBoolExecutor/mockPanicHandler
type execTracker struct {
    count  int32
    mu     sync.Mutex
    tasks  []Task
    panics []panicRecord
    times  []int64
}

// executor 通用执行器：计数并记录任务
func (et *execTracker) executor(task Task) {
    atomic.AddInt32(&et.count, 1)
    et.mu.Lock()
    et.tasks = append(et.tasks, task)
    et.mu.Unlock()
    task.Run()
}

// timingExecutor 记录执行时间戳的执行器
func (et *execTracker) timingExecutor(task Task) {
    et.mu.Lock()
    et.times = append(et.times, time.Now().UnixMilli())
    et.mu.Unlock()
    task.Run()
}

// panicHandler 记录 panic 事件
func (et *execTracker) panicHandler(task Task, recovered any) {
    et.mu.Lock()
    et.panics = append(et.panics, panicRecord{task, recovered})
    et.mu.Unlock()
}

// newTestWheel 创建测试用 TimerWheel，自动注册 Stop 清理
func newTestWheel(t *testing.T, tick time.Duration, executor func(Task)) *TimerWheel {
    t.Helper()
    if executor == nil {
        executor = func(task Task) { task.Run() }
    }
    w := NewTimerWheel(WithTickInterval(tick), WithExecutor(executor))
    t.Cleanup(func() { w.Stop() })
    return w
}

// assertTaskCount 断言时间轮的 taskCount 等于预期值
func assertTaskCount(t *testing.T, w *TimerWheel, expected int32, msg string) {
    t.Helper()
    if got := atomic.LoadInt32(&w.taskCount); got != expected {
        t.Fatalf("%s: taskCount=%d, 预期 %d", msg, got, expected)
    }
}

// assertClean 合并检查 taskCount 和槽位残留
func assertClean(t *testing.T, w *TimerWheel, msg string) {
    t.Helper()
    if got := atomic.LoadInt32(&w.taskCount); got != 0 {
        t.Fatalf("%s: taskCount=%d, 预期 0", msg, got)
    }
    if remaining := countSlotRemaining(w); remaining != 0 {
        t.Fatalf("%s: 槽位残留 %d 个任务, 预期 0", msg, remaining)
    }
}

// countSlotRemaining 遍历所有槽位，统计残留任务指针数（仅被 assertClean 使用）
func countSlotRemaining(w *TimerWheel) int {
    total := 0
    for _, layer := range w.layers {
        for ci := 0; ci < len(layer.cells); ci++ {
            layer.cells[ci].mu.Lock()
            total += len(layer.cells[ci].tasks)
            layer.cells[ci].mu.Unlock()
        }
    }
    return total
}

// newPanicTestCtx 创建带 panic 捕获的时间轮测试上下文，用于 ScheduleCustom panic 测试
type panicTestCtx struct {
    panicCalls       int32
    handlerRecovered atomic.Value
    notified         chan struct{}
    W                *TimerWheel
    Tick             time.Duration
}

func newPanicTestCtx(tick time.Duration) *panicTestCtx {
    c := &panicTestCtx{notified: make(chan struct{}, 1), Tick: tick}
    c.W = NewTimerWheel(
        WithTickInterval(tick),
        WithPanicHandler(func(_ Task, recovered any) {
            atomic.AddInt32(&c.panicCalls, 1)
            c.handlerRecovered.Store(recovered)
            select {
            case c.notified <- struct{}{}:
            default:
            }
        }),
        WithExecutor(func(task Task) { task.Run() }),
    )
    return c
}

// waitFor 等待 channel 信号，超时则 fatal
func waitFor(t *testing.T, ch <-chan struct{}, timeout time.Duration, msg string) {
    t.Helper()
    select {
    case <-ch:
    case <-time.After(timeout):
        t.Fatal(msg)
    }
}

// ==================== Schedule 单次调度 ====================

func TestTimerWheel_Schedule(t *testing.T) {
    t.Parallel()
    // 表驱动：零延迟、负延迟、正常延迟统一验证执行语义
    t.Run("DelayVariants", func(t *testing.T) {
        t.Parallel()
        tick := 50 * time.Millisecond
        tests := []struct {
            name  string
            delay time.Duration
        }{
            {"零延迟", 0},
            {"负延迟当作零延迟", -100 * time.Millisecond},
            {"正常延迟", 100 * time.Millisecond},
        }
        for _, tt := range tests {
            tt := tt
            t.Run(tt.name, func(t *testing.T) {
                t.Parallel()
                var et execTracker
                w := newTestWheel(t, tick, et.executor)

                done := make(chan struct{})
                w.Schedule(tt.delay, FuncTask(func() { close(done) }))
                waitFor(t, done, 500*time.Millisecond, "任务未在预期时间内执行")
            })
        }
    })

    t.Run("ExecutesAfterDelay", func(t *testing.T) {
        t.Parallel()
        tick := 50 * time.Millisecond
        var et execTracker
        w := newTestWheel(t, tick, et.executor)
        delay := 100 * time.Millisecond
        start := time.Now()
        w.Schedule(delay, FuncTask(func() {}))
        time.Sleep(delay + tick*3)
        et.mu.Lock()
        executedCount := len(et.tasks)
        et.mu.Unlock()
        if executedCount == 0 {
            t.Error("任务未执行")
            return
        }
        if elapsed := time.Since(start); elapsed < delay-tick {
            t.Errorf("任务执行过早: %v, 预期至少 %v", elapsed, delay)
        }
    })

    t.Run("ZeroDelay_TaskCountStable", func(t *testing.T) {
        t.Parallel()
        w := newTestWheel(t, 100*time.Millisecond, nil)
        w.Schedule(0, FuncTask(func() {}))
        assertTaskCount(t, w, 0, "零延迟单次任务执行后计数错误")
    })

    t.Run("ZeroDelay_DefaultExecutor_Regression", func(t *testing.T) {
        t.Parallel()
        // 验证零延迟单次任务在默认异步 executor 下 fn 不被提前清空（Bug 回归）
        var ran int32
        w := NewTimerWheel(WithTickInterval(50 * time.Millisecond))
        defer w.Stop()

        const n = 100
        for i := 0; i < n; i++ {
            w.Schedule(0, FuncTask(func() {
                atomic.AddInt32(&ran, 1)
            }))
        }

        deadline := time.After(2 * time.Second)
        for {
            if atomic.LoadInt32(&ran) >= n {
                break
            }
            select {
            case <-deadline:
                t.Fatalf("零延迟任务执行数不足: %d, 预期 %d（fn 可能被提前清空）", atomic.LoadInt32(&ran), n)
            default:
                time.Sleep(10 * time.Millisecond)
            }
        }
    })

    // 边界延迟测试：第0层和高层边界
    t.Run("BoundaryTiming", func(t *testing.T) {
        t.Parallel()
        tick := 50 * time.Millisecond

        tests := []struct {
            name        string
            delay       time.Duration
            extraWait   time.Duration
            maxAcceptFn func(boundaryMs, tickMs int64) int64
        }{
            {"Layer0Boundary", tick * 10, tick * 3, func(b, t int64) int64 { return b*2 - t }},
            {"HighLayerBoundary", tick*10 + 10*time.Millisecond, tick * 4, func(b, t int64) int64 { return b + t*3 }},
        }
        for _, tt := range tests {
            tt := tt
            t.Run(tt.name, func(t *testing.T) {
                t.Parallel()
                var et execTracker
                w := newTestWheel(t, tick, et.timingExecutor)

                start := time.Now().UnixMilli()
                w.Schedule(tt.delay, FuncTask(func() {}))
                time.Sleep(tt.delay*2 + tt.extraWait)

                et.mu.Lock()
                times := et.times
                et.mu.Unlock()
                if len(times) == 0 {
                    t.Fatal("边界延迟任务未执行")
                }
                actualDelay := times[0] - start
                maxAccept := tt.maxAcceptFn(tt.delay.Milliseconds(), tick.Milliseconds())
                if actualDelay > maxAccept {
                    t.Errorf("边界延迟任务执行过晚: 实际延迟 %dms, 预期不超过 %dms",
                        actualDelay, maxAccept)
                }
            })
        }
    })

    // 层结构验证：覆盖范围、长延迟层级、槽位 floor 语义
    t.Run("LayerStructure", func(t *testing.T) {
        t.Parallel()
        // 验证常见长延迟在时间轮覆盖范围内
        t.Run("OverflowCoverage", func(t *testing.T) {
            t.Parallel()
            w := newTestWheel(t, 100*time.Millisecond, nil)

            delay := 3 * time.Hour
            maxCoverage := w.layers[len(w.layers)-1].totalInterval
            if maxCoverage <= delay.Milliseconds() {
                t.Errorf("时间轮最大覆盖 %dms 不足以支持 %v 延迟 (需要 %dms)",
                    maxCoverage, delay, delay.Milliseconds())
            }
        })

        // 长延迟进入高层且短时间内不执行
        t.Run("LongDelayInHigherLayer", func(t *testing.T) {
            t.Parallel()
            var et execTracker
            tick := 100 * time.Millisecond
            w := newTestWheel(t, tick, et.executor)
            w.Schedule(20*time.Second, FuncTask(func() {}))
            if len(w.layers) != 8 {
                t.Fatalf("时间轮层数不正确: %d, 预期 8 层", len(w.layers))
            }
            expectedRanges := []int64{
                tick.Milliseconds() * 10, tick.Milliseconds() * 100,
                tick.Milliseconds() * 1000, tick.Milliseconds() * 10000,
                tick.Milliseconds() * 100000, tick.Milliseconds() * 1000000,
                tick.Milliseconds() * 10000000, tick.Milliseconds() * 100000000,
            }
            for i, layer := range w.layers {
                if layer.totalInterval != expectedRanges[i] {
                    t.Errorf("第 %d 层范围不正确: %d, 预期 %d", i, layer.totalInterval, expectedRanges[i])
                }
            }
            // 短时间内不应执行
            time.Sleep(500 * time.Millisecond)
            if atomic.LoadInt32(&et.count) > 0 {
                t.Error("长延迟任务不应该在这么短时间内执行")
            }
        })

        // 高层槽位使用 floor 语义
        t.Run("HighLayerCellIdxFloor", func(t *testing.T) {
            t.Parallel()
            tick := int64(100)
            interval := tick * 10

            tests := []struct {
                name     string
                delay    int64
                expected int
            }{
                {"刚过边界", interval + 1, 1},
                {"1.5倍", interval + interval/2, 1},
                {"刚好2倍", interval * 2, 2},
                {"接近总范围", interval*9 - 1, 8},
            }
            for _, tt := range tests {
                tt := tt
                t.Run(tt.name, func(t *testing.T) {
                    t.Parallel()
                    steps := int(tt.delay / interval)
                    if steps != tt.expected {
                        t.Errorf("delay=%d: floor steps = %d, 预期 %d", tt.delay, steps, tt.expected)
                    }
                })
            }
        })
    })

    t.Run("TimeRollback", func(t *testing.T) {
        t.Parallel()
        // 模拟时间回调后任务仍可执行
        var et execTracker
        w := newTestWheel(t, 100*time.Millisecond, et.executor)
        w.Schedule(50*time.Millisecond, FuncTask(func() {}))
        time.Sleep(300 * time.Millisecond)
        if atomic.LoadInt32(&et.count) == 0 {
            t.Error("任务未执行")
        }
    })
}

// ==================== ScheduleRepeating 重复任务 ====================

func TestTimerWheel_ScheduleRepeating(t *testing.T) {
    t.Parallel()
    t.Run("ExecutesMultipleTimes", func(t *testing.T) {
        t.Parallel()
        tick := 50 * time.Millisecond
        var et execTracker
        w := newTestWheel(t, tick, et.executor)
        interval := 100 * time.Millisecond
        w.ScheduleRepeating(interval, FuncTask(func() {}))
        time.Sleep(interval*5 + tick*2)
        if count := atomic.LoadInt32(&et.count); count < 3 {
            t.Errorf("重复任务执行次数不足: %d, 预期至少 3 次", count)
        }
    })

    t.Run("Cancel", func(t *testing.T) {
        t.Parallel()
        tick := 50 * time.Millisecond
        var et execTracker
        w := newTestWheel(t, tick, et.executor)
        interval := 100 * time.Millisecond
        handle := w.ScheduleRepeating(interval, FuncTask(func() {}))
        // 等待几次执行后取消
        time.Sleep(interval*2 + tick)
        handle.Cancel()
        countBeforeCancel := atomic.LoadInt32(&et.count)
        time.Sleep(interval * 3)
        countAfterWait := atomic.LoadInt32(&et.count)
        if countAfterWait > countBeforeCancel+1 {
            t.Errorf("取消后仍有新执行: 之前 %d, 之后 %d", countBeforeCancel, countAfterWait)
        }
    })

    t.Run("RepeatingNoDrift", func(t *testing.T) {
        t.Parallel()
        // 重复任务间隔稳定，不会漂移
        var et execTracker
        tick := 50 * time.Millisecond
        w := newTestWheel(t, tick, et.timingExecutor)
        interval := 100 * time.Millisecond
        w.ScheduleRepeating(interval, FuncTask(func() {}))
        time.Sleep(interval*5 + tick*3)
        et.mu.Lock()
        times := et.times
        et.mu.Unlock()
        if len(times) < 4 {
            t.Fatalf("重复任务执行次数不足: %d, 预期至少 4 次", len(times))
        }
        for i := 1; i < len(times) && i < 5; i++ {
            diff := times[i] - times[i-1] - interval.Milliseconds()
            if diff < -tick.Milliseconds()*2 || diff > tick.Milliseconds()*2 {
                t.Errorf("第 %d 次执行间隔偏差过大: 差值 %dms", i, diff)
            }
        }
    })

    t.Run("SlowTask_NoConcurrentReentry", func(t *testing.T) {
        t.Parallel()
        // 验证重复任务在任务体执行时间超过 interval 时不会并发重入（回归）
        tick := 50 * time.Millisecond
        interval := 100 * time.Millisecond
        var maxConcurrent, currentConcurrent, totalRuns int32
        w := NewTimerWheel(WithTickInterval(tick), WithExecutor(func(task Task) { task.Run() }))
        defer w.Stop()
        w.ScheduleRepeating(interval, FuncTask(func() {
            cur := atomic.AddInt32(&currentConcurrent, 1)
            atomic.AddInt32(&totalRuns, 1)
            for {
                old := atomic.LoadInt32(&maxConcurrent)
                if cur <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, cur) {
                    break
                }
            }
            // 模拟慢任务：执行时间是 interval 的 2 倍
            time.Sleep(interval * 2)
            atomic.AddInt32(&currentConcurrent, -1)
        }))
        time.Sleep(interval*8 + tick*4)
        t.Logf("总执行次数: %d, 最大并发: %d", atomic.LoadInt32(&totalRuns), atomic.LoadInt32(&maxConcurrent))
        if atomic.LoadInt32(&maxConcurrent) > 1 {
            t.Errorf("重复任务并发重入: 最大并发数 %d, 预期 1", atomic.LoadInt32(&maxConcurrent))
        }
    })
}

// ==================== ScheduleCustom 自定义调度 ====================

func TestTimerWheel_ScheduleCustom(t *testing.T) {
    t.Parallel()
    t.Run("ReturnZeroStops", func(t *testing.T) {
        t.Parallel()
        tick := 50 * time.Millisecond
        var et execTracker
        w := newTestWheel(t, tick, et.executor)
        callCount := 0
        w.ScheduleCustom(func(now time.Time) time.Time {
            callCount++
            if callCount > 3 {
                return time.Time{}
            }
            return now.Add(80 * time.Millisecond)
        }, FuncTask(func() {}))
        time.Sleep(500 * time.Millisecond)
        if count := atomic.LoadInt32(&et.count); count > 3 {
            t.Errorf("任务执行次数过多: %d, 预期最多 3 次", count)
        }
    })

    t.Run("ImmediateThenReschedule", func(t *testing.T) {
        t.Parallel()
        // 首次立即执行后仍可继续重调度
        tick := 50 * time.Millisecond
        var et execTracker
        w := newTestWheel(t, tick, et.executor)
        callCount := 0
        w.ScheduleCustom(func(now time.Time) time.Time {
            callCount++
            if callCount == 1 {
                return now
            }
            if callCount <= 3 {
                return now.Add(tick)
            }
            return time.Time{}
        }, FuncTask(func() {}))
        time.Sleep(tick*4 + 100*time.Millisecond)
        if got := atomic.LoadInt32(&et.count); got != 3 {
            t.Errorf("首次立即执行后的重调度次数错误: %d, 预期 3", got)
        }
        assertTaskCount(t, w, 0, "立即执行链路结束后计数错误")
    })

    t.Run("NoStackOverflow", func(t *testing.T) {
        t.Parallel()
        // 持续返回过去时间不会栈溢出
        var et execTracker
        tick := 50 * time.Millisecond
        w := newTestWheel(t, tick, et.executor)
        callCount := 0
        w.ScheduleCustom(func(now time.Time) time.Time {
            callCount++
            if callCount > 2000 {
                return time.Time{}
            }
            return now.Add(-time.Millisecond)
        }, FuncTask(func() {}))
        time.Sleep(tick * 3)
        if atomic.LoadInt32(&et.count) == 0 {
            t.Error("持续过去时间的任务未被执行")
        }
        time.Sleep(tick * 2)
        assertTaskCount(t, w, 0, "持续过去时间任务结束后计数错误")
    })

    t.Run("MaxImmediateReschedule_ReleasesTaskCount", func(t *testing.T) {
        t.Parallel()
        // 超过即时重调度上限后的计数处理
        tick := 20 * time.Millisecond
        var runCount, scheduleCalls int32
        var taskDone uint32
        done := make(chan struct{}, 1)
        w := NewTimerWheel(
            WithTickInterval(tick),
            WithExecutor(func(task Task) { atomic.AddInt32(&runCount, 1); task.Run() }),
        )
        defer w.Stop()
        const stopAfter = 1300
        w.ScheduleCustom(func(now time.Time) time.Time {
            call := atomic.AddInt32(&scheduleCalls, 1)
            if call > stopAfter {
                atomic.StoreUint32(&taskDone, 1)
                select {
                case done <- struct{}{}:
                default:
                }
                return time.Time{}
            }
            return now.Add(-time.Millisecond)
        }, FuncTask(func() {}))
        waitFor(t, done, 3*time.Second, "任务未在预期时间内终止，可能未触发强制推迟或出现死循环")
        time.Sleep(tick * 4)
        if atomic.LoadUint32(&taskDone) == 0 {
            t.Fatal("任务未按预期结束")
        }
        if got := atomic.LoadInt32(&runCount); got < 1025 {
            t.Fatalf("任务执行次数过少: %d，预期至少超过 maxImmediateReschedule", got)
        }
        if got := atomic.LoadInt32(&scheduleCalls); got < 1026 {
            t.Fatalf("调度函数调用次数过少: %d，预期至少超过 maxImmediateReschedule", got)
        }
        assertTaskCount(t, w, 0, "任务结束后计数错误")
        time.Sleep(tick * 2)
        assertTaskCount(t, w, 0, "二次检查计数错误")
    })

    t.Run("Terminate_LastRunCompleted_Regression", func(t *testing.T) {
        t.Parallel()
        // 验证 schedule 返回零值终止时，最后一次执行必须真正完成（回归）
        tick := 50 * time.Millisecond
        var execCount int32
        var lastRunCompleted uint32
        done := make(chan struct{}, 1)

        w := NewTimerWheel(WithTickInterval(tick))
        defer w.Stop()

        w.ScheduleCustom(func(now time.Time) time.Time {
            if atomic.LoadInt32(&execCount) >= 3 {
                return time.Time{}
            }
            return now.Add(tick)
        }, FuncTask(func() {
            count := atomic.AddInt32(&execCount, 1)
            atomic.StoreUint32(&lastRunCompleted, 1)
            if count == 3 {
                select {
                case done <- struct{}{}:
                default:
                }
            }
        }))

        waitFor(t, done, 3*time.Second, "未在预期时间内收到第 3 次执行通知")

        time.Sleep(tick * 4)
        if got := atomic.LoadInt32(&execCount); got != 3 {
            t.Errorf("执行次数错误: %d, 预期 3", got)
        }
        if atomic.LoadUint32(&lastRunCompleted) == 0 {
            t.Error("最后一次执行的 fn 未真正完成")
        }
        assertTaskCount(t, w, 0, "终止后计数错误")
    })
}

// ==================== Cancel 取消任务 ====================

func TestTimerWheel_Cancel(t *testing.T) {
    t.Parallel()
    t.Run("BeforeExecution", func(t *testing.T) {
        t.Parallel()
        // 取消延迟任务（含幂等性验证）
        var et execTracker
        w := newTestWheel(t, 50*time.Millisecond, et.executor)
        handle := w.Schedule(200*time.Millisecond, FuncTask(func() {}))
        time.Sleep(50 * time.Millisecond)
        handle.Cancel()
        // 等待足够长让原任务本应执行
        time.Sleep(300 * time.Millisecond)
        if atomic.LoadInt32(&et.count) > 0 {
            t.Error("已取消的任务不应该执行")
        }
        // 幂等性：多次取消不应该 panic
        for i := 0; i < 10; i++ {
            handle.Cancel()
        }
    })

    t.Run("ImmediateTaskCountDecrement", func(t *testing.T) {
        t.Parallel()
        // 取消后 taskCount 立即归零，不需要等到槽位被处理
        w := newTestWheel(t, 50*time.Millisecond, nil)
        const n = 100
        handles := make([]TaskHandle, n)
        for i := range handles {
            handles[i] = w.Schedule(10*time.Second, FuncTask(func() {}))
        }
        assertTaskCount(t, w, n, "入槽后计数错误")
        for _, h := range handles {
            h.Cancel()
        }
        assertTaskCount(t, w, 0, "取消后计数未立即归零")
    })

    // 合并 PreventsCascadeReAdd + PhysicallyRemovesFromSlot
    t.Run("CancelCleanUp", func(t *testing.T) {
        t.Parallel()
        var et execTracker
        w := newTestWheel(t, 50*time.Millisecond, et.executor)
        // 取消的重复任务在 cascade 重调度时不会被重新加入时间轮
        handle := w.ScheduleRepeating(5*time.Second, FuncTask(func() {}))
        handle.Cancel()
        time.Sleep(500 * time.Millisecond)
        if atomic.LoadInt32(&et.count) > 0 {
            t.Errorf("取消的重复任务仍然执行了 %d 次", atomic.LoadInt32(&et.count))
        }
        assertClean(t, w, "取消后")
    })

    t.Run("ConcurrentScheduleCancelRace", func(t *testing.T) {
        t.Parallel()
        // 并发压力测试：多 goroutine 同时调度并立即取消
        // 回归：修复 addTask 中 Cancel 在 cancelled 二次检查后、slotInfo 设置前执行导致槽位泄漏
        tick := 50 * time.Millisecond
        w := newTestWheel(t, tick, nil)

        const rounds = 200
        const perRound = 50

        for r := 0; r < rounds; r++ {
            var wg sync.WaitGroup
            handles := make([]TaskHandle, perRound)

            for i := 0; i < perRound; i++ {
                wg.Add(1)
                go func(idx int) {
                    defer wg.Done()
                    handles[idx] = w.Schedule(10*time.Second, FuncTask(func() {}))
                }(i)
            }
            wg.Wait()

            for i := 0; i < perRound; i++ {
                wg.Add(1)
                go func(idx int) {
                    defer wg.Done()
                    if handles[idx] != nil {
                        handles[idx].Cancel()
                    }
                }(i)
            }
            wg.Wait()
        }

        time.Sleep(tick * 2)
        assertClean(t, w, "并发取消后")
    })
}

// ==================== Stop 停止调度器 ====================

func TestTimerWheel_Stop(t *testing.T) {
    t.Parallel()
    t.Run("Idempotent", func(t *testing.T) {
        t.Parallel()
        var et execTracker
        w := newTestWheel(t, 50*time.Millisecond, et.executor)
        w.Schedule(500*time.Millisecond, FuncTask(func() {}))
        w.Stop()
        time.Sleep(600 * time.Millisecond)
        t.Logf("停止后任务执行次数: %d", atomic.LoadInt32(&et.count))
        // 幂等性：多次停止不应该 panic
        for i := 0; i < 5; i++ {
            w.Stop()
        }
    })

    t.Run("StopThenSchedule", func(t *testing.T) {
        t.Parallel()
        // Stop 后再 Schedule 不会执行（含句柄安全性验证）
        var et execTracker
        w := newTestWheel(t, 50*time.Millisecond, et.executor)
        w.Stop()
        handle := w.Schedule(100*time.Millisecond, FuncTask(func() {}))
        time.Sleep(200 * time.Millisecond)
        if atomic.LoadInt32(&et.count) > 0 {
            t.Error("Stop 后的任务不应该执行")
        }
        // 返回的句柄应该是已取消状态，多次取消不应该 panic
        handle.Cancel()
        handle.Cancel()
    })

    t.Run("NoExecutionAfterStop", func(t *testing.T) {
        t.Parallel()
        var et execTracker
        tick := 50 * time.Millisecond
        w := newTestWheel(t, tick, et.executor)
        w.Schedule(200*time.Millisecond, FuncTask(func() {}))
        w.Schedule(300*time.Millisecond, FuncTask(func() {}))
        w.Schedule(400*time.Millisecond, FuncTask(func() {}))
        // 等待任务进入时间轮但未到期
        time.Sleep(tick)
        w.Stop()
        executedBefore := atomic.LoadInt32(&et.count)
        time.Sleep(500 * time.Millisecond)
        executedAfter := atomic.LoadInt32(&et.count)
        if executedAfter > executedBefore+1 {
            t.Errorf("Stop 后仍有任务执行: 之前 %d, 之后 %d", executedBefore, executedAfter)
        }
        assertTaskCount(t, w, 0, "Stop 后计数未归零")
    })

    t.Run("NoExecutionAfterCascadePick", func(t *testing.T) {
        t.Parallel()
        // 验证高层任务被 cascade 取出后，Stop 也不会继续执行到期任务
        var et execTracker
        w := newTestWheel(t, 50*time.Millisecond, et.executor)
        layer := w.layers[1]
        currentIdx := 0

        dueTime := time.Now().UnixMilli() - 1
        tsk := &task{
            taskIF:   FuncTask(func() {}),
            execTime: dueTime,
        }
        atomic.StoreUint32(&tsk.counted, 1)
        atomic.StoreInt32(&w.taskCount, 1)
        atomic.StoreInt64(&w.lastTime, dueTime)

        tsk.taskID = atomic.AddInt32(&w.nextTaskId, 1)
        layer.cells[currentIdx].mu.Lock()
        layer.cells[currentIdx].tasks = []*task{tsk}
        layer.cells[currentIdx].mu.Unlock()

        // 模拟 Stop 发生在 cascade 已经取出任务之后、真正执行之前
        atomic.StoreUint32(&w.stopped, 1)
        w.processLayerTasks(layer, currentIdx, false)

        if got := atomic.LoadInt32(&et.count); got != 0 {
            t.Fatalf("stopped 后 cascade 任务仍被执行: %d", got)
        }
        assertTaskCount(t, w, 0, "stopped 后 counted 未回收")

        if got := atomic.LoadUint32(&tsk.counted); got != 0 {
            t.Fatalf("任务 counted 标记未清理: %d", got)
        }
    })

    t.Run("AddTaskStopped_ReleasesRescheduleCount", func(t *testing.T) {
        t.Parallel()
        // 验证停止后的计数回收正确
        var et execTracker
        tick := 50 * time.Millisecond
        w := newTestWheel(t, tick, et.executor)
        handle := w.Schedule(500*time.Millisecond, FuncTask(func() {}))
        time.Sleep(tick * 2)
        w.Stop()
        time.Sleep(50 * time.Millisecond)
        assertTaskCount(t, w, 0, "停止后计数未归零")
        handle.Cancel()
    })
}

// ==================== Panic panic 处理 ====================

func TestTimerWheel_Panic(t *testing.T) {
    t.Parallel()
    // 表驱动：不同 panic 来源统一验证
    t.Run("PanicSources", func(t *testing.T) {
        t.Parallel()
        tick := 50 * time.Millisecond
        tests := []struct {
            name  string
            setup func(w *TimerWheel, et *execTracker)
        }{
            {"TaskPanic", func(w *TimerWheel, _ *execTracker) {
                w.Schedule(tick, FuncTask(func() { panic("boom") }))
            }},
            {"SchedulePanic", func(w *TimerWheel, _ *execTracker) {
                callCount := 0
                w.ScheduleCustom(func(now time.Time) time.Time {
                    callCount++
                    if callCount >= 2 {
                        panic("boom")
                    }
                    return now.Add(tick)
                }, FuncTask(func() {}))
            }},
            {"InitialSchedulePanic", func(w *TimerWheel, _ *execTracker) {
                w.ScheduleCustom(func(now time.Time) time.Time { panic("boom") }, FuncTask(func() {}))
            }},
        }
        for _, tt := range tests {
            tt := tt
            t.Run(tt.name, func(t *testing.T) {
                t.Parallel()
                var et execTracker
                w := NewTimerWheel(WithTickInterval(tick), WithPanicHandler(et.panicHandler), WithExecutor(et.executor))
                defer w.Stop()
                tt.setup(w, &et)

                time.Sleep(tick * 6)

                if len(et.panics) == 0 {
                    t.Fatal("panic 处理器未被调用")
                }
                if et.panics[0].Recovered != "boom" {
                    t.Errorf("panic 值错误: %v", et.panics[0].Recovered)
                }
                assertClean(t, w, "panic 后")
            })
        }
    })

    // ScheduleFunc panic 后的资源清理：受害任务终止、健康任务继续运行
    t.Run("ScheduleCustom_PanicInSchedule", func(t *testing.T) {
        t.Parallel()
        ctx := newPanicTestCtx(20 * time.Millisecond)
        defer ctx.W.Stop()

        var victimRuns int32
        var healthyRuns int32
        healthyDone := make(chan struct{}, 1)

        var scheduleCalls int32
        // 受害任务：第3次 schedule 调用时 panic
        ctx.W.ScheduleCustom(func(now time.Time) time.Time {
            call := atomic.AddInt32(&scheduleCalls, 1)
            if call == 3 {
                panic("schedule boom")
            }
            return now.Add(ctx.Tick)
        }, FuncTask(func() { atomic.AddInt32(&victimRuns, 1) }))

        // 健康任务：验证时间轮继续运行
        ctx.W.ScheduleRepeating(ctx.Tick, FuncTask(func() {
            if atomic.AddInt32(&healthyRuns, 1) == 3 {
                select {
                case healthyDone <- struct{}{}:
                default:
                }
            }
        }))

        waitFor(t, ctx.notified, 2*time.Second, "未收到 panicHandler 通知")
        waitFor(t, healthyDone, 2*time.Second, "其他任务未继续执行，说明 schedule panic 影响了时间轮运行")
        time.Sleep(ctx.Tick * 4)

        if got := atomic.LoadInt32(&ctx.panicCalls); got != 1 {
            t.Fatalf("panicHandler 调用次数错误: %d, 预期 1", got)
        }
        if recovered := ctx.handlerRecovered.Load(); recovered != "schedule boom" {
            t.Fatalf("panicHandler 收到的 panic 值错误: %v", recovered)
        }
        if got := atomic.LoadInt32(&victimRuns); got != 2 {
            t.Fatalf("panic 前任务执行次数错误: %d, 预期 2", got)
        }
        assertTaskCount(t, ctx.W, 1, "panic 后计数错误（应仅剩健康重复任务）")
        ctx.W.Stop()
        assertTaskCount(t, ctx.W, 0, "Stop 后计数错误")
    })

    // 首次 schedule 调用 panic 时被正确捕获
    t.Run("ScheduleCustom_InitialSchedulePanic", func(t *testing.T) {
        t.Parallel()
        ctx := newPanicTestCtx(20 * time.Millisecond)
        defer ctx.W.Stop()

        var taskRuns int32
        handle := ctx.W.ScheduleCustom(func(now time.Time) time.Time {
            panic("initial schedule boom")
        }, FuncTask(func() { atomic.AddInt32(&taskRuns, 1) }))

        if handle == nil {
            t.Fatal("ScheduleCustom 返回了 nil handle")
        }
        waitFor(t, ctx.notified, time.Second, "未收到 panicHandler 通知")

        if got := atomic.LoadInt32(&ctx.panicCalls); got != 1 {
            t.Fatalf("panicHandler 调用次数错误: %d, 预期 1", got)
        }
        if recovered := ctx.handlerRecovered.Load(); recovered != "initial schedule boom" {
            t.Fatalf("panicHandler 收到的 panic 值错误: %v", recovered)
        }
        if got := atomic.LoadInt32(&taskRuns); got != 0 {
            t.Fatalf("任务不应执行，实际执行次数: %d", got)
        }
        assertTaskCount(t, ctx.W, 0, "首次 panic 后计数错误")
    })

    // panicHandler 自己再次 panic 时 taskCount 仍归零（任务 panic / 调度 panic 统一验证）
    t.Run("PanicHandler_Repanics", func(t *testing.T) {
        t.Parallel()
        tick := 50 * time.Millisecond

        tests := []struct {
            name  string
            setup func(w *TimerWheel)
        }{
            {"TaskCountReleases", func(w *TimerWheel) {
                w.Schedule(50*time.Millisecond, FuncTask(func() { panic("task boom") }))
            }},
            {"SchedulePanic", func(w *TimerWheel) {
                callCount := 0
                w.ScheduleCustom(func(now time.Time) time.Time {
                    callCount++
                    if callCount >= 2 {
                        panic("schedule boom")
                    }
                    return now.Add(tick)
                }, FuncTask(func() {}))
            }},
        }
        for _, tt := range tests {
            tt := tt
            t.Run(tt.name, func(t *testing.T) {
                t.Parallel()
                w := NewTimerWheel(
                    WithTickInterval(tick),
                    WithPanicHandler(func(task Task, recovered any) { panic("handler boom") }),
                    WithExecutor(func(task Task) { task.Run() }),
                )
                defer w.Stop()
                tt.setup(w)
                time.Sleep(tick * 6)
                assertTaskCount(t, w, 0, "panicHandler 二次 panic 后计数未归零")
            })
        }
    })

    // 自定义 executor 自身 panic 时 taskCount 仍归零（直接 panic / Run 后 panic 统一验证）
    t.Run("ExecutorPanic", func(t *testing.T) {
        t.Parallel()
        tick := 50 * time.Millisecond

        tests := []struct {
            name     string
            executor func(Task)
            taskBody func(runs *int32) Task
            assert   func(t *testing.T, w *TimerWheel, runs *int32)
        }{
            {"TaskCountReleases",
                func(task Task) { panic("executor boom") },
                func(_ *int32) Task { return FuncTask(func() {}) },
                func(t *testing.T, w *TimerWheel, _ *int32) {
                    assertTaskCount(t, w, 0, "executor panic 后计数未归零")
                },
            },
            {"AfterRun_OnDoneOnlyOnce",
                func(task Task) { task.Run(); panic("executor boom after run") },
                func(runs *int32) Task { return FuncTask(func() { atomic.AddInt32(runs, 1) }) },
                func(t *testing.T, w *TimerWheel, runs *int32) {
                    if got := atomic.LoadInt32(runs); got != 1 {
                        t.Fatalf("任务执行次数错误: %d, 预期 1", got)
                    }
                    assertTaskCount(t, w, 0, "executor 在 task.Run 后 panic 导致计数错误")
                    time.Sleep(tick)
                    assertTaskCount(t, w, 0, "二次检查计数错误")
                },
            },
        }
        for _, tt := range tests {
            tt := tt
            t.Run(tt.name, func(t *testing.T) {
                t.Parallel()
                var runs int32
                w := NewTimerWheel(WithTickInterval(tick), WithExecutor(tt.executor))
                defer w.Stop()
                w.Schedule(50*time.Millisecond, tt.taskBody(&runs))
                time.Sleep(tick * 6)
                tt.assert(t, w, &runs)
            })
        }
    })
}

// ==================== Concurrent 并发测试 ====================

func TestTimerWheel_Concurrent(t *testing.T) {
    t.Parallel()
    t.Run("ConcurrentSchedule", func(t *testing.T) {
        t.Parallel()
        var et execTracker
        tick := 50 * time.Millisecond
        w := newTestWheel(t, tick, et.executor)

        var wg sync.WaitGroup
        numGoroutines := 100
        numTasksPerGoroutine := 10

        for i := 0; i < numGoroutines; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for j := 0; j < numTasksPerGoroutine; j++ {
                    w.Schedule(time.Duration(j*10+100)*time.Millisecond, FuncTask(func() {}))
                }
            }()
        }
        wg.Wait()

        time.Sleep(time.Duration(numTasksPerGoroutine*10+200)*time.Millisecond + tick*2)

        expectedTasks := numGoroutines * numTasksPerGoroutine
        count := atomic.LoadInt32(&et.count)
        if count < int32(expectedTasks)*9/10 {
            t.Errorf("并发调度执行任务数不足: %d, 预期约 %d", count, expectedTasks)
        }
    })
}

// ==================== Defense 防御性测试 ====================

func TestTimerWheel_Defense(t *testing.T) {
    t.Parallel()
    t.Run("NilTask", func(t *testing.T) {
        t.Parallel()
        w := newTestWheel(t, 50*time.Millisecond, nil)
        // Schedule(nil) / ScheduleRepeating(nil) / ScheduleCustom(nil, ...) 均应返回可取消句柄
        w.Schedule(100*time.Millisecond, nil).Cancel()
        w.ScheduleRepeating(100*time.Millisecond, nil).Cancel()
        w.ScheduleCustom(nil, FuncTask(func() {})).Cancel()
        w.ScheduleCustom(func(now time.Time) time.Time { return now.Add(time.Second) }, nil).Cancel()
        assertTaskCount(t, w, 0, "nil task 输入导致计数非零")
    })

    t.Run("FuncTask", func(t *testing.T) {
        t.Parallel()
        var executed uint32
        FuncTask(func() { atomic.StoreUint32(&executed, 1) }).Run()
        if atomic.LoadUint32(&executed) == 0 {
            t.Error("FuncTask.Run() 未执行")
        }
        // nil FuncTask 不应该 panic
        var nilFn FuncTask
        nilFn.Run()
    })
}

// ==================== Slice 存储回归测试 ====================

// TestTimerCell_RemoveTask_SwapWithLast 验证 timerCell.removeTask 的 swap-with-last 正确性
func TestTimerCell_RemoveTask_SwapWithLast(t *testing.T) {
    t.Parallel()
    // 构造 5 个任务，删除中间的第 3 个（taskID=3），验证顺序无关
    var c timerCell
    tasks := make([]*task, 5)
    for i := 0; i < 5; i++ {
        tasks[i] = &task{taskID: int32(i + 1)}
        c.addTask(tasks[i])
    }
    if len(c.tasks) != 5 {
        t.Fatalf("添加后长度错误: %d, 预期 5", len(c.tasks))
    }

    // 删除 taskID=3（中间元素）
    if !c.removeTask(3) {
        t.Fatal("removeTask(3) 返回 false")
    }
    if len(c.tasks) != 4 {
        t.Fatalf("删除后长度错误: %d, 预期 4", len(c.tasks))
    }

    // swap-with-last: taskID=3 的位置被 taskID=5 替换
    found := make(map[int32]bool)
    for _, tsk := range c.tasks {
        if tsk == nil {
            t.Fatal("切片中存在 nil 元素")
        }
        found[tsk.taskID] = true
    }
    for _, id := range []int32{1, 2, 4, 5} {
        if !found[id] {
            t.Errorf("taskID=%d 不在切片中", id)
        }
    }
    if found[3] {
        t.Error("taskID=3 仍在切片中")
    }

    // 删除不存在的 taskID 应返回 false
    if c.removeTask(999) {
        t.Error("删除不存在的 taskID 应返回 false")
    }
    if len(c.tasks) != 4 {
        t.Fatalf("删除不存在的 taskID 后长度变化: %d", len(c.tasks))
    }

    // 删除唯一元素
    var c2 timerCell
    tsk := &task{taskID: 42}
    c2.addTask(tsk)
    if !c2.removeTask(42) {
        t.Fatal("删除唯一元素失败")
    }
    if len(c2.tasks) != 0 {
        t.Fatalf("删除唯一元素后长度: %d, 预期 0", len(c2.tasks))
    }
}

// TestTimerWheel_SliceStorage_NoLeak 验证切片存储在大量 add/remove 后无泄漏
func TestTimerWheel_SliceStorage_NoLeak(t *testing.T) {
    t.Parallel()
    tick := 50 * time.Millisecond
    w := newTestWheel(t, tick, nil)

    // 批量调度并立即取消，验证切片和 taskCount 都干净
    const n = 500
    handles := make([]TaskHandle, n)
    for i := 0; i < n; i++ {
        handles[i] = w.Schedule(10*time.Second, FuncTask(func() {}))
    }
    assertTaskCount(t, w, n, "批量调度后计数错误")

    for _, h := range handles {
        h.Cancel()
    }
    assertTaskCount(t, w, 0, "全部取消后计数错误")
    assertClean(t, w, "批量取消后")
}

// TestTimerWheel_ProcessLayerTasks_DrainsSlice 验证 processLayerTasks 正确清空切片
func TestTimerWheel_ProcessLayerTasks_DrainsSlice(t *testing.T) {
    t.Parallel()
    var et execTracker
    w := newTestWheel(t, 50*time.Millisecond, et.executor)

    // 同时调度多个零延迟任务，都会落入同一个槽位并立即执行
    const n = 20
    for i := 0; i < n; i++ {
        w.Schedule(0, FuncTask(func() {}))
    }

    // 零延迟任务同步执行，等待完成
    time.Sleep(200 * time.Millisecond)

    if got := atomic.LoadInt32(&et.count); got != n {
        t.Errorf("零延迟批量任务执行数错误: %d, 预期 %d", got, n)
    }
    assertTaskCount(t, w, 0, "零延迟批量任务执行后计数错误")
    assertClean(t, w, "零延迟批量任务执行后")
}

// TestTimerWheel_ConcurrentCancel_SliceIntegrity 验证并发取消时切片完整性
func TestTimerWheel_ConcurrentCancel_SliceIntegrity(t *testing.T) {
    t.Parallel()
    tick := 100 * time.Millisecond
    w := newTestWheel(t, tick, nil)

    const n = 200
    handles := make([]TaskHandle, n)
    var wg sync.WaitGroup

    // 并发调度
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            handles[idx] = w.Schedule(10*time.Second, FuncTask(func() {}))
        }(i)
    }
    wg.Wait()

    // 并发取消
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            if handles[idx] != nil {
                handles[idx].Cancel()
            }
        }(i)
    }
    wg.Wait()

    time.Sleep(tick * 2)
    assertClean(t, w, "并发取消后")
}

// ==================== 基准测试 ====================

func BenchmarkTimerWheel(b *testing.B) {
    b.Run("Schedule", func(b *testing.B) {
        w := NewTimerWheel()
        defer w.Stop()
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            w.Schedule(time.Millisecond*100, FuncTask(func() {}))
        }
    })
    b.Run("ScheduleParallel", func(b *testing.B) {
        w := NewTimerWheel()
        defer w.Stop()
        b.RunParallel(func(pb *testing.PB) {
            for pb.Next() {
                w.Schedule(time.Millisecond*100, FuncTask(func() {}))
            }
        })
    })
    b.Run("Cancel", func(b *testing.B) {
        w := NewTimerWheel()
        defer w.Stop()
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            w.Schedule(time.Second, FuncTask(func() {})).Cancel()
        }
    })
    b.Run("Repeating", func(b *testing.B) {
        w := NewTimerWheel()
        defer w.Stop()
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            w.ScheduleRepeating(time.Millisecond*100, FuncTask(func() {}))
        }
    })
}

// ==================== removeTask 线性插值优化基准测试 ====================

// benchmarkRemoveTask 用于对比优化前后的删除性能
// 有序场景：任务按 ID 递增添加
func benchmarkRemoveOrdered(b *testing.B, size int) {
    var c timerCell
    // 按顺序添加任务
    for i := 0; i < size; i++ {
        c.tasks = append(c.tasks, &task{taskID: int32(i + 1)})
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // 循环删除不同位置的任务
        idx := i % size
        c.mu.Lock()
        c.removeTask(int32(idx + 1))
        // 重新添加以保持切片大小
        c.tasks = append(c.tasks, &task{taskID: int32(idx + 1)})
        c.mu.Unlock()
    }
}

// 弱有序场景：模拟 swap-with-last 导致的部分乱序
func benchmarkRemoveWeaklyOrdered(b *testing.B, size int) {
    var c timerCell
    for i := 0; i < size; i++ {
        c.tasks = append(c.tasks, &task{taskID: int32(i + 1)})
    }

    // 模拟 20% 的 swap-with-last 删除
    rng := make([]int, size)
    for i := range rng {
        rng[i] = i
    }
    for i := 0; i < size/5; i++ {
        delIdx := rng[i]
        lastIdx := len(c.tasks) - 1
        if delIdx != lastIdx {
            c.tasks[delIdx] = c.tasks[lastIdx]
        }
        c.tasks = c.tasks[:lastIdx]
        c.tasks = append(c.tasks, &task{taskID: int32(size + i + 1)})
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        idx := i % size
        c.mu.Lock()
        c.removeTask(int32(idx + 1))
        c.tasks = append(c.tasks, &task{taskID: int32(idx + 1)})
        c.mu.Unlock()
    }
}

// 乱序场景：模拟大量 swap-with-last 导致的随机顺序
func benchmarkRemoveRandomOrder(b *testing.B, size int) {
    var c timerCell
    for i := 0; i < size; i++ {
        c.tasks = append(c.tasks, &task{taskID: int32(i + 1)})
    }

    // 模拟 50% 的 swap-with-last 删除
    for i := 0; i < size/2; i++ {
        delIdx := i
        lastIdx := len(c.tasks) - 1
        if delIdx != lastIdx {
            c.tasks[delIdx] = c.tasks[lastIdx]
        }
        c.tasks = c.tasks[:lastIdx]
        c.tasks = append(c.tasks, &task{taskID: int32(size + i + 1)})
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        idx := i % size
        c.mu.Lock()
        c.removeTask(int32(idx + 1))
        c.tasks = append(c.tasks, &task{taskID: int32(idx + 1)})
        c.mu.Unlock()
    }
}

func BenchmarkTimerCell_RemoveTask(b *testing.B) {
    sizes := []int{1, 5, 10, 50, 100, 500}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("ordered/size=%d", size), func(b *testing.B) {
            benchmarkRemoveOrdered(b, size)
        })
        b.Run(fmt.Sprintf("weakly_ordered/size=%d", size), func(b *testing.B) {
            benchmarkRemoveWeaklyOrdered(b, size)
        })
        b.Run(fmt.Sprintf("random/size=%d", size), func(b *testing.B) {
            benchmarkRemoveRandomOrder(b, size)
        })
    }
}
