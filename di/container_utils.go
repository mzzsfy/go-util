// Package di 提供容器核心工具函数
// 包含服务获取、父容器查找、存在性检查和选项追加等核心功能
package di

import (
	"fmt"
	"reflect"
)

// GetNamed 获取命名服务实例
func (c *container) GetNamed(serviceType any, name string) (any, error) {
	return c.getNamedByType(parseReflectType(serviceType), name)
}

// getNamedByType 以 reflect.Type 直接查找,跳过 parseReflectType
func (c *container) getNamedByType(t reflect.Type, name string) (any, error) {
	key := cacheKey{t, name}

	// 无锁读取缓存实例
	if instance, found := c.getCachedInstance(key); found {
		return instance, nil
	}

	// 缓存未命中,RLock 查找 provider
	c.mu.RLock()
	entry, exists := c.providers[key]
	c.mu.RUnlock()

	if !exists {
		return c.getFromParentOrError(t, name)
	}
	return c.create(entry, name)
}

// getFromParentOrError 从父容器获取或返回错误
// 当本地容器找不到提供者时，尝试从父容器查找
// 参数:
//   - t: 反射类型
//   - name: 服务名称
//
// 返回:
//   - 服务实例
//   - 可能的错误
func (c *container) getFromParentOrError(t reflect.Type, name string) (any, error) {
	if c.parent != nil {
		instance, err := c.parent.GetNamed(t, name)
		if err != nil {
			return nil, fmt.Errorf("parent container failed to provide %s with name '%s': %w", t, name, err)
		}
		return instance, nil
	}
	return nil, providerNotFoundError(t, name)
}

// findProviderOwner 沿父链查找 provider 定义所在容器
// 用于依赖解析等需要跨作用域定位 provider 的场景
func (c *container) findProviderOwner(key cacheKey) (*container, providerEntry, bool) {
	for cur := c; cur != nil; cur = cur.parent {
		cur.mu.RLock()
		entry, ok := cur.providers[key]
		cur.mu.RUnlock()
		if ok {
			return cur, entry, true
		}
	}
	return nil, providerEntry{}, false
}

// HasNamed 检查服务是否存在
// 会同时检查当前容器和父容器
// 参数:
//   - serviceType: 服务类型
//   - name: 服务名称
//
// 返回:
//   - 如果服务已注册返回 true
func (c *container) HasNamed(serviceType any, name string) bool {
	t := parseReflectType(serviceType)
	key := cacheKey{t, name}

	c.mu.RLock()
	_, exists := c.providers[key]
	parent := c.parent
	c.mu.RUnlock()

	if exists {
		return true
	}
	if parent != nil {
		return parent.HasNamed(serviceType, name)
	}
	return false
}

// AppendOption 追加容器选项
// 启动后不可使用，会 panic
// 参数:
//   - opt: 容器选项列表
//
// 返回:
//   - 可能的错误（捕获 panic）
func (c *container) AppendOption(opt ...ContainerOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while appending options: %v", r)
		}
	}()
	for _, option := range opt {
		option(c)
	}
	return
}
