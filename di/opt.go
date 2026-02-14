// Package di 提供依赖注入容器的选项配置功能
package di

import (
	"context"
	"time"

	"github.com/mzzsfy/go-util/helper"
)

// ContainerContext 容器上下文
// 实现了 context.Context 接口，用于在钩子中传递上下文信息
type ContainerContext struct {
	// parent 父上下文
	parent context.Context
	// err 错误信息
	err error
}

// Deadline 返回上下文的截止时间
// 委托给父上下文实现
func (c ContainerContext) Deadline() (deadline time.Time, ok bool) {
	if c.parent != nil {
		return c.parent.Deadline()
	}
	return time.Time{}, false
}

// Done 返回上下文的完成通道
// 委托给父上下文实现
func (c ContainerContext) Done() <-chan struct{} {
	if c.parent != nil {
		return c.parent.Done()
	}
	return nil
}

// Err 返回上下文的错误信息
func (c ContainerContext) Err() error {
	return c.err
}

// Value 返回上下文中键对应的值
// 委托给父上下文实现
func (c ContainerContext) Value(key any) any {
	if c.parent != nil {
		return c.parent.Value(key)
	}
	return nil
}

// EntryInfo 服务条目信息
// 包含服务实例的元数据，在钩子函数中使用
type EntryInfo struct {
	// Name 服务名称
	Name string
	// Instance 服务实例
	Instance any
	// Ctx 容器上下文
	Ctx ContainerContext
}

// ProviderOption 服务提供者选项函数
// 用于配置提供者的行为
type ProviderOption func(*providerConfig)

// providerConfig 服务提供者配置
// 定义服务创建和销毁的行为
type providerConfig struct {
	// loadMode 加载模式
	loadMode LoadMode
	// beforeCreate 创建前钩子列表
	// 可以拦截、替换或阻止服务创建
	beforeCreate []func(Container, EntryInfo) (any, error)
	// afterCreate 创建后钩子列表
	// 可以修改或替换已创建的服务实例
	afterCreate []func(Container, EntryInfo) (any, error)
	// beforeDestroy 销毁前钩子列表
	beforeDestroy []func(Container, EntryInfo)
	// afterDestroy 销毁后钩子列表
	afterDestroy []func(Container, EntryInfo)
}

// WithCondition 设置条件函数
// 只有条件满足时才调用 provider
// condition: 条件判断函数，返回 true 表示允许创建
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
// mode: 加载模式（Default/Immediate/Lazy/Transient）
func WithLoadMode(mode LoadMode) ProviderOption {
	return func(pc *providerConfig) {
		pc.loadMode = mode
	}
}

// WithBeforeCreate 设置创建前钩子
// f: 钩子函数，可以返回替换的实例或错误
func WithBeforeCreate(f func(Container, EntryInfo) (any, error)) ProviderOption {
	return func(pc *providerConfig) {
		pc.beforeCreate = append(pc.beforeCreate, f)
	}
}

// WithAfterCreate 设置创建后钩子
// f: 钩子函数，可以修改或替换已创建的实例
func WithAfterCreate(f func(Container, EntryInfo) (any, error)) ProviderOption {
	return func(pc *providerConfig) {
		pc.afterCreate = append(pc.afterCreate, f)
	}
}

// WithBeforeDestroy 设置销毁前钩子
// f: 钩子函数，在实例销毁前执行
func WithBeforeDestroy(f func(Container, EntryInfo)) ProviderOption {
	return func(pc *providerConfig) {
		pc.beforeDestroy = append(pc.beforeDestroy, f)
	}
}

// WithAfterDestroy 设置销毁后钩子
// f: 钩子函数，在实例销毁后执行
func WithAfterDestroy(f func(Container, EntryInfo)) ProviderOption {
	return func(pc *providerConfig) {
		pc.afterDestroy = append(pc.afterDestroy, f)
	}
}

// ContainerOption 容器选项函数
// 用于配置容器的全局行为
type ContainerOption func(*container)

// checkNotStarted 检查容器是否未启动
// 如果已启动则 panic
func checkNotStarted(c *container) {
	if c.started {
		panic(helper.StringError("cannot add options to a started container"))
	}
}

// WithContainerBeforeCreate 设置容器级别的创建前钩子
// 对所有服务生效
func WithContainerBeforeCreate(f func(Container, EntryInfo) (any, error)) ContainerOption {
	return func(c *container) {
		checkNotStarted(c)
		c.beforeCreate = append(c.beforeCreate, f)
	}
}

// WithContainerAfterCreate 设置容器级别的创建后钩子
// 对所有服务生效
func WithContainerAfterCreate(f func(Container, EntryInfo) (any, error)) ContainerOption {
	return func(c *container) {
		checkNotStarted(c)
		c.afterCreate = append(c.afterCreate, f)
	}
}

// WithContainerBeforeDestroy 设置容器级别的销毁前钩子
// 对所有服务生效
func WithContainerBeforeDestroy(f func(Container, EntryInfo)) ContainerOption {
	return func(c *container) {
		checkNotStarted(c)
		c.beforeDestroy = append(c.beforeDestroy, f)
	}
}

// WithContainerAfterDestroy 设置容器级别的销毁后钩子
// 对所有服务生效
func WithContainerAfterDestroy(f func(Container, EntryInfo)) ContainerOption {
	return func(c *container) {
		checkNotStarted(c)
		c.afterDestroy = append(c.afterDestroy, f)
	}
}

// WithContainerOnStart 设置启动前钩子
// 在 Start 方法调用时执行
func WithContainerOnStart(f func(Container) error) ContainerOption {
	return func(c *container) {
		checkNotStarted(c)
		c.onStartup = append(c.onStartup, f)
	}
}

// WithContainerAfterStart 设置启动后钩子
// 在 Start 方法调用后执行
func WithContainerAfterStart(f func(Container) error) ContainerOption {
	return func(c *container) {
		checkNotStarted(c)
		c.afterStartup = append(c.afterStartup, f)
	}
}

// WithContainerBeforeShutdown 设置关闭前钩子
// 在其他关闭钩子之前执行
func WithContainerBeforeShutdown(f ShutdownHook) ContainerOption {
	return func(c *container) {
		hooks := make([]ShutdownHook, 0, len(c.shutdown)+1)
		hooks = append(hooks, f)
		hooks = append(hooks, c.shutdown...)
		c.shutdown = hooks
	}
}

// WithContainerShutdown 设置关闭钩子
// 在容器关闭时执行
func WithContainerShutdown(f ShutdownHook) ContainerOption {
	return func(c *container) {
		c.shutdown = append(c.shutdown, f)
	}
}
