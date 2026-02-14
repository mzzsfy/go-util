// Package di 提供容器辅助函数和错误处理
package di

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mzzsfy/go-util/config"
)

// ========== 锁辅助函数 ==========

// withRLock 使用读锁执行函数
func (c *container) withRLock(fn func()) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	fn()
}

// withWLock 使用写锁执行函数
func (c *container) withWLock(fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	fn()
}

// withRLockResult 使用读锁执行函数并返回结果
func withRLockResult[T any](c *container, fn func() T) T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fn()
}

// withWLockResult 使用写锁执行函数并返回结果
func withWLockResult[T any](c *container, fn func() T) T {
	c.mu.Lock()
	defer c.mu.Unlock()
	return fn()
}

// ========== 类型解析函数 ==========

// parseReflectType 解析类型，支持 reflect.Type 和普通类型
func parseReflectType(serviceType any) reflect.Type {
	if t, ok := serviceType.(reflect.Type); ok {
		return t
	}
	return reflect.TypeOf(serviceType)
}

// typeKey 生成类型键，格式为 "类型名#名称"
func typeKey(t reflect.Type, name string) string {
	if name != "" {
		return t.String() + "#" + name
	}
	return t.String()
}

// ========== 错误创建函数 ==========

// makeError 创建格式化错误
func makeError(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// circularDependencyError 循环依赖错误
func circularDependencyError(t reflect.Type, name string) error {
	return makeError("circular dependency detected for type %s with name '%s'", t, name)
}

// providerNotFoundError 提供者未找到错误
func providerNotFoundError(t reflect.Type, name string) error {
	return makeError("no provider found for type %s with name '%s'", t, name)
}

// providerExistsError 提供者已存在错误
func providerExistsError(t reflect.Type, name string) error {
	return makeError("provider already registered: type %s with name '%s' exists", t, name)
}

// hookError 钩子执行错误
func hookError(hookType string, index int, t reflect.Type, name string, err error) error {
	return makeError("%s hook[%d] failed for type %s with name '%s': %w", hookType, index, t, name, err)
}

// fieldInjectionError 字段注入错误
func fieldInjectionError(fieldName string, err error) error {
	return makeError("failed to inject field %s: %w", fieldName, err)
}

// conversionError 类型转换错误
func conversionError(fromType, toType reflect.Type) error {
	return makeError("cannot convert value from type %s to field type %s", fromType, toType)
}

// destroyCallbackError 销毁回调错误
func destroyCallbackError(t reflect.Type, name string, err error) error {
	return makeError("DestroyCallback failed for %s with name '%s': %w", t, name, err)
}

// shutdownError 关闭错误
func shutdownError(t reflect.Type, name string, err error) error {
	return makeError("shutdown failed for %s with name '%s': %w", t, name, err)
}

// ========== 统计更新函数 ==========

// updateStats 通用统计更新函数
func (c *container) updateStats(updateFn func(stats *containerStats)) {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()
	updateFn(&c.stats)
}

// updateGetCallsStats 更新 GetCalls 统计
func (c *container) updateGetCallsStats() {
	c.updateStats(func(stats *containerStats) {
		stats.getCalls++
	})
}

// updateCreateStats 更新创建统计信息
func (c *container) updateCreateStats(startTime time.Time) {
	c.updateStats(func(stats *containerStats) {
		stats.createdInstances++
		stats.getCalls++
		stats.createDuration += time.Since(startTime)
	})
}

// ========== 缓存辅助函数 ==========

// getCachedInstance 获取缓存的实例
func (c *container) getCachedInstance(key string) any {
	return withRLockResult(c, func() any { return c.instances[key] })
}

// getProvider 获取提供者
func (c *container) getProvider(key string) (providerEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, exists := c.providers[key]
	return entry, exists
}

// checkCacheWithReadLock 使用读锁检查缓存
func (c *container) checkCacheWithReadLock(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if instance, exists := c.instances[key]; exists {
		c.updateGetCallsStats()
		return instance, true
	}
	return nil, false
}

// checkCacheWithWriteLock 使用写锁检查缓存和循环依赖
func (c *container) checkCacheWithWriteLock(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if instance, exists := c.instances[key]; exists {
		c.updateGetCallsStats()
		return instance, true
	}
	if c.loading[key] {
		return nil, false
	}
	return nil, false
}

// checkAndGetCachedInstance 双重检查并获取缓存的实例
func (c *container) checkAndGetCachedInstance(key string) (any, bool) {
	if instance, found := c.checkCacheWithReadLock(key); found {
		return instance, true
	}
	return c.checkCacheWithWriteLock(key)
}

// cacheInstance 缓存实例，非 Transient 模式下才会缓存
func (c *container) cacheInstance(entry providerEntry, name string, instance any) {
	if entry.config.loadMode != LoadModeTransient {
		key := typeKey(entry.reflectType, name)
		c.instances[key] = instance
	}
}

// cacheAndRegisterHooks 缓存实例并注册销毁钩子
func (c *container) cacheAndRegisterHooks(entry providerEntry, name string, instance any) {
	c.withWLock(func() {
		c.cacheInstance(entry, name, instance)
		c.registerDestroyHook(entry, name, instance)
	})
}

// ========== 统计接口实现 ==========

// GetStats 获取容器统计信息
func (c *container) GetStats() ContainerStats {
	c.statsMu.RLock()
	defer c.statsMu.RUnlock()
	return ContainerStats{
		CreatedInstances: c.stats.createdInstances,
		GetCalls:         c.stats.getCalls,
		ProvideCalls:     c.stats.provideCalls,
		ConfigHits:       c.stats.configHits,
		ConfigMisses:     c.stats.configMisses,
		CreateDuration:   c.stats.createDuration,
	}
}

// ResetStats 重置统计信息
func (c *container) ResetStats() {
	c.statsMu.Lock()
	defer c.statsMu.Unlock()
	c.stats = containerStats{}
}

// GetAverageCreateDuration 获取平均创建耗时
func (c *container) GetAverageCreateDuration() time.Duration {
	c.statsMu.RLock()
	defer c.statsMu.RUnlock()
	if c.stats.createdInstances == 0 {
		return 0
	}
	return c.stats.createDuration / time.Duration(c.stats.createdInstances)
}

// ========== 配置接口实现 ==========

// SetConfigSource 设置配置源
func (c *container) SetConfigSource(source ConfigSource) {
	c.configMu.Lock()
	defer c.configMu.Unlock()
	if source == nil {
		panic("configSource cannot be nil")
	}
	c.configSource = source
}

// GetConfigSource 获取当前配置源
func (c *container) GetConfigSource() ConfigSource {
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	return c.configSource
}

// Value 获取配置值
func (c *container) Value(key string) config.Value {
	return c.getConfigValue(key)
}
