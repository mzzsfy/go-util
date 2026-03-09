// Package di 提供配置源实现
package di

import (
	"sync"

	"github.com/mzzsfy/go-util/config"
)

// mapConfigSource 基于 map 的配置源实现
// 提供线程安全的配置存储和访问
type mapConfigSource struct {
	// data 配置数据存储
	data map[string]any
	// mu 读写锁，保证线程安全
	mu sync.RWMutex
}

// NewMapConfigSource 创建基于 map 的配置源
// 返回可修改的配置源实例
func NewMapConfigSource() ConfigModifySource {
	return &mapConfigSource{
		data: make(map[string]any),
	}
}

// Get 获取配置值
// 参数:
//   - key: 配置键名
// 返回:
//   - 配置值包装器，如果键不存在返回 nil 值
func (m *mapConfigSource) Get(key string) config.Value {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return config.ValueFrom(m.data[key])
}

// Set 设置配置值
// 参数:
//   - key: 配置键名
//   - value: 配置值
func (m *mapConfigSource) Set(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Has 检查配置是否存在
// 参数:
//   - key: 配置键名
// 返回:
//   - 如果配置存在返回 true
func (m *mapConfigSource) Has(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.data[key]
	return exists
}

// Clear 清空所有配置
// 重置为空的配置映射
func (m *mapConfigSource) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]any)
}
