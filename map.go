package util

func MapSet[M ~map[K]V, K comparable, V any](m *M, k K, v V) *M {
    if IsZero(m) {
        *m = make(M)
    }
    m2 := *m
    m2[k] = v
    return m
}

func Map1[K comparable, V any](k K, v V) map[K]V {
    m := make(map[K]V)
    m[k] = v
    return m
}

func Map2[K comparable, V any](k K, v V, k1 K, v1 V) map[K]V {
    m := make(map[K]V)
    m[k] = v
    m[k1] = v1
    return m
}

func Map3[K comparable, V any](k K, v V, k1 K, v1 V, k2 K, v2 V) map[K]V {
    m := make(map[K]V)
    m[k] = v
    m[k1] = v1
    m[k2] = v2
    return m
}

// CopyMap 复制kv到另一个map,返回为target,可nil
func CopyMap[M ~map[K]V, K comparable, V any](source, target M) M {
    if target == nil {
        target = make(M, len(source))
    }
    for k, v := range source {
        target[k] = v
    }
    return target
}

//todo: MapBuilder
