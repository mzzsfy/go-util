// Package di 实现依赖注入容器的实例创建功能
package di

import (
	"fmt"
	"reflect"
	"time"
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
// 处理循环依赖检查、依赖准备和实例创建
func (c *container) createNewInstance(entry providerEntry, name string, key cacheKey, startTime time.Time) (any, error) {
	// 检查循环依赖并标记加载(合并为单次 Lock)
	if err := c.tryMarkLoading(key, entry.reflectType, name); err != nil {
		return nil, err
	}
	defer c.clearLoadingFlag(key)

	// 懒加载模式先创建依赖
	if err := c.prepareLazyDependencies(entry, key); err != nil {
		return nil, err
	}

	// 调用 provider 构造实例
	instance, err := entry.provider(c)
	if err != nil {
		return nil, err
	}

	// 双重检查：其他 goroutine 可能已创建实例
	if existingInstance, found := c.checkExistingInstanceDuringCreation(key, entry.config.loadMode); found {
		c.updateGetCallsStats()
		return existingInstance, nil
	}

	// 完成创建流程
	return c.finalizeInstanceCreation(entry, name, instance, startTime)
}

// finalizeInstanceCreation 完成实例创建的最后步骤
// 执行注入、缓存和钩子
func (c *container) finalizeInstanceCreation(entry providerEntry, name string, instance any, startTime time.Time) (any, error) {
	// 验证并注入依赖
	instance, err := c.validateAndInject(instance)
	if err != nil || instance == nil {
		return instance, err
	}

	// 缓存实例并注册销毁钩子
	c.cacheAndRegisterHooks(entry, name, instance)
	c.updateCreateStats(startTime)

	// 执行 afterCreate 钩子
	return c.executeAfterCreateHooks(entry, name, instance)
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

// tryMarkLoading 检查循环依赖并标记加载
// 如果正在加载说明存在循环依赖,否则标记并返回 nil
func (c *container) tryMarkLoading(key cacheKey, reflectType reflect.Type, name string) error {
	c.mu.Lock()
	if c.loading[key] {
		c.mu.Unlock()
		return circularDependencyError(reflectType, name)
	}
	c.loading[key] = true
	c.mu.Unlock()
	return nil
}

// checkExistingInstanceDuringCreation 检查创建过程中是否有其他 goroutine 已创建实例
// 用于并发场景下的双重检查
func (c *container) checkExistingInstanceDuringCreation(key cacheKey, loadMode LoadMode) (any, bool) {
	if loadMode == LoadModeTransient {
		return nil, false
	}
	return c.getCachedInstance(key)
}

// prepareLazyDependencies 准备懒加载依赖
// 对于懒加载模式，先创建所有依赖
func (c *container) prepareLazyDependencies(entry providerEntry, key cacheKey) error {
	if entry.config.loadMode != LoadModeLazy {
		return nil
	}

	depend, err := c.findDepend(entry.reflectType)
	if err != nil {
		c.clearLoadingFlag(key)
		return err
	}

	return c.createDependencies(depend)
}

// clearLoadingFlag 清除加载标记
func (c *container) clearLoadingFlag(key cacheKey) {
	c.mu.Lock()
	delete(c.loading, key)
	c.mu.Unlock()
}

// createDependencies 创建依赖实例
// 按顺序创建所有依赖
func (c *container) createDependencies(depend []cacheKey) error {
	for i, k := range depend {
		entry, exists := c.providers[k]
		if !exists {
			return fmt.Errorf("dependency[%d] not found: provider %s does not exist", i, k.String())
		}
		if _, err := c.create(entry, k.name); err != nil {
			return fmt.Errorf("failed to create dependency[%d] %s: %w", i, k.String(), err)
		}
	}
	return nil
}

