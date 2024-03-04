package storage

type Map[K comparable, V any] interface {
    Has(key K) bool
    Get(key K) (value V, ok bool)
    GetSimple(key K) V
    Put(key K, value V)
    Delete(key K)
    // Iter 尽量不要在回调中删除元素,部分map可能不支持
    Iter(cb func(k K, v V) (stop bool)) bool
    Clean()
    Count() int
}

type IterDeleteMap[K comparable, V any] interface {
    Map[K, V]
    IterDelete(cb func(k K, v V) (del, stop bool)) bool
}

type MakeMap[K comparable, V any] interface {
    createMap() Map[K, V]
}

type MapImpl[K comparable, V any] func() Map[K, V]

func (m MapImpl[K, V]) createMap() Map[K, V] {
    return m()
}

// NewMap returns a new map. opt like MapTypeXXXX
func NewMap[K comparable, V any](opt ...MakeMap[K, V]) (r Map[K, V]) {
    if len(opt) > 0 {
        return opt[0].createMap()
    }
    return MapTypeSwiss[K, V](16).createMap()
}
