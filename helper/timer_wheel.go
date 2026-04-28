package helper

import (
    "sync"
    "sync/atomic"
    "time"
)

// Option 配置选项
type Option func(*TimerWheel)

// WithTickInterval 设置 tick 间隔
func WithTickInterval(d time.Duration) Option {
    return func(w *TimerWheel) {
        if d >= 5*time.Millisecond && d <= 10*time.Second {
            w.tickInterval = d
        }
    }
}

// WithExecutor 设置自定义执行器
func WithExecutor(fn func(Task)) Option {
    return func(w *TimerWheel) {
        if fn != nil {
            w.executor = fn
        }
    }
}

// WithPanicHandler 设置 panic 处理器
func WithPanicHandler(fn func(task Task, recovered any)) Option {
    return func(w *TimerWheel) {
        w.panicHandler = fn
    }
}

// TimerWheel 时间轮调度器
type TimerWheel struct {
    tickInterval time.Duration
    tickMs       int64 // tickInterval 的毫秒数，构造时缓存避免热路径重复计算
    executor     func(Task)
    panicHandler func(task Task, recovered any)

    // 时间轮层数据
    layers     []*wheelLayer
    lastTime   int64 // 使用原子访问避免并发竞争
    taskCount  int32
    nextTaskId int32 // 自增任务 ID，用于 map O(1) 删除

    stop    chan struct{}
    stopped uint32 // 0=running, 1=stopped
}

type wheelLayer struct {
    level         int
    interval      int64 // 每格时间（毫秒）
    totalInterval int64 // 整层时间（毫秒）
    cells         [wheelSize]timerCell
    idx           int32 // 使用原子访问避免并发竞争
}

// timerCell 槽位单元，使用切片存储任务
// 删除策略：大切片用线性插值预估位置加速查找，小切片或 miss 时回退遍历，找到后用 swap-with-last 删除
type timerCell struct {
    mu    sync.Mutex
    tasks []*task
}

// addTask 添加任务到切片
func (c *timerCell) addTask(t *task) {
    c.tasks = append(c.tasks, t)
}

// doRemove 执行 swap-with-last 删除
func (c *timerCell) doRemove(idx int) {
    lastIdx := len(c.tasks) - 1
    if idx != lastIdx {
        c.tasks[idx] = c.tasks[lastIdx]
    }
    c.tasks[lastIdx] = nil
    c.tasks = c.tasks[:lastIdx]
}

// removeTask 删除指定 taskID 的任务
// 使用线性插值预估位置加速查找，miss 时回退遍历
func (c *timerCell) removeTask(taskID int32) bool {
    n := len(c.tasks)
    if n == 0 {
        return false
    }

    // 小切片直接遍历，避免插值计算开销
    if n <= 32 {
        for i := range c.tasks {
            if c.tasks[i].taskID == taskID {
                c.doRemove(i)
                return true
            }
        }
        return false
    }

    // 线性插值预估位置
    firstID := c.tasks[0].taskID
    lastID := c.tasks[n-1].taskID

    if firstID != lastID {
        estIdx := int((taskID - firstID) * int32(n-1) / (lastID - firstID))
        if estIdx >= 0 && estIdx < n && c.tasks[estIdx].taskID == taskID {
            c.doRemove(estIdx)
            return true
        }
    }

    // 回退遍历
    for i := range c.tasks {
        if c.tasks[i].taskID == taskID {
            c.doRemove(i)
            return true
        }
    }
    return false
}

// NewTimerWheel 创建时间轮
func NewTimerWheel(opts ...Option) *TimerWheel {
    tick := 100 * time.Millisecond
    w := &TimerWheel{
        tickInterval: tick,
        executor:     defaultExecutor,
        stop:         make(chan struct{}),
    }
    atomic.StoreInt64(&w.lastTime, time.Now().UnixMilli())

    for _, opt := range opts {
        opt(w)
    }

    // 缓存 tickInterval 的毫秒值，避免热路径重复计算
    w.tickMs = w.tickInterval.Milliseconds()

    // 初始化 8 层时间轮，默认 tick=100ms 时覆盖到约 115.7 天
    // 层级覆盖范围：
    // - 第0层：1秒 (100ms * 10)
    // - 第1层：10秒
    // - 第2层：100秒 (~1.67分钟)
    // - 第3层：1000秒 (~16.7分钟)
    // - 第4层：10000秒 (~2.78小时)
    // - 第5层：100000秒 (~27.8小时)
    // - 第6层：1000000秒 (~11.6天)
    // - 第7层：10000000秒 (~115.7天)
    interval := w.tickMs
    for i := 0; i < 8; i++ {
        layer := &wheelLayer{
            level:         i,
            interval:      interval,
            totalInterval: interval * wheelSize,
        }
        for j := 0; j < wheelSize; j++ {
            layer.cells[j].tasks = nil
        }
        w.layers = append(w.layers, layer)
        interval *= wheelSize
    }

    go w.run()

    return w
}

// Schedule 调度单次延迟任务
func (w *TimerWheel) Schedule(delay time.Duration, taskObj Task) TaskHandle {
    // nil task 防御：返回已取消句柄，不增加 taskCount
    if taskObj == nil {
        return newCancelledHandle()
    }

    if w.isStopped() {
        return newCancelledHandle()
    }

    if delay < 0 {
        delay = 0
    }

    now := time.Now().UnixMilli()
    execTime := now + delay.Milliseconds()
    t := &task{slotInfo: -1}
    t.wheel = w
    t.taskIF = taskObj
    t.execTime = execTime

    // 二次检查 stopped，与 Stop 的原子操作协调
    if w.isStopped() {
        return newCancelledHandle()
    }

    w.addTask(t, execTime, now, false)

    return t
}

// ScheduleRepeating 调度重复任务
func (w *TimerWheel) ScheduleRepeating(interval time.Duration, taskObj Task) TaskHandle {
    // nil task 防御：返回已取消句柄，不增加 taskCount
    if taskObj == nil {
        return newCancelledHandle()
    }

    if w.isStopped() {
        return newCancelledHandle()
    }

    if interval < w.tickInterval {
        interval = w.tickInterval
    }

    now := time.Now().UnixMilli()
    intervalMs := interval.Milliseconds()
    execTime := now + intervalMs
    t := &task{slotInfo: -1}
    t.wheel = w
    t.taskIF = taskObj
    t.interval = intervalMs
    t.execTime = execTime

    // 二次检查 stopped
    if w.isStopped() {
        return newCancelledHandle()
    }

    w.addTask(t, execTime, now, false)

    return t
}

// ScheduleCustom 自定义调度
func (w *TimerWheel) ScheduleCustom(schedule ScheduleFunc, taskObj Task) TaskHandle {
    // nil task 或 nil schedule 防御：返回已取消句柄，不增加 taskCount
    if taskObj == nil || schedule == nil {
        return newCancelledHandle()
    }

    if w.isStopped() {
        return newCancelledHandle()
    }

    nowTime := time.Now()
    t := &task{slotInfo: -1}
    t.wheel = w
    t.taskIF = taskObj
    t.schedule = schedule

    firstTime, panicked := w.safeScheduleCall(t, nowTime)
    if panicked || firstTime.IsZero() {
        // schedule panic 或返回零值，返回已取消句柄
        return newCancelledHandle()
    }

    execTime := firstTime.UnixMilli()
    t.execTime = execTime

    // 二次检查 stopped，与 Stop 的原子操作协调
    if w.isStopped() {
        return newCancelledHandle()
    }

    w.addTask(t, execTime, nowTime.UnixMilli(), false)

    return t
}

// Stop 停止调度器
func (w *TimerWheel) Stop() {
    if atomic.SwapUint32(&w.stopped, 1) != 0 {
        return
    }

    for _, layer := range w.layers {
        for i := 0; i < len(layer.cells); i++ {
            cell := &layer.cells[i]
            cell.mu.Lock()
            tasks := cell.tasks
            cell.tasks = nil
            cell.mu.Unlock()
            for _, t := range tasks {
                if atomic.SwapUint32(&t.counted, 0) != 0 {
                    atomic.AddInt32(&w.taskCount, -1)
                }
            }
        }
    }

    close(w.stop)
}
