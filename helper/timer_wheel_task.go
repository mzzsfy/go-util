package helper

import (
    "sync"
    "sync/atomic"
    "time"
)

// wheelSize 每层时间轮的槽位数量
const wheelSize = 10

// TimerWheel 基于分层时间轮的定时任务调度器
// 设计思路：多层不同刻度的轮盘嵌套，单 goroutine 驱动，O(1) 添加/取消，
// 支持 one-shot、固定间隔、cron 表达式等多种调度模式

// Task 可执行任务接口
type Task interface {
    Run()
}

// TaskHandle 任务句柄，仅提供取消能力
type TaskHandle interface {
    Cancel()
}

// ScheduleFunc 调度函数，返回下次执行时间，返回零值表示终止
// 参数 now: 首次调用时为当前时间，后续调用为上次的计划执行时间（非实际执行时间）
type ScheduleFunc func(now time.Time) time.Time

// FuncTask 将 func() 包装为 Task
type FuncTask func()

func (f FuncTask) Run() {
    if f != nil {
        f()
    }
}

// task 内部任务结构
// 字段按 8 字节对齐排列，无 padding
type task struct {
    taskIF        Task         // 任务接口，零分配赋值
    schedule      ScheduleFunc // 自定义调度函数
    execTime      int64        // 绝对执行时间（毫秒）
    interval      int64        // 固定重复间隔（毫秒），>0 时替代 schedule 闭包
    cancelled     uint32       // 0=false, 1=true
    counted       uint32       // 0=false, 1=true; 标记任务是否已占用 taskCount
    immediateRuns uint32       // 连续即时执行次数，避免同步 executor 下深递归
    slotInfo      int32        // 槽位编码: layer*wheelSize+cell, -1=不在槽中
    taskID        int32        // 槽位内唯一任务编号，用于 O(1) 删除
    _             [4]byte      // 对齐填充
    wheel         *TimerWheel  // 所属时间轮，用于 Cancel 时回收计数
}

// 已取消任务单例，避免重复分配
var cancelledTask task

func init() {
    atomic.StoreUint32(&cancelledTask.cancelled, 1)
    atomic.StoreInt32(&cancelledTask.slotInfo, -1)
}

func (t *task) Run() {
    if atomic.LoadUint32(&t.cancelled) != 0 {
        return
    }
    if t.taskIF != nil {
        t.taskIF.Run()
    }
}

func (t *task) Cancel() {
    if !atomic.CompareAndSwapUint32(&t.cancelled, 0, 1) {
        return
    }
    // 立即回收 taskCount 并从槽位物理移除，避免内存堆积
    if w := t.wheel; w != nil {
        t.wheel = nil
        if atomic.SwapUint32(&t.counted, 0) != 0 {
            atomic.AddInt32(&w.taskCount, -1)
        }
        w.removeFromSlot(t)
    }
}

// safeTaskPool 复用 safeTask 结构体，减少堆分配
var safeTaskPool = sync.Pool{
    New: func() any { return &safeTask{} },
}

// safeTask 提供 panic 恢复与执行完成后处理
type safeTask struct {
    task       *task
    handler    func(task Task, recovered any)
    doneCalled uint32
}

// finish 幂等完成处理：归还对象到池，然后执行重调度或回收
func (s *safeTask) finish(panicked bool) {
    if !atomic.CompareAndSwapUint32(&s.doneCalled, 0, 1) {
        return
    }
    t := s.task
    w := t.wheel
    s.task = nil
    s.handler = nil
    safeTaskPool.Put(s)
    if w == nil {
        return
    }
    if panicked {
        if atomic.SwapUint32(&t.counted, 0) != 0 {
            atomic.AddInt32(&w.taskCount, -1)
        }
        return
    }
    if w.isSkipped(t) {
        return
    }
    // 重复任务执行后重调度，单次任务回收计数
    if t.interval > 0 || t.schedule != nil {
        w.rescheduleExecutedTask(t, time.Now().UnixMilli())
        return
    }
    if atomic.SwapUint32(&t.counted, 0) != 0 {
        atomic.AddInt32(&w.taskCount, -1)
    }
}

func (s *safeTask) Run() {
    panicked := false
    defer func() {
        if r := recover(); r != nil {
            panicked = true
            // panicHandler 自身 panic 不能打断清理流程
            if s.handler != nil {
                func() {
                    defer func() { recover() }()
                    s.handler(s.task, r)
                }()
            }
        }
        s.finish(panicked)
    }()
    s.task.Run()
}
