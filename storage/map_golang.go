package storage

type goMap[K comparable, V any] struct {
    m map[K]V
}

func (g *goMap[K, V]) Has(key K) bool {
    _, ok := g.m[key]
    return ok
}

func (g *goMap[K, V]) Get(key K) (value V, ok bool) {
    value, ok = g.m[key]
    return
}

func (g *goMap[K, V]) GetSimple(key K) (value V) {
    value, _ = g.m[key]
    return
}

func (g *goMap[K, V]) Put(key K, value V) {
    g.m[key] = value
}

func (g *goMap[K, V]) Delete(key K) {
    delete(g.m, key)
}

func (g *goMap[K, V]) Iter(cb func(k K, v V) (stop bool)) bool {
    for k, v := range g.m {
        if stop := cb(k, v); stop {
            return true
        }
    }
    return false
}

func (g *goMap[K, V]) IterDelete(cb func(k K, v V) (del bool, stop bool)) bool {
    for k, v := range g.m {
        if del, stop := cb(k, v); del {
            delete(g.m, k)
            if stop {
                return true
            }
        }
    }
    return false
}

func (g *goMap[K, V]) Clean() {
    g.m = make(map[K]V)
}

func (g *goMap[K, V]) Count() int {
    return len(g.m)
}

func MapTypeGo[K comparable, V any](cap int) MakeMap[K, V] {
    return MapImpl[K, V](func() Map[K, V] {
        return &goMap[K, V]{m: make(map[K]V, cap)}
    })
}
