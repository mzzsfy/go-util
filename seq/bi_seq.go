package seq

type BiTuple[K, V any] struct {
    K K
    V V
}

// BiSeq 一种特殊的双元素集合,可以用于链式操作
type BiSeq[K, V any] func(k func(K, V))

//======生成========

// BiFrom 从BiSeq生成BiSeq
func BiFrom[K, V any](f BiSeq[K, V]) BiSeq[K, V] {
    return f
}

func BiFromT[K, V any](k K, v V) BiSeq[K, V] {
    return func(t func(K, V)) {
        t(k, v)
    }
}

func BiFromSeq[K, V, T any](seq Seq[T], cast func(T) (K, V)) BiSeq[K, V] {
    return func(t func(K, V)) { seq(func(t1 T) { t(cast(t1)) }) }
}
func BiFromIterator[K, V any](it BiIterator[K, V]) BiSeq[K, V] {
    return func(t func(K, V)) {
        for {
            k, v, ok := it()
            if !ok {
                break
            }
            t(k, v)
        }
    }
}

func BiFromTuple[K, V any](ts ...BiTuple[K, V]) BiSeq[K, V] {
    return func(t func(K, V)) {
        for _, v := range ts {
            t(v.K, v.V)
        }
    }
}

// BiFromTupleRepeat 重复生成BiSeq,limit为0时无限重复
func BiFromTupleRepeat[K, V any](limit int, ts ...BiTuple[K, V]) BiSeq[K, V] {
    return func(t func(K, V)) {
        if limit > 0 {
            for i := 0; i < limit; i++ {
                for _, v := range ts {
                    t(v.K, v.V)
                }
            }
        } else {
            for {
                for _, v := range ts {
                    t(v.K, v.V)
                }
            }
        }
    }
}
func BiFromTRepeat[K, V any](k K, v V, limit ...int) BiSeq[K, V] {
    return func(t func(K, V)) {
        if len(limit) > 0 && limit[0] > 0 {
            l := limit[0]
            for i := 0; i < l; i++ {
                t(k, v)
            }
        } else {
            for {
                t(k, v)
            }
        }
    }
}

// BiFromTreeT 树转BiSeq,其他的场景请使用seq的FromTree系列方法再转换为biSeq
func BiFromTreeT[K, V any](k K, v V, getChild func(K, V) BiSeq[K, V]) BiSeq[K, V] {
    return func(f func(K, V)) {
        f(k, v)
        getChild(k, v).ForEach(func(k K, v V) { BiFromTreeT(k, v, getChild).ForEach(f) })
    }
}

// BiFromMap 从map生成BiSeq
func BiFromMap[K comparable, V any](m map[K]V) BiSeq[K, V] {
    return func(t func(K, V)) {
        for k, v := range m {
            t(k, v)
        }
    }
}

// BiFromMapRepeat 从map生成BiSeq
func BiFromMapRepeat[K comparable, V any](m map[K]V, limit ...int) BiSeq[K, V] {
    return func(t func(K, V)) {
        if len(limit) > 0 && limit[0] > 0 {
            l := limit[0]
            for i := 0; i < l; i++ {
                for k, v := range m {
                    t(k, v)
                }
            }
        } else {
            for {
                for k, v := range m {
                    t(k, v)
                }
            }
        }
    }
}

//======静态转换方法,xxxK 表示操作左侧参数========

// BiCastAny 从BiSeq[any,any]强制转换为BiSeq[K,V]
func BiCastAny[K, V any](seq BiSeq[any, any]) BiSeq[K, V] {
    return func(c func(K, V)) { seq(func(k, v any) { c(k.(K), v.(V)) }) }
}

// BiCastAnyK 从BiSeq[any,any]强制转换为BiSeq[K,V]
func BiCastAnyK[K, V any](seq BiSeq[any, V]) BiSeq[K, V] {
    return func(c func(K, V)) { seq(func(k any, v V) { c(k.(K), v) }) }
}

// BiCastAnyV 从BiSeq[any,any]强制转换为BiSeq[K,V]
func BiCastAnyV[V, K any](seq BiSeq[K, any]) BiSeq[K, V] {
    return func(c func(K, V)) { seq(func(k K, v any) { c(k, v.(V)) }) }
}

// BiCastAnyT 从BiSeq[any,any]强制转换为BiSeq[K,V],简便写法
func BiCastAnyT[K, V any](seq BiSeq[any, any], _ K, _ V) BiSeq[K, V] {
    return func(c func(K, V)) { seq(func(k, v any) { c(k.(K), v.(V)) }) }
}

// BiCastAnyVT 从BiSeq[any,any]强制转换为BiSeq[K,V],简便写法
func BiCastAnyVT[V, K any](seq BiSeq[K, any], _ V) BiSeq[K, V] {
    return func(c func(K, V)) { seq(func(k K, v any) { c(k, v.(V)) }) }
}

// BiCastAnyKT 从BiSeq[any,any]强制转换为BiSeq[K,V],简便写法
func BiCastAnyKT[K, V any](seq BiSeq[any, V], _ K) BiSeq[K, V] {
    return func(c func(K, V)) { seq(func(k any, v V) { c(k.(K), v) }) }
}

// BiJoin 合并多个Seq
func BiJoin[K, V any](seqs ...BiSeq[K, V]) BiSeq[K, V] {
    return func(c func(K, V)) {
        for _, seq := range seqs {
            seq(func(k K, v V) { c(k, v) })
        }
    }
}

// BiJoinL 合并2个不同Seq,右边转换为左边的类型
func BiJoinL[K, V, K1, V1 any](seq1 BiSeq[K, V], seq2 BiSeq[K1, V1], cast func(K1, V1) (K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        seq1(func(k K, v V) { c(k, v) })
        seq2(func(k K1, v V1) { c(cast(k, v)) })
    }
}

// BiJoinBy 合并2个不同Seq,统一转换为新类型
func BiJoinBy[K1, V1, K2, V2, K, V any](seq1 BiSeq[K1, V1], cast1 func(K1, V1) (K, V), seq2 BiSeq[K2, V2], cast2 func(K2, V2) (K, V), ) BiSeq[K, V] {
    return func(c func(K, V)) {
        seq1(func(k K1, v V1) { c(cast1(k, v)) })
        seq2(func(k K2, v V2) { c(cast2(k, v)) })
    }
}

func BiToMap[K comparable, V any](seq BiSeq[K, V]) map[K]V {
    m := make(map[K]V)
    seq(func(k K, v V) {
        m[k] = v
    })
    return m
}
