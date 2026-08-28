package storage

type arrayMap[K comparable, V any] struct {
	keys   []K
	values []V
}

func (m *arrayMap[K, V]) Has(key K) bool {
	for _, k := range m.keys {
		if k == key {
			return true
		}
	}
	return false
}

func (m *arrayMap[K, V]) Get(key K) (value V, exist bool) {
	for i, k := range m.keys {
		if k == key {
			return m.values[i], true
		}
	}
	return
}

func (m *arrayMap[K, V]) GetSimple(key K) (value V) {
	value, _ = m.Get(key)
	return
}

func (m *arrayMap[K, V]) Put(key K, value V) {
	for i, k := range m.keys {
		if k == key {
			m.values[i] = value
			return
		}
	}
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
}

// swapDelete 用末尾元素覆盖位置i, O(1)删除, 不保证顺序
func (m *arrayMap[K, V]) swapDelete(i int) {
	last := len(m.keys) - 1
	m.keys[i] = m.keys[last]
	m.values[i] = m.values[last]
	m.keys = m.keys[:last]
	m.values = m.values[:last]
}

func (m *arrayMap[K, V]) Delete(key K) {
	for i, k := range m.keys {
		if k == key {
			m.swapDelete(i)
			return
		}
	}
}

func (m *arrayMap[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
	for i, k := range m.keys {
		if cb(k, m.values[i]) {
			return true
		}
	}
	return false
}

// IterDelete 倒序遍历, 删除时用末尾元素覆盖(swap-delete), O(1)删除不影响未遍历元素
func (m *arrayMap[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
	for i := len(m.keys) - 1; i >= 0; i-- {
		del, stop := cb(m.keys[i], m.values[i])
		if del {
			m.swapDelete(i)
		}
		if stop {
			return true
		}
	}
	return false
}

func (m *arrayMap[K, V]) Clean() {
	m.keys = m.keys[:0]
	m.values = m.values[:0]
}

func (m *arrayMap[K, V]) Count() int {
	return len(m.keys)
}

// MapTypeArray array底层的map,适合小数据量(低于50个元素),空间利用率高,性能与key的数量成正比
func MapTypeArray[K comparable, V any](size int) MakeMap[K, V] {
	return MapImpl[K, V](func() Map[K, V] {
		return &arrayMap[K, V]{
			keys:   make([]K, 0, size),
			values: make([]V, 0, size),
		}
	})
}
