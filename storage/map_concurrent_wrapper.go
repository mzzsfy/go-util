package storage

import (
    "github.com/mzzsfy/go-util/unsafe"
    "sync"
    _ "unsafe"
)

type concurrentWrapper[K comparable, V any] struct {
    shards []Map[K, V]
    locks  []sync.RWMutex
    hash   unsafe.Hasher[K]
}

func (m *concurrentWrapper[K, V]) Get(key K) (V, bool) {
    slot := idxFn(m.hash.Hash(key))
    m.locks[slot].RLock()
    defer m.locks[slot].RUnlock()
    return m.shards[slot].Get(key)
}

func (m *concurrentWrapper[K, V]) GetSimple(key K) (value V) {
    value, _ = m.Get(key)
    return
}

func (m *concurrentWrapper[K, V]) Has(key K) bool {
    slot := idxFn(m.hash.Hash(key))
    m.locks[slot].RLock()
    defer m.locks[slot].RUnlock()
    return m.shards[slot].Has(key)
}

func (m *concurrentWrapper[K, V]) Delete(key K) {
    slot := idxFn(m.hash.Hash(key))
    m.locks[slot].Lock()
    defer m.locks[slot].Unlock()
    m.shards[slot].Delete(key)
}

func (m *concurrentWrapper[K, V]) Put(key K, value V) {
    slot := idxFn(m.hash.Hash(key))
    m.locks[slot].Lock()
    defer m.locks[slot].Unlock()
    m.shards[slot].Put(key, value)
}
func (m *concurrentWrapper[K, V]) Clean() {
    for i := 0; i < slotNumber; i++ {
        m.locks[i].Lock()
        if m.shards[i] != nil {
            m.shards[i].Clean()
        }
        m.locks[i].Unlock()
    }
}

func (m *concurrentWrapper[K, V]) Count() int {
    var count int
    for i := 0; i < slotNumber; i++ {
        m.locks[i].RLock()
        if m.shards[i] != nil {
            count += m.shards[i].Count()
        }
        m.locks[i].RUnlock()
    }
    return count
}

func (m *concurrentWrapper[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
    for i := 0; i < slotNumber; i++ {
        m.locks[i].Lock()
        if m.shards[i] != nil {
            if m.shards[i].Iter(cb) {
                m.locks[i].Unlock()
                return true
            }
        }
        m.locks[i].Unlock()
    }
    return false
}

func (m *concurrentWrapper[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
    return IterDelete[K, V](m, cb)
}

func MapTypeConcurrentWrapper[K comparable, V any](m MakeMap[K, V]) MakeMap[K, V] {
    return MapImpl[K, V](func() Map[K, V] {
        c := &concurrentWrapper[K, V]{
            hash:   NewDefaultHasher[K](),
            shards: make([]Map[K, V], slotNumber),
            locks:  make([]sync.RWMutex, slotNumber),
        }
        for i := 0; i < slotNumber; i++ {
            c.shards[i] = m.createMap()
        }
        return c
    })
}
