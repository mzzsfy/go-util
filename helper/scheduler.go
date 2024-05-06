package helper

import (
    "fmt"
    "github.com/mzzsfy/go-util/concurrent"
    "math"
    "sync"
    "sync/atomic"
    "time"
)

type joinError struct {
    errs []error
}

func (e *joinError) Error() string {
    var b []byte
    for i, err := range e.errs {
        if i > 0 {
            b = append(b, '\n')
        }
        b = append(b, err.Error()...)
    }
    return string(b)
}

func (e *joinError) Unwrap() []error {
    return e.errs
}

func joinErrs(errs ...error) error {
    n := 0
    for _, err := range errs {
        if err != nil {
            n++
        }
    }
    if n == 0 {
        return nil
    }
    e := &joinError{
        errs: make([]error, 0, n),
    }
    for _, err := range errs {
        if err != nil {
            e.errs = append(e.errs, err)
        }
    }
    return e
}

type Task interface {
    Cancel()
}

type CustomSchedulerTime interface {
    NextTime(time.Time) time.Time
}

type schedulerTask struct {
    fn func()
    //返回下一次运行的时间,小于等于当前时间则不再运行
    nextTime func(time.Time) time.Time
    time     int64
}

func (s schedulerTask) Cancel() {
    s.fn = nil
    s.nextTime = nil
}

type schedulerLayer struct {
    level int
    //每一格的时间,毫秒
    ceilInterval    int64
    allCeilInterval int64
    //每层的时间为检查间隔*10
    adding [10]int32
    cells  [10]concurrent.Queue[schedulerTask]
    //下一次运行的cell
    idx int
}

func (s *schedulerLayer) doNext(scheduler *Scheduler) {
    after := time.Now().UnixMilli() - scheduler.lastTime
    if after < scheduler.interval {
        return
    }
    rounds := 1 + int(after/scheduler.interval)
    for i := 0; i < rounds; i++ {
        s.next(scheduler, time.Now().UnixMilli())
    }
}

func (s *schedulerLayer) next(scheduler *Scheduler, now int64) {
    if s.level == 0 {
        atomic.CompareAndSwapInt32(&scheduler.runLock, 0, 1)
        c := s.cells[s.idx]
        t := time.UnixMilli(now)
        if atomic.CompareAndSwapInt32(&s.adding[s.idx], 0, 1) {
            for {
                task, b := c.Dequeue()
                if !b {
                    break
                }
                s.callTask(scheduler, t, task)
            }
            atomic.StoreInt32(&s.adding[s.idx], 0)
        }
        s.idx++
        if s.idx == 9 {
            scheduler.taskLayers[s.level+1].next(scheduler, now)
        }
        scheduler.lastTime += s.ceilInterval
        if s.idx == 10 {
            s.idx = 0
        }
        atomic.StoreInt32(&scheduler.runLock, 0)
        return
    }
    c := s.cells[s.idx]
    if atomic.CompareAndSwapInt32(&s.adding[s.idx], 0, 1) {
        idx := s.idx
        go func() {
            defer atomic.StoreInt32(&s.adding[idx], 0)
            //非最低层,将当前单元格的所有元素添加到下层
            for {
                task, b := c.Dequeue()
                if !b {
                    break
                }
                if scheduler.stopped {
                    return
                }
                err := scheduler.addTask(task)
                if err != nil {
                    s.callTask(scheduler, time.UnixMilli(now), task)
                }
            }
        }()
    }
    s.idx++
    if s.idx == 9 {
        if len(scheduler.taskLayers)-1 == s.level {
            s.disposePending(scheduler)
        } else {
            scheduler.taskLayers[s.level+1].next(scheduler, now)
        }
    }
    if s.idx == 10 {
        s.idx = 0
    }
}

func (s *schedulerLayer) disposePending(scheduler *Scheduler) {
    maxTime := scheduler.lastTime + s.ceilInterval*10
    var addTasks bool
    for _, task := range scheduler.pendingTasks {
        if task.time < maxTime {
            addTasks = true
            break
        }
    }
    if addTasks {
        scheduler.pendingLock.Lock()
        defer scheduler.pendingLock.Unlock()
        old := scheduler.pendingTasks
        scheduler.pendingTasks = nil
        for _, task := range old {
            if task.time < maxTime {
                if scheduler.stopped {
                    return
                }
                err := scheduler.addTask(task)
                if err != nil {
                    s.callTask(scheduler, time.Now(), task)
                }
            } else {
                scheduler.pendingTasks = append(scheduler.pendingTasks, task)
            }
        }
    }
}

func (s *schedulerLayer) callTask(scheduler *Scheduler, now time.Time, task schedulerTask) {
    if task.fn != nil {
        if scheduler.DoCallFn != nil {
            scheduler.DoCallFn(task.fn)
        } else {
            go task.fn()
        }
        if task.nextTime != nil {
            nextTime := task.nextTime(now)
            if nextTime.After(now) {
                task.time = nextTime.UnixMilli()
                if scheduler.stopped {
                    return
                }
                err := scheduler.addTask(task)
                if err != nil && !scheduler.stopped {
                    s.callTask(scheduler, now, task)
                }
                return
            }
        }
    }
    atomic.AddInt32(&scheduler.taskCount, -1)
}

func (s *schedulerLayer) addTask(scheduler *Scheduler, task schedulerTask) (err error) {
    delay := task.time - scheduler.lastTime
    if s.level == 0 {
        if delay < 0 {
            return fmt.Errorf("you can't add a expired task,lastTime:%d,task:%d", scheduler.lastTime, task.time)
        }
        if delay == 0 {
            s.callTask(scheduler, time.Now(), task)
            return
        }
    }
    //延迟高于当前层支持,尝试添加到下一层
    if delay > s.allCeilInterval {
        if s.level < len(scheduler.taskLayers)-1 {
            scheduler.taskLayers[s.level+1].addTask(scheduler, task)
        } else {
            level := len(scheduler.taskLayers)
            scheduler.pendingLock.Lock()
            if len(scheduler.taskLayers) == level {
                //未扩容
                scheduler.pendingTasks = append(scheduler.pendingTasks, task)
                //尝试扩容
                if len(scheduler.taskLayers) < scheduler.maxLevel && scheduler.taskCount > 100 && float32(len(scheduler.pendingTasks))/float32(scheduler.taskCount) > 0.15 {
                    scheduler.expansion()
                } else {
                    scheduler.pendingLock.Unlock()
                }
                atomic.AddInt32(&scheduler.taskCount, 1)
                return
            } else {
                //已扩容
                scheduler.pendingLock.Unlock()
                scheduler.taskLayers[s.level+1].addTask(scheduler, task)
            }
        }
        return
    }
    atomic.AddInt32(&scheduler.taskCount, 1)
    var i int
    if s.level > 0 {
        i = (int((delay-s.ceilInterval)/s.ceilInterval) + s.idx) % 10
        if i < 0 {
            i = -i
        }
    } else {
        i = (int(delay/s.ceilInterval) + s.idx) % 10
    }
    if s.level == 0 && i == 0 && atomic.CompareAndSwapInt32(&scheduler.runLock, 1, 1) {
        //正在运行
        s.callTask(scheduler, time.Now(), task)
        return
    }
    s.cells[i].Enqueue(task)
    return
}

type Scheduler struct {
    //检查间隔
    interval int64
    //任务层,每层是一个时间轮,时间为interval的倍数
    taskLayers []schedulerLayer
    maxLevel   int

    lastTime  int64
    runLock   int32
    taskCount int32

    //最外层都放不下的任务
    pendingLock  sync.Mutex
    pendingTasks []schedulerTask

    //自定义任务运行方式,默认使用 go fn()
    DoCallFn func(func())
    stop     chan struct{}
    stopped  bool
}

func (s *Scheduler) addTask(task schedulerTask) error {
    if s.stopped {
        return fmt.Errorf("scheduler has been stopped")
    }
    if task.fn == nil {
        return fmt.Errorf("task.fn can't be nil")
    }
    return s.taskLayers[0].addTask(s, task)
}

func (s *Scheduler) AddDelayTask(delay time.Duration, task func()) error {
    if delay < 0 {
        return fmt.Errorf("delay must be greater than 0")
    }
    return s.addTask(schedulerTask{
        fn:   task,
        time: time.Now().Add(delay).UnixMilli(),
    })
}

func (s *Scheduler) AddIntervalTask(interval time.Duration, task func()) error {
    if interval.Milliseconds() < s.interval {
        return fmt.Errorf("interval must be greater than scheduler interval")
    }
    return s.AddCustomizeTask(func(t time.Time) time.Time { return t.Add(interval) }, task)
}

func (s *Scheduler) AddCustomizeTask(customizeTime func(time.Time) time.Time, task func()) error {
    return s.addTask(schedulerTask{
        fn:       task,
        nextTime: customizeTime,
        time:     customizeTime(time.Now()).UnixMilli(),
    })
}

// AddCronTask 添加一个cron任务,支持5位到7位的cron表达式
func (s *Scheduler) AddCronTask(cron string, task func()) error {
    sc, err := ParseCron(cron)
    if err != nil {
        return err
    }
    return s.taskLayers[0].addTask(s, schedulerTask{
        fn:       task,
        nextTime: sc.NextTime,
        time:     sc.NextTime(time.Now()).UnixMilli(),
    })
}

func (s *Scheduler) Stop() {
    if s.stopped {
        return
    }
    s.stopped = true
    close(s.stop)
}

func (s *Scheduler) WaitStop() {
    <-s.stop
}

func (s *Scheduler) log() string {
    var b []byte
    b = append(b, fmt.Sprintln("taskCount:", s.taskCount)...)
    b = append(b, fmt.Sprintln("pendingTasks:", len(s.pendingTasks))...)
    for _, layer := range s.taskLayers {
        b = append(b, fmt.Sprintln("level:", layer.level, (time.Millisecond*time.Duration(layer.ceilInterval)).String())...)
        for i := range layer.cells {
            cell := layer.cells[(i+layer.idx)%10]
            b = append(b, fmt.Sprintln("cell:", (time.Duration(i+1)*time.Millisecond*time.Duration(layer.ceilInterval)).String(), cell.Size())...)
        }
    }
    return string(b)
}

//todo:cron表达式支持

func (s *Scheduler) expansion() {
    defer s.pendingLock.Unlock()
    ceilInterval := s.taskLayers[len(s.taskLayers)-1].ceilInterval * 10
    now := time.Now().UnixMilli()
    newLayers := schedulerLayer{
        level:           len(s.taskLayers),
        ceilInterval:    ceilInterval,
        allCeilInterval: ceilInterval * 10,
    }
    for i := 0; i < 10; i++ {
        newLayers.cells[i] = concurrent.NewQueue[schedulerTask]()
    }
    ceilInterval *= 10
    s.taskLayers = append(s.taskLayers, newLayers)
    tasks := s.pendingTasks
    s.pendingTasks = make([]schedulerTask, 0)
    for _, task := range tasks {
        if task.time-now < ceilInterval {
            newLayers.addTask(s, task)
            continue
        }
        s.pendingTasks = append(s.pendingTasks, task)
    }
}

// NewScheduler 创建一个调度器
func NewScheduler(checkInterval ...time.Duration) *Scheduler {
    interval := time.Millisecond * 100
    if len(checkInterval) > 0 {
        interval = checkInterval[0]
    }
    //最小间隔,再小精度无法保证
    if interval < time.Millisecond*5 {
        panic("interval must be greater than 10ms")
    }
    if interval > time.Second*10 {
        panic("interval must be less than 10s")
    }
    level := int(math.Log10(float64(time.Hour.Milliseconds()))) + 1
    if level < 5 {
        level = 5
    }
    s := &Scheduler{
        interval:   interval.Milliseconds(),
        taskLayers: make([]schedulerLayer, 3),
        maxLevel:   level,
        stop:       make(chan struct{}),
        lastTime:   time.Now().UnixMilli(),
    }
    for i := range s.taskLayers {
        layer := schedulerLayer{
            level:           i,
            ceilInterval:    interval.Milliseconds(),
            allCeilInterval: interval.Milliseconds() * 10,
        }
        for j := 0; j < 10; j++ {
            layer.cells[j] = concurrent.NewQueue[schedulerTask]()
        }
        s.taskLayers[i] = layer
        interval *= 10
    }
    go func() {
        s.lastTime = time.Now().UnixMilli()
        tick := time.NewTicker(time.Millisecond * time.Duration(s.interval))
        defer tick.Stop()
        for {
            select {
            case <-s.stop:
                return
            case <-tick.C:
                s.taskLayers[0].doNext(s)
            }
        }
    }()
    return s
}
