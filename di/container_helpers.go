// Package di 提供容器辅助函数和错误处理
package di

import (
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mzzsfy/go-util/config"
)

// ========== 类型解析函数 ==========

// parseReflectType 解析类型，支持 reflect.Type 和普通类型
func parseReflectType(serviceType any) reflect.Type {
	if t, ok := serviceType.(reflect.Type); ok {
		return t
	}
	return reflect.TypeOf(serviceType)
}

// ========== 错误创建函数 ==========

// circularDependencyError 循环依赖错误
func circularDependencyError(t reflect.Type, name string) error {
	return fmt.Errorf("circular dependency detected for type %s with name '%s'", t, name)
}

// providerNotFoundError 提供者未找到错误
func providerNotFoundError(t reflect.Type, name string) error {
	return fmt.Errorf("no provider found for type %s with name '%s'", t, name)
}

// providerExistsError 提供者已存在错误
func providerExistsError(t reflect.Type, name string) error {
	return fmt.Errorf("provider already registered: type %s with name '%s' exists", t, name)
}

// hookError 钩子执行错误
func hookError(hookType string, index int, t reflect.Type, name string, err error) error {
	return fmt.Errorf("%s hook[%d] failed for type %s with name '%s': %w", hookType, index, t, name, err)
}

// fieldInjectionError 字段注入错误
func fieldInjectionError(fieldName string, err error) error {
	return fmt.Errorf("failed to inject field %s: %w", fieldName, err)
}

// conversionError 类型转换错误
func conversionError(fromType, toType reflect.Type) error {
	return fmt.Errorf("cannot convert value from type %s to field type %s", fromType, toType)
}

// destroyCallbackError 销毁回调错误
func destroyCallbackError(t reflect.Type, name string, err error) error {
	return fmt.Errorf("DestroyCallback failed for %s with name '%s': %w", t, name, err)
}

// shutdownError 关闭错误
func shutdownError(t reflect.Type, name string, err error) error {
	return fmt.Errorf("shutdown failed for %s with name '%s': %w", t, name, err)
}

// providerNilResultError provider 返回 nil 实例错误
func providerNilResultError(t reflect.Type, name string) error {
	return fmt.Errorf("provider returned nil instance for type %s with name '%s'", t, name)
}

// ========== 多错误聚合 ==========

// multiError 聚合多个错误
// Unwrap 返回全部错误,支持 errors.Is/As 逐个匹配
// 已知局限: Unwrap() []error 需 go1.20+ 的 errors 包才会被遍历, go1.18 下 errors.Is/As 无法逐个匹配
type multiError struct {
	errs []error
}

// Error 拼接全部错误信息
func (m *multiError) Error() string {
	var sb strings.Builder
	sb.WriteString("multiple errors: ")
	for i, err := range m.errs {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Unwrap 返回全部错误供 errors.Is/As 遍历
func (m *multiError) Unwrap() []error {
	return m.errs
}

// joinErrors 聚合错误
// 全为 nil 返回 nil;单个非 nil 原样返回;多个时聚合为 multiError
func joinErrors(errs []error) error {
	count := 0
	for _, err := range errs {
		if err != nil {
			count++
		}
	}
	if count == 0 {
		return nil
	}
	if count == 1 {
		for _, err := range errs {
			if err != nil {
				return err
			}
		}
	}
	merged := make([]error, 0, count)
	for _, err := range errs {
		if err != nil {
			merged = append(merged, err)
		}
	}
	return &multiError{errs: merged}
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

// getCachedInstance RLock 读取缓存实例
func (c *container) getCachedInstance(key cacheKey) (any, bool) {
	c.mu.RLock()
	instance, exists := c.instances[key]
	c.mu.RUnlock()
	return instance, exists
}

// checkAndGetCachedInstance 获取缓存的实例
func (c *container) checkAndGetCachedInstance(key cacheKey) (any, bool) {
	if instance, found := c.getCachedInstance(key); found {
		c.updateGetCallsStats()
		return instance, true
	}
	return nil, false
}

// storeInstance 加锁写入实例（独立调用）
func (c *container) storeInstance(key cacheKey, instance any) {
	c.mu.Lock()
	c.instances[key] = instance
	c.mu.Unlock()
}

// cacheInstance 缓存实例，调用者必须已持有 c.mu.Lock
func (c *container) cacheInstance(entry providerEntry, name string, instance any) {
	if entry.config.loadMode != LoadModeTransient {
		c.instances[cacheKey{entry.reflectType, name}] = instance
	}
}

// cacheAndRegisterHooks 缓存实例并注册销毁钩子
func (c *container) cacheAndRegisterHooks(entry providerEntry, name string, instance any) {
	c.mu.Lock()
	c.cacheInstance(entry, name, instance)
	c.registerDestroyHook(entry, name, instance)
	c.mu.Unlock()
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
	if source == nil {
		c.configMu.Unlock()
		panic("configSource cannot be nil")
	}
	c.configSource = source
	c.configMu.Unlock()
}

// GetConfigSource 获取当前配置源
func (c *container) GetConfigSource() ConfigSource {
	c.configMu.RLock()
	source := c.configSource
	c.configMu.RUnlock()
	return source
}

// Value 获取配置值
func (c *container) Value(key string) config.Value {
	return c.getConfigValue(key)
}
