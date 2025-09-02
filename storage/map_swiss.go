//go:build go1.24

package storage

func MapTypeSwiss[K comparable, V any](size uint32) MakeMap[K, V] {
    return MapTypeConcurrentWrapper(MapTypeGo(size))
}
