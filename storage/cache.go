package storage

import "time"

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

//todo
