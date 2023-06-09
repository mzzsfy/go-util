package util

import (
    "fmt"
    "math"
    "os"
    "os/signal"
    "sort"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
)

var (
    inited  atomic.Value
    lock    = sync.Mutex{}
    initFns []*fn
    exitFns []*fn
    funcWg  *sync.WaitGroup
)

type fn struct {
    name  string
    order int
    f     func()
}

// DoFuncWaitCallback 当初始化或者退出时,如果任务有前后依赖关系,调用此方法开始阻塞,返回一个解除阻塞的回调函数
func DoFuncWaitCallback() func() {
    if funcWg == nil {
        lock.Lock()
        if funcWg == nil {
            funcWg = &sync.WaitGroup{}
        }
        lock.Unlock()
    }
    funcWg.Add(1)
    once := sync.Once{}
    return func() { once.Do(funcWg.Done) }
}

// AfterInit 注册init回调,使用DoAfterInit时被调用,默认order为0
func AfterInit(name string, f func()) {
    AfterInitOrder(name, f, 0)
}
func AfterInitOrder(name string, f func(), order int) {
    if inited.Load() != nil {
        panic("初始化完成后不允许再次注册回调")
    }
    lock.Lock()
    defer lock.Unlock()
    initFns = append(initFns, &fn{
        name:  name,
        order: order,
        f:     f,
    })
    sort.Slice(initFns, func(i, j int) bool {
        return initFns[i].order < initFns[j].order
    })
}

func DoAfterInit() (success bool) {
    if !inited.CompareAndSwap(nil, true) {
        panic("不允许重复执行Init")
    }
    group := sync.WaitGroup{}
    success = true
    lastOrder := math.MinInt32
    for _, fn := range initFns {
        f := fn
        //分批异步进行
        if lastOrder > f.order {
            time.Sleep(time.Millisecond)
            if funcWg != nil {
                funcWg.Wait()
                lock.Lock()
                funcWg = nil
                lock.Unlock()
            }
        }
        lastOrder = f.order
        group.Add(1)
        go func() {
            TryWithStack(f.f, func(err Err) {
                success = false
                fmt.Fprintf(os.Stderr, "%s执行init回调时发生错误!错误是: %+v\n%s", f.name, err.Error, FormatStack(err.Stack[:2]))
            })
            group.Done()
        }()
    }
    initFns = nil
    group.Wait()
    if !success {
        Exit("init时发生错误")
    }
    success = success && !exiting
    return
}

// BeforeExit 注册非异步优雅关闭回调,默认order为0
func BeforeExit(name string, f func()) {
    BeforeExitOrder(name, f, 0)
}
func BeforeExitOrder(name string, f func(), order int) {
    lock.Lock()
    defer lock.Unlock()
    exitFns = append(exitFns, &fn{
        name:  name,
        order: order,
        f:     f,
    })
    sort.Slice(exitFns, func(i, j int) bool {
        return exitFns[i].order < exitFns[j].order
    })
}

func init() {
    go func() {
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Kill, os.Interrupt, syscall.SIGTERM)
        <-c
        doExit(0)
    }()
}

func doExit(code int) {
    exiting = true
    group := sync.WaitGroup{}
    lastOrder := math.MinInt32
    for _, fn := range exitFns {
        f := fn
        //分批异步进行
        if lastOrder > f.order {
            time.Sleep(time.Millisecond)
            if funcWg != nil {
                lock.Lock()
                funcWg = nil
                lock.Unlock()
            }
        }
        lastOrder = f.order
        group.Add(1)
        go func() {
            defer func() {
                recover()
                group.Done()
            }()
            f.f()
        }()
    }
    go func() {
        time.Sleep(5 * time.Second)
        println("正常退出超时(5s),强制退出!")
        realExit(code)
    }()
    group.Wait()
    realExit(code)
}

// DoBlock 阻塞当前goroutine,直到panic
func DoBlock() {
    if a := recover(); a != nil {
        Exit("出现错误,%v", a)
    }
    select {}
}
