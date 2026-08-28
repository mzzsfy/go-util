// Package di 提供实例管理功能
package di

import (
	"errors"
	"fmt"
	"reflect"
)

// collectMatchingInstances 收集匹配类型的实例
// 先快照匹配的 provider,锁外创建实例,避免遍历期间并发注册修改 map
func (c *container) collectMatchingInstances(t reflect.Type) (map[string]any, error) {
	c.mu.RLock()
	matches := make([]cacheKey, 0)
	entries := make([]providerEntry, 0)
	for key, entry := range c.providers {
		if entry.reflectType.AssignableTo(t) {
			matches = append(matches, key)
			entries = append(entries, entry)
		}
	}
	c.mu.RUnlock()

	results := make(map[string]any, len(matches))
	for i, key := range matches {
		if err := c.collectSingleInstance(results, key, entries[i]); err != nil {
			return nil, err
		}
	}

	return results, nil
}

// collectSingleInstance 收集单个匹配的实例
// 处理条件失败等特殊情况
func (c *container) collectSingleInstance(results map[string]any, key cacheKey, entry providerEntry) error {
	instance, err := c.GetNamed(entry.reflectType, key.name)
	if err != nil {
		if errors.Is(err, ErrorConditionFail) {
			return nil
		}
		return fmt.Errorf("failed to collect instance %s: %w", key.String(), err)
	}

	results[key.String()] = instance
	return nil
}

// mergeParentResults 合并父容器的结果
// 将父容器中匹配的实例合并到结果集
func (c *container) mergeParentResults(results map[string]any, serviceType any) error {
	if c.parent == nil {
		return nil
	}

	parentResults, err := c.parent.GetNamedAll(serviceType)
	if err != nil {
		return fmt.Errorf("failed to get instances from parent container: %w", err)
	}

	for k, v := range parentResults {
		results[k] = v
	}

	return nil
}

// GetNamedAll 获取指定类型的所有命名实例
// 支持接口匹配，性能较低
func (c *container) GetNamedAll(serviceType any) (map[string]any, error) {
	t := parseReflectType(serviceType)

	if err := c.checkBlacklistForGetAll(t); err != nil {
		return nil, err
	}

	results, err := c.collectMatchingInstances(t)
	if err != nil {
		return nil, err
	}

	if err := c.mergeParentResults(results, serviceType); err != nil {
		return nil, err
	}

	return results, nil
}

// checkBlacklistForGetAll 检查 GetNamedAll 的黑名单类型
// 某些类型不允许使用 GetNamedAll
func (c *container) checkBlacklistForGetAll(t reflect.Type) error {
	if isBlacklistType(t) {
		return fmt.Errorf("cannot use GetNamedAll for type %s", t)
	}
	return nil
}

// GetAllInstances 获取所有已缓存的实例
func (c *container) GetAllInstances() map[string]any {
	c.mu.RLock()
	results := make(map[string]any, len(c.instances))
	for k, v := range c.instances {
		results[k.String()] = v
	}
	c.mu.RUnlock()
	return results
}

// GetProviders 获取所有注册的提供者信息
func (c *container) GetProviders() map[string]string {
	c.mu.RLock()
	results := make(map[string]string, len(c.providers))
	for k, entry := range c.providers {
		results[k.String()] = entry.reflectType.String()
	}
	c.mu.RUnlock()
	return results
}

// ReplaceInstance 运行时替换已注册的服务实例
func (c *container) ReplaceInstance(serviceType any, name string, newInstance any) error {
	t := parseReflectType(serviceType)
	key := cacheKey{t, name}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.providers[key]
	if !exists {
		return fmt.Errorf("cannot replace instance: no provider registered for type %s with name '%s'", t, name)
	}

	if newInstance != nil {
		newInstanceType := reflect.TypeOf(newInstance)
		if !newInstanceType.AssignableTo(entry.reflectType) {
			return fmt.Errorf("cannot replace instance: new instance type %s is not assignable to %s",
				newInstanceType, entry.reflectType)
		}
	}

	c.instances[key] = newInstance
	return nil
}

// RemoveInstance 移除已缓存的实例
func (c *container) RemoveInstance(serviceType any, name string) error {
	t := parseReflectType(serviceType)
	key := cacheKey{t, name}

	c.mu.Lock()
	delete(c.instances, key)
	c.mu.Unlock()
	return nil
}

// ClearInstances 清空所有缓存的实例
func (c *container) ClearInstances() {
	c.mu.Lock()
	c.instances = make(map[cacheKey]any)
	c.mu.Unlock()
}

// GetInstanceCount 获取当前缓存的实例数量
func (c *container) GetInstanceCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.instances)
}

// GetProviderCount 获取注册的提供者数量
func (c *container) GetProviderCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.providers)
}
