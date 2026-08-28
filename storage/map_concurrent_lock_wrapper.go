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
	v, ok := m.m.Get(key)
	m.lock.RUnlock()
	return v, ok
}

func (m *rwWrapper[K, V]) GetSimple(key K) (value V) {
	m.lock.RLock()
	value = m.m.GetSimple(key)
	m.lock.RUnlock()
	return
}

func (m *rwWrapper[K, V]) Has(key K) bool {
	m.lock.RLock()
	ok := m.m.Has(key)
	m.lock.RUnlock()
	return ok
}

func (m *rwWrapper[K, V]) Delete(key K) {
	m.lock.Lock()
	m.m.Delete(key)
	m.lock.Unlock()
}

func (m *rwWrapper[K, V]) Put(key K, value V) {
	m.lock.Lock()
	m.m.Put(key, value)
	m.lock.Unlock()
}

func (m *rwWrapper[K, V]) Clean() {
	m.lock.Lock()
	m.m.Clean()
	m.lock.Unlock()
}

func (m *rwWrapper[K, V]) Count() int {
	m.lock.RLock()
	n := m.m.Count()
	m.lock.RUnlock()
	return n
}

func (m *rwWrapper[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
	m.lock.RLock()
	r := m.m.Iter(cb)
	m.lock.RUnlock()
	return r
}

func (m *rwWrapper[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
	m.lock.Lock()
	if idm, ok := m.m.(IterDeleteMap[K, V]); ok {
		r := idm.IterDelete(cb)
		m.lock.Unlock()
		return r
	}
	r := IterDelete[K, V](m.m, cb)
	m.lock.Unlock()
	return r
}

// MapTypeConcurrentLockWrapper 轻量级的并发包装
func MapTypeConcurrentLockWrapper[K comparable, V any](m MakeMap[K, V]) MakeMap[K, V] {
	return MapImpl[K, V](func() Map[K, V] {
		return &rwWrapper[K, V]{m: m.createMap()}
	})
}
