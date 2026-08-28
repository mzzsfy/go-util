package storage

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Delete(key K)
	Clear()
	Size() int
}

type TimedCache[K comparable, V any] interface {
	Cache[K, V]
	SetWithTimeout(key K, value V, timeout time.Duration)
	TTL(key K) time.Duration
}

// CacheWrap 对底层 Cache 的 GetOr 加载路径做加锁包装。
// 并发契约: GetOr 在锁内调用底层 Cache 的 Get/Set, 且 CacheWrap 的其余方法
// (Get/Set/Delete/Clear/Size)不加锁直接透传, 因此底层 Cache 的实现必须自身并发安全,
// 否则 GetOr 与其他方法并发调用会产生数据竞争。
type CacheWrap[K comparable, V any] struct {
	lock sync.Mutex
	Cache[K, V]
}

// GetOr 命中直接返回; 未命中时加锁 double-check 后调用 def 生成并写入
// def 在持有内部锁的状态下执行, 不应回调本对象的其他方法(会重入死锁), 且必须无副作用
func (c *CacheWrap[K, V]) GetOr(key K, def func() V) V {
	get, b := c.Cache.Get(key)
	if b {
		return get
	}
	c.lock.Lock()
	if get, b = c.Cache.Get(key); b {
		c.lock.Unlock()
		return get
	}
	v := def()
	c.Cache.Set(key, v)
	c.lock.Unlock()
	return v
}

func NewCacheWrap[K comparable, V any](cache Cache[K, V]) *CacheWrap[K, V] {
	return &CacheWrap[K, V]{Cache: cache}
}
