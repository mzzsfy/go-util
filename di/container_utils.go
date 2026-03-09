// Package di 提供容器核心工具函数
// 包含服务获取、父容器查找、存在性检查和选项追加等核心功能
package di

import (
	"fmt"
	"reflect"
)

// GetNamed 获取命名服务实例
// 这是服务获取的主要入口，支持缓存和父容器查找
// 参数:
//   - serviceType: 服务类型，可以是 reflect.Type 或具体类型
//   - name: 服务名称，为空时使用默认名称
// 返回:
//   - 服务实例
//   - 可能的错误
func (c *container) GetNamed(serviceType any, name string) (any, error) {
	t := parseReflectType(serviceType)
	key := typeKey(t, name)

	// 首先检查缓存
	if instance := c.getCachedInstance(key); instance != nil {
		return instance, nil
	}

	// 检查提供者
	entry, exists := c.getProvider(key)
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

// HasNamed 检查服务是否存在
// 会同时检查当前容器和父容器
// 参数:
//   - serviceType: 服务类型
//   - name: 服务名称
// 返回:
//   - 如果服务已注册返回 true
func (c *container) HasNamed(serviceType any, name string) bool {
	t := parseReflectType(serviceType)
	key := typeKey(t, name)

	c.mu.RLock()
	defer c.mu.RUnlock()

	// 检查当前容器
	if _, exists := c.providers[key]; exists {
		return true
	}
	// 检查父容器
	if c.parent != nil {
		return c.parent.HasNamed(serviceType, name)
	}
	return false
}

// AppendOption 追加容器选项
// 启动后不可使用，会 panic
// 参数:
//   - opt: 容器选项列表
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
