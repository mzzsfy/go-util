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

type CacheWrap[K comparable, V any] struct {
    lock sync.Mutex
    Cache[K, V]
}

func (c *CacheWrap[K, V]) GetOr(key K, def func() V) V {
    get, b := c.Cache.Get(key)
    if b {
        return get
    }
    c.lock.Lock()
    defer c.lock.Unlock()
    if get, b = c.Cache.Get(key); b {
        return get
    }
    v := def()
    c.Cache.Set(key, v)
    return v
}

func NewCacheWrap[K comparable, V any](cache Cache[K, V]) *CacheWrap[K, V] {
    return &CacheWrap[K, V]{Cache: cache}
}
