package storage

import (
	"testing"
)

// mockCache 简单的内存缓存实现,用于测试
type mockCache[K comparable, V any] struct {
	data map[K]V
}

func newMockCache[K comparable, V any]() *mockCache[K, V] {
	return &mockCache[K, V]{data: make(map[K]V)}
}

func (m *mockCache[K, V]) Get(key K) (V, bool) {
	v, ok := m.data[key]
	return v, ok
}

func (m *mockCache[K, V]) Set(key K, value V) {
	m.data[key] = value
}

func (m *mockCache[K, V]) Delete(key K) {
	delete(m.data, key)
}

func (m *mockCache[K, V]) Clear() {
	m.data = make(map[K]V)
}

func (m *mockCache[K, V]) Size() int {
	return len(m.data)
}

// TestCacheWrap_GetOr_Hit 测试缓存命中场景
func TestCacheWrap_GetOr_Hit(t *testing.T) {
	t.Parallel()
	cache := NewCacheWrap[string, int](newMockCache[string, int]())
	cache.Set("key1", 100)

	// 缓存命中,不应该调用 def 函数
	result := cache.GetOr("key1", func() int {
		t.Error("不应该调用 def 函数")
		return 0
	})
	if result != 100 {
		t.Errorf("期望 100, 实际 %d", result)
	}
}

// TestCacheWrap_GetOr_Miss 测试缓存未命中场景
func TestCacheWrap_GetOr_Miss(t *testing.T) {
	t.Parallel()
	cache := NewCacheWrap[string, int](newMockCache[string, int]())

	callCount := 0
	result := cache.GetOr("key1", func() int {
		callCount++
		return 200
	})
	if result != 200 {
		t.Errorf("期望 200, 实际 %d", result)
	}
	if callCount != 1 {
		t.Errorf("期望调用 1 次, 实际 %d 次", callCount)
	}

	// 再次获取应该命中缓存,不再调用 def
	result = cache.GetOr("key1", func() int {
		t.Error("不应该再次调用 def 函数")
		return 0
	})
	if result != 200 {
		t.Errorf("期望 200, 实际 %d", result)
	}
}

// TestCacheWrap_GetOr_Concurrent 测试并发场景下的 double-check
func TestCacheWrap_GetOr_Concurrent(t *testing.T) {
	t.Parallel()
	cache := NewCacheWrap[int, int](newMockCache[int, int]())

	// 启动多个 goroutine 同时获取同一个 key
	done := make(chan int, 10)
	for i := 0; i < 10; i++ {
		go func() {
			result := cache.GetOr(1, func() int {
				return 42
			})
			done <- result
		}()
	}

	// 所有结果应该是正确的值
	for i := 0; i < 10; i++ {
		if result := <-done; result != 42 {
			t.Errorf("期望 42, 实际 %d", result)
		}
	}

	// 缓存应该只有一个条目
	if cache.Size() != 1 {
		t.Errorf("期望缓存大小 1, 实际 %d", cache.Size())
	}
}

// TestCacheWrap_NewCacheWrap 测试构造函数
func TestCacheWrap_NewCacheWrap(t *testing.T) {
	t.Parallel()
	mock := newMockCache[string, string]()
	wrap := NewCacheWrap[string, string](mock)

	if wrap == nil {
		t.Error("NewCacheWrap 不应返回 nil")
		return
	}

	// 验证底层 cache 正确注入
	wrap.Set("test", "value")
	if v, ok := mock.Get("test"); !ok || v != "value" {
		t.Error("CacheWrap 应该正确代理到底层 cache")
	}
}
