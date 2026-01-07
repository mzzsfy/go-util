package di

import (
    "context"
    "time"

    "github.com/mzzsfy/go-util/helper"
)

type ContainerContext struct {
    parent context.Context
    err    error
}

func (c ContainerContext) Deadline() (deadline time.Time, ok bool) {
    if c.parent != nil {
        return c.parent.Deadline()
    }
    return time.Time{}, false
}

func (c ContainerContext) Done() <-chan struct{} {
    if c.parent != nil {
        return c.parent.Done()
    }
    return nil
}

func (c ContainerContext) Err() error {
    return c.err
}

func (c ContainerContext) Value(key any) any {
    if c.parent != nil {
        return c.parent.Value(key)
    }
    return nil
}

// EntryInfo 包含服务实例和名称信息
type EntryInfo struct {
    Name     string
    Instance any
    Ctx      ContainerContext
}

// ProviderOption 服务提供者选项函数
type ProviderOption func(*providerConfig)

// providerConfig 服务提供者配置
type providerConfig struct {
    loadMode      LoadMode                                  // 加载模式
    beforeCreate  []func(Container, EntryInfo) (any, error) //加载前调用,用于替换和阻止
    afterCreate   []func(Container, EntryInfo) (any, error) //加载后调用,用于替换和删除
    beforeDestroy []func(Container, EntryInfo)              //销毁前调用
    afterDestroy  []func(Container, EntryInfo)              //销毁后调用
}

// WithCondition 设置条件函数，只有条件满足时才调用 provider
func WithCondition(condition func(Container) bool) ProviderOption {
    return func(pc *providerConfig) {
        pc.beforeCreate = append(pc.beforeCreate, func(c Container, info EntryInfo) (any, error) {
            if condition(c) {
                return nil, nil
            }
            return nil, ErrorConditionFail
        })
    }
}

// WithLoadMode 设置加载模式
func WithLoadMode(mode LoadMode) ProviderOption {
    return func(pc *providerConfig) {
        pc.loadMode = mode
    }
}

// WithBeforeCreate 设置创建前钩子
func WithBeforeCreate(f func(Container, EntryInfo) (any, error)) ProviderOption {
    return func(pc *providerConfig) {
        pc.beforeCreate = append(pc.beforeCreate, f)
    }
}

// WithAfterCreate 设置创建后钩子
func WithAfterCreate(f func(Container, EntryInfo) (any, error)) ProviderOption {
    return func(pc *providerConfig) {
        pc.afterCreate = append(pc.afterCreate, f)
    }
}

// WithBeforeDestroy 设置销毁前钩子
func WithBeforeDestroy(f func(Container, EntryInfo)) ProviderOption {
    return func(pc *providerConfig) {
        pc.beforeDestroy = append(pc.beforeDestroy, f)
    }
}

// WithAfterDestroy 设置销毁后钩子
func WithAfterDestroy(f func(Container, EntryInfo)) ProviderOption {
    return func(pc *providerConfig) {
        pc.afterDestroy = append(pc.afterDestroy, f)
    }
}

// ContainerOption 容器选项函数
type ContainerOption func(*container)

// WithContainerBeforeCreate 设置容器级别的创建前钩子
func WithContainerBeforeCreate(f func(Container, EntryInfo) (any, error)) ContainerOption {
    return func(c *container) {
        if c.started {
            panic(helper.StringError("cannot add options to a started container"))
        }
        c.beforeCreate = append(c.beforeCreate, f)
    }
}

// WithContainerAfterCreate 设置容器级别的创建后钩子
func WithContainerAfterCreate(f func(Container, EntryInfo) (any, error)) ContainerOption {
    return func(c *container) {
        if c.started {
            panic(helper.StringError("cannot add options to a started container"))
        }
        c.afterCreate = append(c.afterCreate, f)
    }
}

// WithContainerBeforeDestroy 设置容器级别的销毁前钩子
func WithContainerBeforeDestroy(f func(Container, EntryInfo)) ContainerOption {
    return func(c *container) {
        if c.started {
            panic(helper.StringError("cannot add options to a started container"))
        }
        c.beforeDestroy = append(c.beforeDestroy, f)
    }
}

// WithContainerAfterDestroy 设置容器级别的销毁后钩子
func WithContainerAfterDestroy(f func(Container, EntryInfo)) ContainerOption {
    return func(c *container) {
        if c.started {
            panic(helper.StringError("cannot add options to a started container"))
        }
        c.afterDestroy = append(c.afterDestroy, f)
    }
}

// WithContainerOnStart 设置启动前钩子（在Start方法调用时执行）
func WithContainerOnStart(f func(Container) error) ContainerOption {
    return func(c *container) {
        if c.started {
            panic(helper.StringError("cannot add options to a started container"))
        }
        c.onStartup = append(c.onStartup, f)
    }
}

// WithContainerAfterStart 设置启动后钩子（在Start方法调用后执行）
func WithContainerAfterStart(f func(Container) error) ContainerOption {
    return func(c *container) {
        if c.started {
            panic(helper.StringError("cannot add options to a started container"))
        }
        c.afterStartup = append(c.afterStartup, f)
    }
}

func WithContainerBeforeShutdown(f ShutdownHook) ContainerOption {
    return func(c *container) {
        // Create new slice with the new hook at the beginning, then append existing hooks
        // This matches the original intent but fixes the copy bug
        hooks := make([]ShutdownHook, 0, len(c.shutdown)+1)
        hooks = append(hooks, f)
        hooks = append(hooks, c.shutdown...)
        c.shutdown = hooks
    }
}
func WithContainerShutdown(f ShutdownHook) ContainerOption {
    return func(c *container) {
        c.shutdown = append(c.shutdown, f)
    }
}
