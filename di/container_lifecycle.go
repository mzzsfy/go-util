// Package di 提供容器生命周期管理功能
package di

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync/atomic"
	"syscall"

	"github.com/mzzsfy/go-util/helper"
)

// Start 启动容器
// 调用启动钩子，标记容器为已启动状态
// Shutdown 后容器不可复活,重复 Start 返回错误
func (c *container) Start() error {
	c.mu.Lock()
	if c.started {
		c.mu.Unlock()
		return helper.StringError("container is already started")
	}
	// 关闭流程已开始则拒绝启动,shutdownStarted 为终态闩
	if atomic.LoadInt32(&c.shutdownStarted) == shutdownStateStart {
		c.mu.Unlock()
		return helper.StringError("container is shut down and cannot be restarted")
	}
	c.started = true
	// 锁内快照启动钩子,与选项追加互斥
	onStartup := c.onStartup
	afterStartup := c.afterStartup
	c.mu.Unlock()

	return c.executeStartupHooks(onStartup, afterStartup)
}

// executeStartupHooks 执行启动钩子
// 按顺序执行启动前后钩子,任一失败回滚启动状态并返回错误
// 全部成功后锁内置空钩子列表,防止重复执行
func (c *container) executeStartupHooks(onStartup, afterStartup []func(Container) error) error {
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

	if err := executeHooks(onStartup, "startup"); err != nil {
		return err
	}

	if err := executeHooks(afterStartup, "after startup"); err != nil {
		return err
	}

	c.mu.Lock()
	c.onStartup = nil
	c.afterStartup = nil
	c.mu.Unlock()
	return nil
}

// Shutdown 关闭容器
// 执行所有关闭钩子并清理资源
// 并发调用只执行一次钩子,其余调用返回已关闭错误
// 关闭后容器进入终态,不可再 Start
func (c *container) Shutdown(ctx context.Context) (err error) {
	if !atomic.CompareAndSwapInt32(&c.shutdownStarted, shutdownStateIdle, shutdownStateStart) {
		return helper.StringError("container is already shutting down")
	}

	// 锁内完成钩子快照,解锁后执行,避免钩子内 Get 死锁
	c.mu.Lock()
	preHooks := make([]ShutdownHook, len(c.preShutdown))
	copy(preHooks, c.preShutdown)
	hooks := make([]ShutdownHook, len(c.shutdown))
	copy(hooks, c.shutdown)
	c.mu.Unlock()

	defer func() {
		// 钩子 panic 时仍完成清理,保持容器状态一致
		if r := recover(); r != nil {
			c.mu.Lock()
			c.cleanupResources()
			c.mu.Unlock()
			panic(r)
		}
	}()

	err = c.executeShutdownHooks(ctx, preHooks, hooks)

	c.mu.Lock()
	c.cleanupResources()
	c.mu.Unlock()

	return err
}

// executeShutdownHooks 执行关闭钩子
// 前置钩子按插入序执行,主列表按注册逆序执行(依赖序销毁:被依赖者最后销毁)
// 须在容器锁外调用,钩子内部可能获取锁;全部钩子执行完毕后聚合所有错误
func (c *container) executeShutdownHooks(ctx context.Context, preHooks, hooks []ShutdownHook) error {
	errs := make([]error, 0, len(preHooks)+len(hooks))
	for _, hook := range preHooks {
		if hookErr := hook(ctx); hookErr != nil {
			errs = append(errs, hookErr)
		}
	}
	for i := len(hooks) - 1; i >= 0; i-- {
		if hookErr := hooks[i](ctx); hookErr != nil {
			errs = append(errs, hookErr)
		}
	}
	return joinErrors(errs)
}

// cleanupResources 清理容器资源
// 重置所有内部状态
// 调用者必须已持有 c.mu.Lock
func (c *container) cleanupResources() {
	c.providers = make(map[cacheKey]providerEntry)
	c.instances = make(map[cacheKey]any)
	c.shutdown = nil
	c.preShutdown = nil
	c.started = false

	c.configMu.Lock()
	c.configSource = NewMapConfigSource()
	c.configMu.Unlock()

	// 使用 atomic 重置统计信息
	atomic.StoreInt64(&c.stats.createdInstances, 0)
	atomic.StoreInt64(&c.stats.getCalls, 0)
	atomic.StoreInt64(&c.stats.provideCalls, 0)
	atomic.StoreInt64(&c.stats.configHits, 0)
	atomic.StoreInt64(&c.stats.configMisses, 0)
	atomic.StoreInt64(&c.stats.createDuration, 0)

	close(c.done)
}

// signalNotify/osExit 信号监听与进程退出的注入点,仅测试替换使用
var (
	signalNotify = signal.Notify
	osExit       = os.Exit
)

// ShutdownOnSignals 监听系统信号并自动关闭
// 默认监听 SIGTERM 和 Interrupt 信号
// Shutdown 出错时以非零码退出,错误反映到进程退出状态
func (c *container) ShutdownOnSignals(signals ...os.Signal) {
	if len(signals) == 0 {
		signals = []os.Signal{syscall.SIGTERM, os.Interrupt}
	}

	sigChan := make(chan os.Signal, 1)
	signalNotify(sigChan, signals...)

	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal %s, shutting down...\n", sig)
		exitCode := 0
		if err := c.Shutdown(context.Background()); err != nil {
			fmt.Printf("Shutdown error: %v\n", err)
			exitCode = 1
		}
		osExit(exitCode)
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
		providers:    make(map[cacheKey]providerEntry),
		instances:    make(map[cacheKey]any),
		loading:      make(map[cacheKey]*loadingState),
		parent:       c,
		done:         make(chan struct{}),
		configSource: inheritedConfigSource,
	}
	// 锁内注册子容器关闭钩子,与 Shutdown 的锁内快照互斥
	c.mu.Lock()
	c.shutdown = append(c.shutdown, func(ctx context.Context) error { return c2.Shutdown(ctx) })
	c.mu.Unlock()
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
// 显式销毁钩子按既有行为注册;实现销毁接口的实例由容器持有时统一注册
func (c *container) registerDestroyHook(entry providerEntry, name string, instance any) {
	if !c.needsDestroyHook(entry, instance) {
		return
	}

	destroyHook := c.createDestroyHook(entry, name, instance)
	c.shutdown = append(c.shutdown, destroyHook)
}

// needsDestroyHook 判断实例是否需要注册销毁钩子
// Transient 实例不被容器持有,无显式钩子时不承担销毁职责
func (c *container) needsDestroyHook(entry providerEntry, instance any) bool {
	if c.hasDestroyHooks(entry) {
		return true
	}
	if entry.config.loadMode == LoadModeTransient {
		return false
	}
	return instanceNeedsDestroy(instance)
}

// instanceNeedsDestroy 检查实例是否实现销毁接口
func instanceNeedsDestroy(instance any) bool {
	if _, ok := instance.(DestroyCallback); ok {
		return true
	}
	_, ok := instance.(ServiceLifecycle)
	return ok
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
