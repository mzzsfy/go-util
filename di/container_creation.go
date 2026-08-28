// Package di 实现依赖注入容器的实例创建功能
package di

import (
	"fmt"
	"reflect"
	"time"

	gounsafe "github.com/mzzsfy/go-util/unsafe"
)

// create 主创建函数
// 协调各个创建阶段：钩子执行、缓存检查、实例创建
func (c *container) create(entry providerEntry, name string) (any, error) {
	startTime := time.Now()
	key := cacheKey{entry.reflectType, name}

	// 有 beforeCreate 钩子时执行,并重新检查缓存(钩子可能创建实例)
	if len(entry.config.beforeCreate) > 0 || len(c.beforeCreate) > 0 {
		instance, err := c.executeBeforeCreateHooks(entry, name)
		if err != nil {
			return nil, err
		}
		if instance != nil {
			return c.finalizeInstanceCreation(entry, name, instance, startTime)
		}
		if instance, found := c.checkAndGetCachedInstance(key); found {
			return instance, nil
		}
	}

	// 创建新实例
	return c.createNewInstance(entry, name, key, startTime)
}

// createNewInstance 创建新实例
// 以 singleflight 语义协调创建权:同 key 并发请求等待首个创建完成,同 goroutine 递归判定为循环依赖
func (c *container) createNewInstance(entry providerEntry, name string, key cacheKey, startTime time.Time) (any, error) {
	state := &loadingState{goid: gounsafe.GoID(), done: make(chan struct{})}
	for {
		wait, err := c.tryMarkLoading(key, state, entry.reflectType, name)
		if err != nil {
			return nil, err
		}
		if wait == nil {
			// 获得创建权,创建结束后通知等待方
			defer c.finishLoading(key, state)
			return c.runCreation(entry, name, key, startTime)
		}
		// 其他 goroutine 正在创建,等待创建结束后重试
		<-wait.done
		if instance, found := c.getCachedInstance(key); found {
			c.updateGetCallsStats()
			return instance, nil
		}
		// 创建失败或 Transient 未缓存,重新竞争创建权
	}
}

// runCreation 已获得创建权,执行实际创建流程
func (c *container) runCreation(entry providerEntry, name string, key cacheKey, startTime time.Time) (any, error) {
	// 懒加载模式先创建依赖
	if err := c.prepareLazyDependencies(entry); err != nil {
		return nil, err
	}

	// 调用 provider 构造实例
	instance, err := entry.provider(c)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, providerNilResultError(entry.reflectType, name)
	}

	// 双重检查：其他路径可能已创建实例
	if existingInstance, found := c.checkExistingInstanceDuringCreation(key, entry.config.loadMode); found {
		c.updateGetCallsStats()
		return existingInstance, nil
	}

	// 完成创建流程
	return c.finalizeInstanceCreation(entry, name, instance, startTime)
}

// finalizeInstanceCreation 完成实例创建的最后步骤
// 执行注入、缓存和钩子;afterCreate 失败时回滚缓存,半初始化实例不外泄
func (c *container) finalizeInstanceCreation(entry providerEntry, name string, instance any, startTime time.Time) (any, error) {
	// 验证并注入依赖
	instance, err := c.validateAndInject(instance)
	if err != nil || instance == nil {
		return instance, err
	}

	key := cacheKey{entry.reflectType, name}
	c.mu.Lock()
	c.cacheInstance(entry, name, instance)
	c.mu.Unlock()
	c.updateCreateStats(startTime)

	// 执行 afterCreate 钩子
	inst, err := c.executeAfterCreateHooks(entry, name, instance)
	if err != nil {
		// 回滚:移除缓存,销毁钩子尚未注册,无需清理
		c.mu.Lock()
		if _, exists := c.instances[key]; exists {
			delete(c.instances, key)
		}
		c.mu.Unlock()
		return nil, err
	}

	// afterCreate 成功后按最终实例注册销毁钩子
	c.mu.Lock()
	c.registerDestroyHook(entry, name, inst)
	c.mu.Unlock()

	return inst, nil
}

// executeBeforeCreateHooks 执行 beforeCreate 钩子
// 按顺序执行提供者级别和容器级别的钩子
func (c *container) executeBeforeCreateHooks(entry providerEntry, name string) (any, error) {
	info := EntryInfo{Name: name}

	inst, err := c.executeHookList(entry.config.beforeCreate, info, entry.reflectType, name, "未创建")
	if err != nil {
		return nil, err
	}
	if inst != nil {
		info.Instance = inst
	}

	inst, err = c.executeHookList(c.beforeCreate, info, entry.reflectType, name, "容器 beforeCreate 失败")
	if err != nil {
		return nil, err
	}
	if inst != nil {
		info.Instance = inst
	}
	return info.Instance, nil
}

// executeHookList 执行钩子列表
// 依次执行每个钩子，支持实例替换
func (c *container) executeHookList(hooks []func(Container, EntryInfo) (any, error), info EntryInfo, reflectType reflect.Type, name string, errorMsg string) (any, error) {
	instance := info.Instance
	for i, f := range hooks {
		v, err := f(c, info)
		if err != nil {
			return nil, hookError(errorMsg, i, reflectType, name, err)
		}
		if v != nil {
			instance = v
		}
	}
	return instance, nil
}

// executeAfterCreateHooks 执行 afterCreate 钩子
// 按顺序执行提供者级别和容器级别的钩子
func (c *container) executeAfterCreateHooks(entry providerEntry, name string, instance any) (any, error) {
	info := EntryInfo{Name: name, Instance: instance}

	inst, err := c.executeHookList(entry.config.afterCreate, info, entry.reflectType, name, "afterCreate")
	if err != nil {
		return nil, err
	}
	if inst != nil {
		info.Instance = inst
	}

	inst, err = c.executeHookList(c.afterCreate, info, entry.reflectType, name, "容器 afterCreate")
	if err != nil {
		return nil, err
	}
	if inst != nil {
		info.Instance = inst
	}
	return info.Instance, nil
}

// tryMarkLoading 竞争创建权
// 成功写入 state 返回 (nil, nil);同 key 已有创建者时:
//   - 创建者为本 goroutine 判定为循环依赖,返回错误
//   - 创建者为其他 goroutine 返回其状态,调用方等待 done 后重试
func (c *container) tryMarkLoading(key cacheKey, state *loadingState, reflectType reflect.Type, name string) (*loadingState, error) {
	c.mu.Lock()
	if existing, ok := c.loading[key]; ok {
		c.mu.Unlock()
		if existing.goid == state.goid {
			return nil, circularDependencyError(reflectType, name)
		}
		return existing, nil
	}
	c.loading[key] = state
	c.mu.Unlock()
	return nil, nil
}

// finishLoading 清除加载标记并通知等待方
func (c *container) finishLoading(key cacheKey, state *loadingState) {
	c.mu.Lock()
	delete(c.loading, key)
	c.mu.Unlock()
	close(state.done)
}

// checkExistingInstanceDuringCreation 检查创建过程中是否有其他路径已创建实例
// 用于并发场景下的双重检查
func (c *container) checkExistingInstanceDuringCreation(key cacheKey, loadMode LoadMode) (any, bool) {
	if loadMode == LoadModeTransient {
		return nil, false
	}
	return c.getCachedInstance(key)
}

// prepareLazyDependencies 准备懒加载依赖
// 对于懒加载模式，先创建所有依赖
func (c *container) prepareLazyDependencies(entry providerEntry) error {
	if entry.config.loadMode != LoadModeLazy {
		return nil
	}

	depend, err := c.findDepend(entry.reflectType)
	if err != nil {
		return err
	}

	return c.createDependencies(depend)
}

// createDependencies 创建依赖实例
// 按顺序创建所有依赖,依赖 provider 沿父链解析,实例缓存到定义所在容器
func (c *container) createDependencies(depend []cacheKey) error {
	for i, k := range depend {
		owner, entry, exists := c.findProviderOwner(k)
		if !exists {
			return fmt.Errorf("dependency[%d] not found: provider %s does not exist", i, k.String())
		}
		if _, err := owner.create(entry, k.name); err != nil {
			return fmt.Errorf("failed to create dependency[%d] %s: %w", i, k.String(), err)
		}
	}
	return nil
}
