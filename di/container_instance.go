// Package di 提供实例管理功能
package di

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// getServiceNameFromKey 从 key 中提取服务名称
// key 格式为 "类型名#名称"，返回名称部分
func getServiceNameFromKey(key, defaultTypeName string) string {
	if key == defaultTypeName {
		return ""
	}
	_, serviceName, _ := strings.Cut(key, "#")
	return serviceName
}

// collectMatchingInstances 收集匹配类型的实例
// 遍历所有提供者，找出类型兼容的实例
func (c *container) collectMatchingInstances(t reflect.Type, typeName string) (map[string]any, error) {
	results := make(map[string]any)

	for key, entry := range c.providers {
		if !entry.reflectType.AssignableTo(t) {
			continue
		}

		if err := c.collectSingleInstance(results, key, entry, typeName); err != nil {
			return nil, err
		}
	}

	return results, nil
}

// collectSingleInstance 收集单个匹配的实例
// 处理条件失败等特殊情况
func (c *container) collectSingleInstance(results map[string]any, key string, entry providerEntry, typeName string) error {
	serviceName := getServiceNameFromKey(key, typeName)

	instance, err := c.GetNamed(entry.reflectType, serviceName)
	if err != nil {
		if errors.Is(err, ErrorConditionFail) {
			return nil
		}
		return fmt.Errorf("failed to collect instance %s: %w", key, err)
	}

	results[key] = instance
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

	results, err := c.collectMatchingInstances(t, t.String())
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
// 返回实例的副本，线程安全
func (c *container) GetAllInstances() map[string]any {
	return withRLockResult(c, func() map[string]any {
		results := make(map[string]any, len(c.instances))
		for k, v := range c.instances {
			results[k] = v
		}
		return results
	})
}

// GetProviders 获取所有注册的提供者信息
// 返回类型名称映射，线程安全
func (c *container) GetProviders() map[string]string {
	return withRLockResult(c, func() map[string]string {
		results := make(map[string]string, len(c.providers))
		for k, entry := range c.providers {
			results[k] = entry.reflectType.String()
		}
		return results
	})
}

// ReplaceInstance 运行时替换已注册的服务实例
// 用于测试或热更新场景
func (c *container) ReplaceInstance(serviceType any, name string, newInstance any) error {
	t := parseReflectType(serviceType)
	key := typeKey(t, name)

	return withWLockResult(c, func() error {
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
	})
}

// RemoveInstance 移除已缓存的实例
// 不会移除提供者注册
func (c *container) RemoveInstance(serviceType any, name string) error {
	t := parseReflectType(serviceType)
	key := typeKey(t, name)

	c.withWLock(func() {
		delete(c.instances, key)
	})
	return nil
}

// ClearInstances 清空所有缓存的实例
// 不会清空提供者注册
func (c *container) ClearInstances() {
	c.withWLock(func() {
		c.instances = make(map[string]any)
	})
}

// GetInstanceCount 获取当前缓存的实例数量
func (c *container) GetInstanceCount() int {
	return withRLockResult(c, func() int {
		return len(c.instances)
	})
}

// GetProviderCount 获取注册的提供者数量
func (c *container) GetProviderCount() int {
	return withRLockResult(c, func() int {
		return len(c.providers)
	})
}
