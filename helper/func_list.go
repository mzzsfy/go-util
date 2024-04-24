package helper

import (
    "context"
    "fmt"
    "sync"
)

// FuncList 类似状态模式的工具,这个工具能将一个大功能分成多个小功能,方便重试或反复提取参数等场景
type FuncList struct {
    lock  sync.Mutex
    funcs []func(*FuncListContext)
}

type FuncListContext struct {
    funcs      []func(*FuncListContext)
    callNumber []int
    ctx        context.Context
    idx, idx1  int
    redirect   int
}

func (c *FuncListContext) Context() context.Context {
    return c.ctx
}

func (c *FuncListContext) Value(key any) any {
    return c.ctx.Value(key)
}

func (c *FuncListContext) WithContext(ctx context.Context) {
    if ctx == nil {
        panic("cannot create context from nil parent")
    }
    c.ctx = ctx
}

func (c *FuncListContext) WithValue(key, value any) {
    c.ctx = context.WithValue(c.ctx, key, value)
}

// CallNumber 获取当前步骤被调用次数
func (c *FuncListContext) CallNumber() int {
    return c.callNumber[c.idx1]
}

// FuncCount 获取一共多少个步骤
func (c *FuncListContext) FuncCount() int {
    return len(c.funcs)
}

// FuncIndex 获取当前步骤索引
func (c *FuncListContext) FuncIndex() int {
    return c.idx1
}

//RedirectStep 调整执行位置,负数表示从前面第n步开始重新执行,正数表示跳过下n个步骤,0表示正常执行
func (c *FuncListContext) RedirectStep(i int) {
    c.redirect = i
}

// ContinueStep 主动继续执行后续步骤,一般无需调用,阻塞
func (c *FuncListContext) ContinueStep() {
    idx := c.idx
    if c.redirect != 0 {
        if c.redirect > 0 {
            c.idx += c.redirect
        } else {
            c.idx += c.redirect - 1
        }
        c.redirect = 0
    }
    for ; c.idx < len(c.funcs); c.idx++ {
        if c.idx < 0 {
            c.idx = 0
        }
        idx1 := c.idx
        c.idx1 = idx1
        c.funcs[idx1](c)
        c.callNumber[idx1]++
        if c.redirect != 0 {
            if c.redirect > 0 {
                c.idx += c.redirect
            } else {
                c.idx += c.redirect - 1
            }
            c.redirect = 0
        }
    }
    c.idx1 = idx
}

// AddFunc 注册一个步骤
func (l *FuncList) AddFunc(f func(*FuncListContext)) {
    l.lock.Lock()
    defer l.lock.Unlock()
    l.funcs = append(l.funcs, f)
}

// Complete 执行注册的步骤
func (l *FuncList) Complete(ctx ...context.Context) context.Context {
    c := context.Background()
    if len(ctx) > 0 {
        c = ctx[0]
    }
    l.lock.Lock()
    ctx1 := &FuncListContext{
        ctx:        c,
        funcs:      l.funcs,
        callNumber: make([]int, len(l.funcs)),
    }
    l.lock.Unlock()
    ctx1.ContinueStep()
    return ctx1.ctx
}

// FuncListNamed 使用名称的步骤工具
type FuncListNamed struct {
    lock  sync.Mutex
    funcs []func(*FuncListNamedContext)
    names []string
}

type FuncListNamedContext struct {
    funcs      []func(*FuncListNamedContext)
    names      []string
    callNumber []int
    ctx        context.Context
    idx, idx1  int
    redirect   string
}

func (c *FuncListNamedContext) Context() context.Context {
    return c.ctx
}

func (c *FuncListNamedContext) Value(key any) any {
    return c.ctx.Value(key)
}

func (c *FuncListNamedContext) WithContext(ctx context.Context) {
    if ctx == nil {
        panic("cannot create context from nil parent")
    }
    c.ctx = ctx
}

func (c *FuncListNamedContext) WithValue(key, value any) {
    c.ctx = context.WithValue(c.ctx, key, value)
}

// CallNumber 获取当前步骤被调用次数
func (c *FuncListNamedContext) CallNumber() int {
    return c.callNumber[c.idx1]
}

// FuncCount 获取一共多少个步骤
func (c *FuncListNamedContext) FuncCount() int {
    return len(c.funcs)
}

// FuncIndex 获取当前步骤索引
func (c *FuncListNamedContext) FuncIndex() int {
    return c.idx1
}

// FuncName 获取当前步骤名称
func (c *FuncListNamedContext) FuncName() string {
    return c.names[c.idx1]
}

//RedirectStep 调整执行位置
func (c *FuncListNamedContext) RedirectStep(name string) {
    c.redirect = name
}

//RedirectStepStop 调整执行位置到末尾,阻塞
func (c *FuncListNamedContext) RedirectStepStop() {
    c.idx = len(c.funcs) + 1
}

// ContinueStep 主动继续执行后续步骤,一般无需调用,阻塞
func (c *FuncListNamedContext) ContinueStep() {
    idx := c.idx
    if c.redirect != "" {
        for i, name := range c.names {
            if c.redirect == name {
                c.idx = i
                goto ok
            }
        }
        panic(fmt.Sprintf("没有这个名称:%s", c.redirect))
    ok:
        c.redirect = ""
    }
    for ; c.idx < len(c.funcs); c.idx++ {
        if c.idx < 0 {
            c.idx = 0
        }
        idx1 := c.idx
        c.idx1 = idx1
        c.funcs[idx1](c)
        c.callNumber[idx1]++
        if c.redirect != "" {
            for i, name := range c.names {
                if c.redirect == name {
                    c.idx = i
                    goto ok1
                }
            }
            panic(fmt.Sprintf("没有这个名称:%s", c.redirect))
        ok1:
            c.redirect = ""
        }
    }
    c.idx1 = idx
}

// AddFunc 注册一个步骤
func (l *FuncListNamed) AddFunc(name string, f func(*FuncListNamedContext)) {
    if name == "" {
        panic("名称不能为空")
    }
    l.lock.Lock()
    defer l.lock.Unlock()
    for i, name1 := range l.names {
        if name1 == name {
            panic(fmt.Sprintf("重复注册步骤:(%d)%s", i, name))
        }
    }
    l.funcs = append(l.funcs, f)
    l.names = append(l.names, name)
}

// Complete 执行注册的步骤
func (l *FuncListNamed) Complete(ctx ...context.Context) context.Context {
    c := context.Background()
    if len(ctx) > 0 {
        c = ctx[0]
    }
    l.lock.Lock()
    ctx1 := &FuncListNamedContext{
        ctx:        c,
        funcs:      l.funcs,
        names:      l.names,
        callNumber: make([]int, len(l.funcs)),
    }
    l.lock.Unlock()
    ctx1.ContinueStep()
    return ctx1.ctx
}
