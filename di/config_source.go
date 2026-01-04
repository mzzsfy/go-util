package di

import (
    "sync"

    "github.com/mzzsfy/go-util/config"
)

// mapConfigSource 基于map的配置源实现
type mapConfigSource struct {
    data map[string]any
    mu   sync.RWMutex
}

// NewMapConfigSource 创建基于map的配置源
func NewMapConfigSource() ConfigModifySource {
    return &mapConfigSource{
        data: make(map[string]any),
    }
}

func (m *mapConfigSource) Get(key string) config.Value {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return config.ValueFrom(m.data[key])
}

func (m *mapConfigSource) Set(key string, value any) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.data[key] = value
}

func (m *mapConfigSource) Has(key string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    _, exists := m.data[key]
    return exists
}

func (m *mapConfigSource) Clear() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.data = make(map[string]any)
}
