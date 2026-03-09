// Package di 提供容器生命周期管理功能
package di

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/mzzsfy/go-util/helper"
)

// checkAndSetStartedState 检查并设置启动状态
// 使用双重检查锁定确保线程安全
// 返回:
//   - 如果容器已启动返回错误
func (c *container) checkAndSetStartedState() error {
	// 第一次检查（读锁）
	c.mu.RLock()
	started := c.started
	c.mu.RUnlock()
	if started {
		return helper.StringError("container is already started")
	}

	// 第二次检查（写锁）
	c.mu.Lock()
	if c.started {
		c.mu.Unlock()
		return helper.StringError("container is already started")
	}
	return nil
}

// resetDoneChannelIfNeeded 重置 done 通道
// 如果通道已关闭，创建新的通道
func (c *container) resetDoneChannelIfNeeded() {
	select {
	case <-c.done:
		c.done = make(chan struct{})
	default:
	}
}

// executeStartupHooks 执行启动钩子
// 按顺序执行 onStartup 和 afterStartup 钩子
// 返回:
//   - 任何钩子执行失败时返回错误
func (c *container) executeStartupHooks() error {
	executeHooks := func(hooks []func(Container) error, hookType string) error {
		for i, hook := range hooks {
			if err := hook(c); err != nil {
				c.mu.Lock()
				c.started = false
				c.mu.Unlock()
				return fmt.Errorf("%s hook %d failed: %w", hookType, i, err)
			}
		}
		return nil
	}

	if err := executeHooks(c.onStartup, "startup"); err != nil {
		return err
	}

	if err := executeHooks(c.afterStartup, "after startup"); err != nil {
		return err
	}

	c.onStartup = nil
	c.afterStartup = nil
	return nil
}

// Start 启动容器
// 调用启动钩子，标记容器为已启动状态
func (c *container) Start() error {
	if err := c.checkAndSetStartedState(); err != nil {
		return err
	}

	c.resetDoneChannelIfNeeded()

	c.started = true
	c.mu.Unlock()

	return c.executeStartupHooks()
}

// Shutdown 关闭容器
// 执行所有关闭钩子并清理资源
func (c *container) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.checkAlreadyShutdown(); err != nil {
		return err
	}

	err := c.executeShutdownHooks(ctx)

	c.cleanupResources()

	return err
}

// checkAlreadyShutdown 检查是否已关闭
func (c *container) checkAlreadyShutdown() error {
	select {
	case <-c.Done():
		return helper.StringError("container is already shutting down")
	default:
		return nil
	}
}

// executeShutdownHooks 执行关闭钩子
// 按注册顺序执行所有关闭钩子
func (c *container) executeShutdownHooks(ctx context.Context) error {
	var err error
	for _, hook := range c.shutdown {
		if hookErr := hook(ctx); hookErr != nil {
			err = fmt.Errorf("shutdown failed: %w", hookErr)
		}
	}
	return err
}

// cleanupResources 清理容器资源
// 重置所有内部状态
func (c *container) cleanupResources() {
	c.providers = make(map[string]providerEntry)
	c.instances = make(map[string]any)
	c.shutdown = nil
	c.started = false

	c.configMu.Lock()
	c.configSource = NewMapConfigSource()
	c.configMu.Unlock()

	c.statsMu.Lock()
	c.stats = containerStats{}
	c.statsMu.Unlock()

	close(c.done)
}

// ShutdownOnSignals 监听系统信号并自动关闭
// 默认监听 SIGTERM 和 Interrupt 信号
func (c *container) ShutdownOnSignals(signals ...os.Signal) {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGTERM, os.Interrupt}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)

	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal %s, shutting down...\n", sig)
		if err := c.Shutdown(context.Background()); err != nil {
			fmt.Printf("Shutdown error: %v\n", err)
		}
		os.Exit(0)
	}()
}

// Done 返回关闭通知通道
// 通道关闭时表示容器已关闭
func (c *container) Done() <-chan struct{} {
	return c.done
}

// CreateChildScope 创建子容器
// 子容器继承父容器的配置源，可以访问父容器的服务
func (c *container) CreateChildScope() Container {
	c.configMu.RLock()
	inheritedConfigSource := c.configSource
	c.configMu.RUnlock()

	c2 := &container{
		providers:    make(map[string]providerEntry),
		instances:    make(map[string]any),
		loading:      make(map[string]bool),
		parent:       c,
		done:         make(chan struct{}),
		configSource: inheritedConfigSource,
	}
	c.shutdown = append(c.shutdown, func(ctx context.Context) error { return c2.Shutdown(ctx) })
	return c2
}

// hasDestroyHooks 检查是否有销毁钩子
func (c *container) hasDestroyHooks(entry providerEntry) bool {
	return len(entry.config.beforeDestroy) > 0 ||
		len(entry.config.afterDestroy) > 0 ||
		len(c.beforeDestroy) > 0 ||
		len(c.afterDestroy) > 0
}

// registerDestroyHook 注册销毁钩子
// 将实例的销毁逻辑添加到容器的关闭钩子列表
func (c *container) registerDestroyHook(entry providerEntry, name string, instance any) {
	if !c.hasDestroyHooks(entry) {
		return
	}

	destroyHook := c.createDestroyHook(entry, name, instance)
	c.shutdown = append(c.shutdown, destroyHook)
}

// executeDestroyHookList 执行销毁钩子列表
// 通用的销毁钩子执行函数
func (c *container) executeDestroyHookList(hooks []func(Container, EntryInfo), destroyInfo EntryInfo) {
	for _, f := range hooks {
		f(c, destroyInfo)
	}
}

// executeInstanceDestroy 执行实例销毁
// 调用 DestroyCallback 和 ServiceLifecycle 接口
func (c *container) executeInstanceDestroy(ctx context.Context, instance any, reflectType reflect.Type, name string) error {
	if lifecycle, ok := instance.(DestroyCallback); ok {
		if err := lifecycle.OnDestroyCallback(); err != nil {
			return destroyCallbackError(reflectType, name, err)
		}
	}

	if lifecycle, ok := instance.(ServiceLifecycle); ok {
		if err := lifecycle.Shutdown(ctx); err != nil {
			return shutdownError(reflectType, name, err)
		}
	}

	return nil
}

// createDestroyHook 创建销毁钩子
// 组装完整的销毁流程：前置钩子 -> 实例销毁 -> 后置钩子
func (c *container) createDestroyHook(entry providerEntry, name string, instance any) ShutdownHook {
	beforeDestroy := entry.config.beforeDestroy
	afterDestroy := entry.config.afterDestroy

	return func(ctx context.Context) error {
		containerContext := ContainerContext{parent: ctx}
		destroyInfo := EntryInfo{
			Instance: instance,
			Name:     name,
			Ctx:      containerContext,
		}

		// 执行销毁前钩子：先容器级别，再实例级别
		c.executeDestroyHookList(c.beforeDestroy, destroyInfo)
		c.executeDestroyHookList(beforeDestroy, destroyInfo)

		if err := c.executeInstanceDestroy(ctx, instance, entry.reflectType, name); err != nil {
			return err
		}

		// 执行销毁后钩子：先实例级别，再容器级别
		c.executeDestroyHookList(afterDestroy, destroyInfo)
		c.executeDestroyHookList(c.afterDestroy, destroyInfo)

		return nil
	}
}
