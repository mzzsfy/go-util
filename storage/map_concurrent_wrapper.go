package storage

import (
	"github.com/mzzsfy/go-util/unsafe"
	_ "unsafe"
)

type concurrentWrapper[K comparable, V any] struct {
	shards []Map[K, V]
	locks  []paddedLock
	hash   unsafe.Hasher[K]
}

func (m *concurrentWrapper[K, V]) Get(key K) (V, bool) {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].RLock()
	v, ok := m.shards[slot].Get(key)
	m.locks[slot].RUnlock()
	return v, ok
}

func (m *concurrentWrapper[K, V]) GetSimple(key K) (value V) {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].RLock()
	value = m.shards[slot].GetSimple(key)
	m.locks[slot].RUnlock()
	return
}

func (m *concurrentWrapper[K, V]) Has(key K) bool {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].RLock()
	ok := m.shards[slot].Has(key)
	m.locks[slot].RUnlock()
	return ok
}

func (m *concurrentWrapper[K, V]) Delete(key K) {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].Lock()
	m.shards[slot].Delete(key)
	m.locks[slot].Unlock()
}

func (m *concurrentWrapper[K, V]) Put(key K, value V) {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].Lock()
	m.shards[slot].Put(key, value)
	m.locks[slot].Unlock()
}

func (m *concurrentWrapper[K, V]) Clean() {
	for i := 0; i < slotNumber; i++ {
		m.locks[i].Lock()
		m.shards[i].Clean()
		m.locks[i].Unlock()
	}
}

func (m *concurrentWrapper[K, V]) Count() int {
	var count int
	for i := 0; i < slotNumber; i++ {
		m.locks[i].RLock()
		count += m.shards[i].Count()
		m.locks[i].RUnlock()
	}
	return count
}

func (m *concurrentWrapper[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
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

func (m *concurrentWrapper[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
	// 所有分片类型相同, 仅检查一次
	_, canIterDelete := m.shards[0].(IterDeleteMap[K, V])
	for i := 0; i < slotNumber; i++ {
		m.locks[i].Lock()
		var stopped bool
		if canIterDelete {
			stopped = m.shards[i].(IterDeleteMap[K, V]).IterDelete(cb)
		} else {
			stopped = IterDelete[K, V](m.shards[i], cb)
		}
		m.locks[i].Unlock()
		if stopped {
			return true
		}
	}
	return false
}

func MapTypeConcurrentWrapper[K comparable, V any](m MakeMap[K, V]) MakeMap[K, V] {
	return MapImpl[K, V](func() Map[K, V] {
		c := &concurrentWrapper[K, V]{
			hash:   NewDefaultHasher[K](),
			shards: make([]Map[K, V], slotNumber),
			locks:  make([]paddedLock, slotNumber),
		}
		for i := 0; i < slotNumber; i++ {
			c.shards[i] = m.createMap()
		}
		return c
	})
}

// concurrentGoMap 直接操作 []map[K]V, 消除 Map 接口派发开销
type concurrentGoMap[K comparable, V any] struct {
	shards []map[K]V
	locks  []paddedLock
	hash   unsafe.Hasher[K]
}

func (m *concurrentGoMap[K, V]) Get(key K) (V, bool) {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].RLock()
	v, ok := m.shards[slot][key]
	m.locks[slot].RUnlock()
	return v, ok
}

func (m *concurrentGoMap[K, V]) GetSimple(key K) (value V) {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].RLock()
	value = m.shards[slot][key]
	m.locks[slot].RUnlock()
	return
}

func (m *concurrentGoMap[K, V]) Has(key K) bool {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].RLock()
	_, ok := m.shards[slot][key]
	m.locks[slot].RUnlock()
	return ok
}

func (m *concurrentGoMap[K, V]) Delete(key K) {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].Lock()
	delete(m.shards[slot], key)
	m.locks[slot].Unlock()
}

func (m *concurrentGoMap[K, V]) Put(key K, value V) {
	slot := slotIdx(m.hash.Hash(key))
	m.locks[slot].Lock()
	m.shards[slot][key] = value
	m.locks[slot].Unlock()
}

func (m *concurrentGoMap[K, V]) Clean() {
	for i := 0; i < slotNumber; i++ {
		m.locks[i].Lock()
		m.shards[i] = make(map[K]V, 8)
		m.locks[i].Unlock()
	}
}

func (m *concurrentGoMap[K, V]) Count() int {
	var count int
	for i := 0; i < slotNumber; i++ {
		m.locks[i].RLock()
		count += len(m.shards[i])
		m.locks[i].RUnlock()
	}
	return count
}

func (m *concurrentGoMap[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
	for i := 0; i < slotNumber; i++ {
		m.locks[i].RLock()
		for k, v := range m.shards[i] {
			if cb(k, v) {
				m.locks[i].RUnlock()
				return true
			}
		}
		m.locks[i].RUnlock()
	}
	return false
}

func (m *concurrentGoMap[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
	for i := 0; i < slotNumber; i++ {
		m.locks[i].Lock()
		for k, v := range m.shards[i] {
			del, stop := cb(k, v)
			if del {
				delete(m.shards[i], k)
			}
			if stop {
				m.locks[i].Unlock()
				return true
			}
		}
		m.locks[i].Unlock()
	}
	return false
}

// MapTypeConcurrentGo 返回基于 Go 原生 map 的分片并发 map, 消除接口派发开销
func MapTypeConcurrentGo[K comparable, V any]() MakeMap[K, V] {
	return MapImpl[K, V](func() Map[K, V] {
		c := &concurrentGoMap[K, V]{
			hash:   NewDefaultHasher[K](),
			shards: make([]map[K]V, slotNumber),
			locks:  make([]paddedLock, slotNumber),
		}
		for i := range c.shards {
			c.shards[i] = make(map[K]V, 8)
		}
		return c
	})
}

// MapTypeConcurrentGoWithCap 指定预期总容量的并发 map, 预分配分片避免增长
func MapTypeConcurrentGoWithCap[K comparable, V any](cap int) MakeMap[K, V] {
	perShard := cap / slotNumber
	if perShard < 8 {
		perShard = 8
	}
	return MapImpl[K, V](func() Map[K, V] {
		c := &concurrentGoMap[K, V]{
			hash:   NewDefaultHasher[K](),
			shards: make([]map[K]V, slotNumber),
			locks:  make([]paddedLock, slotNumber),
		}
		for i := range c.shards {
			c.shards[i] = make(map[K]V, perShard)
		}
		return c
	})
}
