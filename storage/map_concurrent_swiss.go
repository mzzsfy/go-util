//go:build !go1.24

package storage

import (
    "github.com/mzzsfy/go-util/unsafe"
    _ "unsafe"
)

type concurrentSwissMap[K comparable, V any] struct {
    shards []*swissMap[K, V]
    locks  []paddedLock
    hash   unsafe.Hasher[K]
}

func (m *concurrentSwissMap[K, V]) Get(key K) (V, bool) {
    hash := m.hash.Hash(key)
    shard := slotIdx(hash)
    m.locks[shard].RLock()
    v, ok := m.shards[shard].GetWithHash(key, hash)
    m.locks[shard].RUnlock()
    return v, ok
}

func (m *concurrentSwissMap[K, V]) GetSimple(key K) (value V) {
    hash := m.hash.Hash(key)
    shard := slotIdx(hash)
    m.locks[shard].RLock()
    value, _ = m.shards[shard].GetWithHash(key, hash)
    m.locks[shard].RUnlock()
    return
}

func (m *concurrentSwissMap[K, V]) Has(key K) bool {
    hash := m.hash.Hash(key)
    shard := slotIdx(hash)
    m.locks[shard].RLock()
    ok := m.shards[shard].HasWithHash(key, hash)
    m.locks[shard].RUnlock()
    return ok
}

func (m *concurrentSwissMap[K, V]) Delete(key K) {
    hash := m.hash.Hash(key)
    shard := slotIdx(hash)
    m.locks[shard].Lock()
    m.shards[shard].DeleteWithHash(key, hash)
    m.locks[shard].Unlock()
}

func (m *concurrentSwissMap[K, V]) Put(key K, value V) {
    hash := m.hash.Hash(key)
    shard := slotIdx(hash)
    m.locks[shard].Lock()
    m.shards[shard].PutWithHash(key, value, hash)
    m.locks[shard].Unlock()
}

func (m *concurrentSwissMap[K, V]) Clean() {
    for i := 0; i < slotNumber; i++ {
        m.locks[i].Lock()
        m.shards[i].Clean()
        m.locks[i].Unlock()
    }
}

func (m *concurrentSwissMap[K, V]) Count() int {
    var count int
    for i := 0; i < slotNumber; i++ {
        m.locks[i].RLock()
        count += m.shards[i].Count()
        m.locks[i].RUnlock()
    }
    return count
}

func (m *concurrentSwissMap[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
    for i := 0; i < slotNumber; i++ {
        m.locks[i].RLock()
        if m.shards[i].Iter(cb) {
            m.locks[i].RUnlock()
            return true
        }
        m.locks[i].RUnlock()
    }
    return false
}

func (m *concurrentSwissMap[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
    for i := 0; i < slotNumber; i++ {
        m.locks[i].Lock()
        if m.shards[i].IterDelete(cb) {
            m.locks[i].Unlock()
            return true
        }
        m.locks[i].Unlock()
    }
    return false
}

func makeSwissConcurrentMap[K comparable, V any]() *concurrentSwissMap[K, V] {
    c := &concurrentSwissMap[K, V]{
        shards: make([]*swissMap[K, V], slotNumber),
        locks:  make([]paddedLock, slotNumber),
        hash:   NewDefaultHasher[K](),
    }
    for i := range c.shards {
        c.shards[i] = makeSwissMap[K, V](0)
    }
    return c
}

func MapTypeSwissConcurrent[K comparable, V any]() MakeMap[K, V] {
    return MapImpl[K, V](func() Map[K, V] { return makeSwissConcurrentMap[K, V]() })
}
