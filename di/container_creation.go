// Package di 实现依赖注入容器的实例创建功能
package di

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// create 主创建函数
// 协调各个创建阶段：钩子执行、缓存检查、实例创建
func (c *container) create(entry providerEntry, name string) (any, error) {
	startTime := time.Now()
	key := typeKey(entry.reflectType, name)

	// 执行 beforeCreate 钩子
	instance, err := c.executeBeforeCreateHooks(entry, name)
	if err != nil {
		return nil, err
	}

	// 如果钩子返回了实例，直接使用
	if instance != nil {
		return c.finalizeInstanceCreation(entry, name, instance, startTime)
	}

	// 检查缓存
	if instance, found := c.checkAndGetCachedInstance(key); found {
		return instance, nil
	}

	// 创建新实例
	return c.createNewInstance(entry, name, key, startTime)
}

// tryGetCachedInstance 尝试获取缓存的实例
// 封装缓存检查逻辑
func (c *container) tryGetCachedInstance(key string) (any, bool) {
	return c.checkAndGetCachedInstance(key)
}

// createNewInstance 创建新实例
// 处理循环依赖检查、依赖准备和实例创建
func (c *container) createNewInstance(entry providerEntry, name string, key string, startTime time.Time) (any, error) {
	// 检查循环依赖
	if err := c.checkCircularDependency(key, entry.reflectType, name); err != nil {
		return nil, err
	}

	// 标记正在创建
	cleanup := c.markLoading(key)
	defer cleanup()

	// 创建实例及其依赖
	instance, err := c.createInstanceWithDependencies(entry, key)
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

// createInstanceWithDependencies 创建实例及其依赖
// 对于懒加载模式，先创建依赖
func (c *container) createInstanceWithDependencies(entry providerEntry, key string) (any, error) {
	if err := c.prepareLazyDependencies(entry, key); err != nil {
		return nil, err
	}

	return c.createProviderInstance(entry)
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

// hookChain 钩子链定义
// 包含钩子列表和错误消息
type hookChain struct {
	hooks    []func(Container, EntryInfo) (any, error)
	errorMsg string
}

// executeHookChains 执行多个钩子链
// 通用的钩子执行函数，支持实例传递
func (c *container) executeHookChains(chains []hookChain, info EntryInfo, reflectType reflect.Type, name string) (any, error) {
	instance := info.Instance
	for _, chain := range chains {
		inst, err := c.executeHookList(chain.hooks, info, reflectType, name, chain.errorMsg)
		if err != nil {
			return nil, err
		}
		if inst != nil {
			instance = inst
			info.Instance = inst
		}
	}
	return instance, nil
}

// executeBeforeCreateHooks 执行 beforeCreate 钩子
// 按顺序执行提供者级别和容器级别的钩子
func (c *container) executeBeforeCreateHooks(entry providerEntry, name string) (any, error) {
	chains := []hookChain{
		{entry.config.beforeCreate, "未创建"},
		{c.beforeCreate, "容器 beforeCreate 失败"},
	}
	info := EntryInfo{Instance: nil, Name: name}
	return c.executeHookChains(chains, info, entry.reflectType, name)
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
	chains := []hookChain{
		{entry.config.afterCreate, "afterCreate"},
		{c.afterCreate, "容器 afterCreate"},
	}
	info := EntryInfo{Name: name, Instance: instance}
	return c.executeHookChains(chains, info, entry.reflectType, name)
}

// checkCircularDependency 检查循环依赖
// 通过 loading map 检测正在创建的实例
func (c *container) checkCircularDependency(key string, reflectType reflect.Type, name string) error {
	c.mu.RLock()
	loading := c.loading[key]
	c.mu.RUnlock()
	if loading {
		return circularDependencyError(reflectType, name)
	}
	return nil
}

// markLoading 标记正在创建的实例
// 返回清理函数，用于 defer 调用
func (c *container) markLoading(key string) (cleanup func()) {
	c.mu.Lock()
	c.loading[key] = true
	c.mu.Unlock()

	return func() {
		c.mu.Lock()
		delete(c.loading, key)
		c.mu.Unlock()
	}
}

// checkExistingInstanceDuringCreation 检查创建过程中是否有其他 goroutine 已创建实例
// 用于并发场景下的双重检查
func (c *container) checkExistingInstanceDuringCreation(key string, loadMode LoadMode) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existingInstance, exists := c.instances[key]; exists && loadMode != LoadModeTransient {
		return existingInstance, true
	}
	return nil, false
}

// prepareLazyDependencies 准备懒加载依赖
// 对于懒加载模式，先创建所有依赖
func (c *container) prepareLazyDependencies(entry providerEntry, key string) error {
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
// 在错误发生时清理状态
func (c *container) clearLoadingFlag(key string) {
	c.mu.Lock()
	delete(c.loading, key)
	c.mu.Unlock()
}

// createDependencies 创建依赖实例
// 按顺序创建所有依赖
func (c *container) createDependencies(depend []string) error {
	for i, k := range depend {
		name := extractNameFromKey(k)
		entry, exists := c.providers[k]
		if !exists {
			return fmt.Errorf("dependency[%d] not found: provider %s does not exist", i, k)
		}
		if _, err := c.create(entry, name); err != nil {
			return fmt.Errorf("failed to create dependency[%d] %s: %w", i, k, err)
		}
	}
	return nil
}

// extractNameFromKey 从键中提取名称
// 键格式为 "类型#名称"
func extractNameFromKey(key string) string {
	ss := strings.SplitN(key, "#", 2)
	if len(ss) > 1 {
		return ss[1]
	}
	return ""
}

// createProviderInstance 通过 provider 函数创建实例
// 调用注册时提供的构造函数
func (c *container) createProviderInstance(entry providerEntry) (any, error) {
	return entry.provider(c)
}

// validateInstance 验证实例是否有效
// 检查实例非空且反射值有效
func (c *container) validateInstance(instance any, instanceValue reflect.Value) bool {
	return instance != nil && instanceValue.IsValid()
}
