package storage

type Map[K comparable, V any] interface {
    Has(key K) bool
    Get(key K) (value V, ok bool)
    Put(key K, value V)
    Delete(key K) (ok bool)
    Iter(cb func(k K, v V) (stop bool)) bool
    IterDelete(cb func(k K, v V) (del, stop bool)) bool
    Clear()
    Count() int
}
