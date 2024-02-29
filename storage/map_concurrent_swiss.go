package storage

import (
    "github.com/mzzsfy/go-util/unsafe"
    "runtime"
    "sync"
    _ "unsafe"
)

var slotNumber, modNumber = runtime.NumCPU(), runtime.NumCPU() - 1
var idxFn = func(hash uint64) int {
    //去除符号位,避免可能的负数
    return int(hash&0x7FFFFFFFFFFFFFFF) % modNumber
}

func init() {
    //32位系统,int为4字节,uint64强转int有损失
    if ^uint(0)>>63 == 0 {
        idxFn = func(key uint64) int {
            //高32位和低32位异或,减少hash冲突
            return int((key^(key>>32))&0x7FFFFFFF) % modNumber
        }
    }
}

type concurrentSwissMap[K comparable, V any] struct {
    shards []*swissMap[K, V]
    locks  []sync.RWMutex
    hash   unsafe.Hasher[K]
}

func (m *concurrentSwissMap[K, V]) Get(key K) (V, bool) {
    hash := m.hash.Hash(key)
    shard := idxFn(hash)
    m.locks[shard].RLock()
    defer m.locks[shard].RUnlock()
    return m.shards[shard].GetWithHash(key, hash)
}

func (m *concurrentSwissMap[K, V]) GetSimple(key K) (value V) {
    value, _ = m.Get(key)
    return
}

func (m *concurrentSwissMap[K, V]) Has(key K) bool {
    hash := m.hash.Hash(key)
    shard := idxFn(hash)
    m.locks[shard].RLock()
    defer m.locks[shard].RUnlock()
    return m.shards[shard].HasWithHash(key, hash)
}

func (m *concurrentSwissMap[K, V]) Delete(key K) {
    hash := m.hash.Hash(key)
    shard := idxFn(hash)
    m.locks[shard].Lock()
    defer m.locks[shard].Unlock()
    m.shards[shard].DeleteWithHash(key, hash)
}

func (m *concurrentSwissMap[K, V]) Put(key K, value V) {
    hash := m.hash.Hash(key)
    shard := idxFn(hash)
    m.locks[shard].Lock()
    defer m.locks[shard].Unlock()
    m.shards[shard].PutWithHash(key, value, hash)
}
func (m *concurrentSwissMap[K, V]) Clear() {
    for i := 0; i < slotNumber; i++ {
        m.locks[i].Lock()
        if m.shards[i] != nil {
            m.shards[i].Clear()
        }
        m.locks[i].Unlock()
    }
}

func (m *concurrentSwissMap[K, V]) Count() int {
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

func (m *concurrentSwissMap[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
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

func (m *concurrentSwissMap[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
    var delKeys []K
    r := m.Iter(func(k K, v V) bool {
        del, stop := cb(k, v)
        if del {
            delKeys = append(delKeys, k)
        }
        return stop
    })
    for _, k := range delKeys {
        m.Delete(k)
    }
    return r
}

func makeSwissConcurrentMap[K comparable, V any]() *concurrentSwissMap[K, V] {
    c := &concurrentSwissMap[K, V]{
        shards: make([]*swissMap[K, V], slotNumber),
        locks:  make([]sync.RWMutex, slotNumber),
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
