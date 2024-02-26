package helper

import (
    "math"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"
)

type fn struct {
    name string
    f    func()
}

// FuncCaller 用于按一定顺序调用多个函数
type FuncCaller struct {
    lock sync.Mutex
    f    []struct {
        order int
        fns   []*fn
    }
}

// AddFnOrder 添加一个回调函数,并指定执行顺序
func (fc *FuncCaller) AddFnOrder(name string, order int, f func()) {
    fc.lock.Lock()
    defer fc.lock.Unlock()
    f1 := &fn{
        name: name,
        f:    f,
    }
    for i, v := range fc.f {
        if v.order == order {
            fc.f[i].fns = append(fc.f[i].fns, f1)
            return
        } else if v.order > order {
            fc.f = append(fc.f[:i], append([]struct {
                order int
                fns   []*fn
            }{{order: order, fns: []*fn{f1}}}, fc.f[i:]...)...)
            return
        }
    }
}

// AddFn 添加一个回调函数,默认执行顺序为0
func (fc *FuncCaller) AddFn(name string, f func()) {
    fc.AddFnOrder(name, 0, f)
}

// CallWithRecover 调用所有注册的回调函数
func (fc *FuncCaller) CallWithRecover() (err Err) {
    fc.lock.Lock()
    defer fc.lock.Unlock()
    wg := sync.WaitGroup{}
    lastOrder := math.MinInt
    for _, v := range fc.f {
        if lastOrder < v.order {
            wg.Wait()
            lastOrder = v.order
        }
        for _, f := range v.fns {
            wg.Add(1)
            go func(f *fn) {
                TryWithStack(f.f, func(e Err) { err = e })
                wg.Done()
            }(f)
        }
    }
    wg.Wait()
    return
}

var (
    inited   = false
    initCall = &FuncCaller{}

    initExit = sync.Once{}
    exitCall = &FuncCaller{}
)

func AfterInit(name string, f func()) {
    AfterInitOrder(name, f, 0)
}

func AfterInitOrder(name string, f func(), order int) {
    initCall.AddFnOrder(name, order, f)
}

func DoAfterInit() (success bool) {
    err := DoAfterInitWithErr()
    if err.Error != nil {
        return false
    }
    return true
}

func DoAfterInitWithErr() (err Err) {
    if inited {
        panic("不允许重复执行Init")
    }
    inited = true
    return initCall.CallWithRecover()
}

func BeforeExit(name string, f func()) {
    BeforeExitOrder(name, f, 0)
}

func BeforeExitOrder(name string, f func(), order int) {
    initExit.Do(func() {
        go func() {
            c := make(chan os.Signal, 1)
            signal.Notify(c, os.Kill, os.Interrupt, syscall.SIGTERM)
            <-c
            doExit(0)
        }()
    })
    exitCall.AddFnOrder(name, order, f)
}

func doExit(code int) {
    exitCall.CallWithRecover()
    time.Sleep(time.Millisecond * 50)
    os.Exit(code)
}
