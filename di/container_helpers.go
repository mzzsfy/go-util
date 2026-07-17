// Package di 提供容器辅助函数和错误处理
package di

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/mzzsfy/go-util/config"
)

// ========== 锁辅助函数 ==========

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

// updateGetCallsStats 更新 GetCalls 统计
func (c *container) updateGetCallsStats() {
	atomic.AddInt64(&c.stats.getCalls, 1)
}

// updateCreateStats 更新创建统计信息
func (c *container) updateCreateStats(startTime time.Time) {
	atomic.AddInt64(&c.stats.createdInstances, 1)
	atomic.AddInt64(&c.stats.getCalls, 1)
	atomic.AddInt64(&c.stats.createDuration, int64(time.Since(startTime)))
}

// ========== 缓存辅助函数 ==========

// getCachedInstance 获取缓存的实例
// 返回实例和是否存在标记，区分 nil 实例和缓存未命中
func (c *container) getCachedInstance(key string) (any, bool) {
	c.mu.RLock()
	instance, exists := c.instances[key]
	c.mu.RUnlock()
	return instance, exists
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
// 使用缓存的 entry.key 避免 typeKey 重复计算
func (c *container) cacheInstance(entry providerEntry, name string, instance any) {
	if entry.config.loadMode != LoadModeTransient {
		// 优先使用缓存的 key
		key := entry.key
		if key == "" {
			key = typeKey(entry.reflectType, name)
		}
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
	return ContainerStats{
		CreatedInstances: int(atomic.LoadInt64(&c.stats.createdInstances)),
		GetCalls:         int(atomic.LoadInt64(&c.stats.getCalls)),
		ProvideCalls:     int(atomic.LoadInt64(&c.stats.provideCalls)),
		ConfigHits:       int(atomic.LoadInt64(&c.stats.configHits)),
		ConfigMisses:     int(atomic.LoadInt64(&c.stats.configMisses)),
		CreateDuration:   time.Duration(atomic.LoadInt64(&c.stats.createDuration)),
	}
}

// ResetStats 重置统计信息
func (c *container) ResetStats() {
	atomic.StoreInt64(&c.stats.createdInstances, 0)
	atomic.StoreInt64(&c.stats.getCalls, 0)
	atomic.StoreInt64(&c.stats.provideCalls, 0)
	atomic.StoreInt64(&c.stats.configHits, 0)
	atomic.StoreInt64(&c.stats.configMisses, 0)
	atomic.StoreInt64(&c.stats.createDuration, 0)
}

// GetAverageCreateDuration 获取平均创建耗时
func (c *container) GetAverageCreateDuration() time.Duration {
	created := atomic.LoadInt64(&c.stats.createdInstances)
	if created == 0 {
		return 0
	}
	duration := atomic.LoadInt64(&c.stats.createDuration)
	return time.Duration(duration) / time.Duration(created)
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
