package storage

import (
    "sync"
    _ "unsafe"
)

type rwWrapper[K comparable, V any] struct {
    lock sync.RWMutex
    m    Map[K, V]
}

func (m *rwWrapper[K, V]) Get(key K) (V, bool) {
    m.lock.RLock()
    defer m.lock.RUnlock()
    return m.m.Get(key)
}

func (m *rwWrapper[K, V]) GetSimple(key K) (value V) {
    value, _ = m.Get(key)
    return
}

func (m *rwWrapper[K, V]) Has(key K) bool {
    m.lock.RLock()
    defer m.lock.RUnlock()
    return m.m.Has(key)
}

func (m *rwWrapper[K, V]) Delete(key K) {
    m.lock.Lock()
    defer m.lock.Unlock()
    m.m.Delete(key)
}

func (m *rwWrapper[K, V]) Put(key K, value V) {
    m.lock.Lock()
    defer m.lock.Unlock()
    m.m.Put(key, value)
}
func (m *rwWrapper[K, V]) Clean() {
    m.lock.Lock()
    defer m.lock.Unlock()
    m.m.Clean()
}

func (m *rwWrapper[K, V]) Count() int {
    m.lock.RLock()
    defer m.lock.RUnlock()
    return m.m.Count()
}

func (m *rwWrapper[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
    m.lock.RLock()
    defer m.lock.RUnlock()
    return m.m.Iter(cb)
}

func (m *rwWrapper[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
    return IterDelete[K, V](m, cb)
}

// MapTypeConcurrentLockWrapper 轻量级的并发包装
func MapTypeConcurrentLockWrapper[K comparable, V any](m MakeMap[K, V]) MakeMap[K, V] {
    return MapImpl[K, V](func() Map[K, V] {
        return &rwWrapper[K, V]{m: m.createMap()}
    })
}
