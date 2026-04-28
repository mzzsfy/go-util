package helper

import (
    "sync/atomic"
    "time"
)

// defaultExecutor 默认执行器，新 goroutine 执行任务
func defaultExecutor(t Task) {
    go t.Run()
}

// newCancelledHandle 创建一个已取消的任务句柄（单例复用）
func newCancelledHandle() *task {
    return &cancelledTask
}

// addTask 添加任务到时间轮
// now: 当前时间戳（毫秒），避免重复调用 time.Now()
func (w *TimerWheel) addTask(t *task, execTime int64, now int64, _ bool) {
    if w.isSkipped(t) {
        return
    }

    // 限制即时重调度次数，避免自定义调度持续返回过去时间时无限递归
    const maxImmediateReschedule = 1024
    countAdded := atomic.LoadUint32(&t.counted) != 0

    t.execTime = execTime
    delay := execTime - now
    if delay <= 0 {
        if t.interval == 0 && t.schedule == nil {
            w.executeTask(t)
            return
        }

        if atomic.AddUint32(&t.immediateRuns, 1) > maxImmediateReschedule {
            // 超过即时执行上限，强制推迟一个 tick，重置计数
            atomic.StoreUint32(&t.immediateRuns, 0)
            delay = w.tickMs
            execTime = now + delay
            t.execTime = execTime
        } else {
            // 重复任务的立即执行：执行完成后统一在 safeTask 中处理重调度
            w.executeTask(t)
            return
        }
    } else {
        atomic.StoreUint32(&t.immediateRuns, 0)
    }

    // 找到目标层级和槽位
    var targetLayer *wheelLayer
    var cellIdx int
    for _, layer := range w.layers {
        if delay < layer.totalInterval {
            currentIdx := int(atomic.LoadInt32(&layer.idx))
            var steps int
            if layer.level == 0 {
                steps = int((delay + layer.interval - 1) / layer.interval)
                if steps < 1 {
                    steps = 1
                }
            } else {
                steps = int(delay / layer.interval)
                if steps < 1 {
                    steps = 1
                }
            }
            cellIdx = (currentIdx + steps - 1) % wheelSize
            targetLayer = layer
            break
        }
    }
    if targetLayer == nil {
        targetLayer = w.layers[len(w.layers)-1]
        cellIdx = wheelSize - 1
    }

    // 直接入槽
    // taskID 预分配：在锁外完成原子自增，减少临界区持有时间
    if t.taskID == 0 {
        t.taskID = atomic.AddInt32(&w.nextTaskId, 1)
    }
    cell := &targetLayer.cells[cellIdx]
    cell.mu.Lock()
    if w.isSkipped(t) {
        cell.mu.Unlock()
        return
    }
    if !countAdded {
        atomic.AddInt32(&w.taskCount, 1)
        atomic.StoreUint32(&t.counted, 1)
    }
    cell.addTask(t)
    atomic.StoreInt32(&t.slotInfo, int32(targetLayer.level*wheelSize+cellIdx))
    // 补偿检查：Cancel 可能在二次 cancelled 检查后、入槽完成前设置了 cancelled，
    // 导致 removeFromSlot 因 slotInfo 尚为 -1 而跳过物理删除。
    // 此处持锁，可安全回滚刚完成的入槽操作。
    if atomic.LoadUint32(&t.cancelled) != 0 {
        cell.removeTask(t.taskID)
        atomic.StoreInt32(&t.slotInfo, -1)
        if atomic.SwapUint32(&t.counted, 0) != 0 {
            atomic.AddInt32(&w.taskCount, -1)
        }
    }
    cell.mu.Unlock()
}

func (w *TimerWheel) removeFromSlot(t *task) {
    slotInfo := atomic.SwapInt32(&t.slotInfo, -1)
    if slotInfo < 0 {
        return
    }
    layerIdx := int(slotInfo / wheelSize)
    cellIdx := int(slotInfo % wheelSize)
    if layerIdx < 0 || layerIdx >= len(w.layers) || cellIdx < 0 || cellIdx >= wheelSize {
        return
    }
    layer := w.layers[layerIdx]
    cell := &layer.cells[cellIdx]
    cell.mu.Lock()
    cell.removeTask(t.taskID)
    cell.mu.Unlock()
}

// safeScheduleCall 安全调用 schedule 函数，捕获 panic
// 返回零值表示终止调度（包括 panic 情况）
func (w *TimerWheel) safeScheduleCall(t *task, scheduleBase time.Time) (nextTime time.Time, panicked bool) {
    defer func() {
        if r := recover(); r != nil {
            panicked = true
            nextTime = time.Time{}
            // panicHandler 自身 panic 不能破坏调度状态
            if w.panicHandler != nil {
                func() {
                    defer func() { recover() }()
                    w.panicHandler(t, r)
                }()
            }
        }
    }()
    nextTime = t.schedule(scheduleBase)
    return nextTime, false
}

func (w *TimerWheel) executeTask(t *task) {
    s := safeTaskPool.Get().(*safeTask)
    s.task = t
    s.handler = w.panicHandler
    s.doneCalled = 0
    // 隔离 executor panic，保证 finish 始终被调用
    defer func() {
        if r := recover(); r != nil {
            // executor 自身 panic（如未调用 task.Run 就 panic）
            if w.panicHandler != nil {
                func() {
                    defer func() { recover() }()
                    w.panicHandler(t, r)
                }()
            }
            s.finish(true)
        }
    }()
    w.executor(s)
}

// rescheduleExecutedTask 处理重复任务执行结束后的重调度或回收
func (w *TimerWheel) rescheduleExecutedTask(t *task, now int64) {
    if w.isSkipped(t) {
        return
    }

    if t.interval > 0 {
        nextExecTime := t.execTime + t.interval
        t.execTime = nextExecTime
        w.addTask(t, nextExecTime, now, true)
        return
    }

    if t.schedule != nil {
        scheduleBase := time.UnixMilli(t.execTime)
        nextTime, _ := w.safeScheduleCall(t, scheduleBase)
        if !nextTime.IsZero() {
            nextExecTime := nextTime.UnixMilli()
            t.execTime = nextExecTime
            w.addTask(t, nextExecTime, now, true)
            return
        }
    }

    if atomic.SwapUint32(&t.counted, 0) != 0 {
        atomic.AddInt32(&w.taskCount, -1)
    }
}

// isStopped 检查定时轮是否已停止
func (w *TimerWheel) isStopped() bool {
    return atomic.LoadUint32(&w.stopped) != 0
}

// isSkipped 检查时间轮是否已停止或任务已取消，若任一成立则回收 taskCount
// 返回 true 表示任务应被跳过
func (w *TimerWheel) isSkipped(t *task) bool {
    if w.isStopped() || atomic.LoadUint32(&t.cancelled) != 0 {
        if atomic.SwapUint32(&t.counted, 0) != 0 {
            atomic.AddInt32(&w.taskCount, -1)
        }
        return true
    }
    return false
}

func (w *TimerWheel) run() {
    ticker := time.NewTicker(w.tickInterval)
    defer ticker.Stop()

    for {
        select {
        case <-w.stop:
            return
        case <-ticker.C:
            w.tick()
        }
    }
}

func (w *TimerWheel) tick() {
    now := time.Now().UnixMilli()
    elapsed := now - atomic.LoadInt64(&w.lastTime)

    // 处理时间回调（elapsed < 0）：更新 lastTime 并跳过本次 tick
    // 这种情况可能发生在系统时间被手动调整或 NTP 同步时
    if elapsed < 0 {
        atomic.StoreInt64(&w.lastTime, now)
        return
    }

    tickMs := w.tickMs
    if elapsed < tickMs {
        return
    }

    // 限制单次追赶轮数：
    // 1. 避免 32 位平台 int 转换溢出
    // 2. 避免系统时间大幅前跳时长时间阻塞调度 goroutine
    const maxCatchUpRounds int64 = 80

    rounds := elapsed / tickMs
    if rounds > maxCatchUpRounds {
        atomic.AddInt64(&w.lastTime, (rounds-maxCatchUpRounds)*tickMs)
        rounds = maxCatchUpRounds
    }
    for i := int64(0); i < rounds; i++ {
        w.advance()
    }
}

func (w *TimerWheel) advance() {
    atomic.AddInt64(&w.lastTime, w.tickMs)

    layer := w.layers[0]
    currentIdx := w.advanceLayerIndex(layer)

    // 处理当前槽位的任务
    w.processLayerTasks(layer, currentIdx, true)

    if currentIdx == wheelSize-1 {
        // 从上层下沉任务
        if len(w.layers) > 1 {
            w.cascade(1)
        }
    }
}

func (w *TimerWheel) cascade(level int) {
    if level >= len(w.layers) {
        return
    }

    layer := w.layers[level]
    currentIdx := w.advanceLayerIndex(layer)

    // 处理当前槽位的任务（cascade 模式：检查到期时间）
    w.processLayerTasks(layer, currentIdx, false)

    if currentIdx == wheelSize-1 {
        w.cascade(level + 1)
    }
}

// advanceLayerIndex 推进层级索引，返回当前索引
func (w *TimerWheel) advanceLayerIndex(layer *wheelLayer) int {
    currentIdx := int(atomic.LoadInt32(&layer.idx))
    // 先推进索引，这样重调度时 addTask 读取的 idx 就是下一个槽位
    // 避免重调度任务落入刚被处理的当前槽位
    newIdx := currentIdx + 1
    if newIdx == wheelSize {
        atomic.StoreInt32(&layer.idx, 0)
    } else {
        atomic.StoreInt32(&layer.idx, int32(newIdx))
    }
    return currentIdx
}

// processLayerTasks 处理层级槽位中的任务
// executeDirectly: true 表示直接执行（advance 模式），false 表示检查到期时间（cascade 模式）
func (w *TimerWheel) processLayerTasks(layer *wheelLayer, currentIdx int, executeDirectly bool) {
    // 取出当前槽位的任务，swap 为新 map 避免复用导致的 map 膨胀
    cell := &layer.cells[currentIdx]
    cell.mu.Lock()
    tasks := cell.tasks
    n := len(tasks)
    if n == 0 {
        cell.mu.Unlock()
        return
    }
    cell.tasks = nil
    cell.mu.Unlock()

    // 使用时间轮逻辑时钟作为 now 基准，避免与 lastTime 混用
    now := atomic.LoadInt64(&w.lastTime)

    for _, t := range tasks {
        // 清除 slotInfo，任务已从槽位取出
        atomic.StoreInt32(&t.slotInfo, -1)

        // 检查 stopped 和 cancelled 状态
        if w.isSkipped(t) {
            continue
        }

        // cascade 模式：检查是否已到期
        if !executeDirectly {
            delay := t.execTime - now
            if delay > 0 {
                // 未到期，下沉到下层，使用逻辑时钟保持一致
                w.addTask(t, t.execTime, now, true)
                continue
            }
        }

        // 执行任务，完成后在 safeTask 中统一判断重调度或回收
        w.executeTask(t)
    }
}
