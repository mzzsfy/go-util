// Package di 提供配置源实现
package di

import (
	"sync"

	"github.com/mzzsfy/go-util/config"
)

// mapConfigSource 基于 map 的配置源实现
// 提供线程安全的配置存储和访问
type mapConfigSource struct {
	// data 配置数据存储,Set 时预创建 config.Value 避免每次 Get 分配
	data map[string]config.Value
	// mu 读写锁，保证线程安全
	mu sync.RWMutex
}

// NewMapConfigSource 创建基于 map 的配置源
// 返回可修改的配置源实例
func NewMapConfigSource() ConfigModifySource {
	return &mapConfigSource{
		data: make(map[string]config.Value),
	}
}

// Get 获取配置值
func (m *mapConfigSource) Get(key string) config.Value {
	m.mu.RLock()
	v, exists := m.data[key]
	m.mu.RUnlock()
	if exists {
		return v
	}
	return config.ValueFrom(nil)
}

// Set 设置配置值
func (m *mapConfigSource) Set(key string, value any) {
	m.mu.Lock()
	m.data[key] = config.ValueFrom(value)
	m.mu.Unlock()
}

// Has 检查配置是否存在
func (m *mapConfigSource) Has(key string) bool {
	m.mu.RLock()
	_, exists := m.data[key]
	m.mu.RUnlock()
	return exists
}

// Clear 清空所有配置
func (m *mapConfigSource) Clear() {
	m.mu.Lock()
	m.data = make(map[string]config.Value)
	m.mu.Unlock()
}
