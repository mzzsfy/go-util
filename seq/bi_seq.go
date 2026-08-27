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
    return func(f1 func(K, V)) { defer stopRecover(); f(f1) }
}

// FromSeq2 从函数生成Seq,是一个便捷方法
func FromSeq2[K, V any](f func(func(K, V) bool)) BiSeq[K, V] {
    return func(t func(K, V)) {
        defer stopRecover()
        f(func(k K, v V) bool {
            t(k, v)
            return true
        })
    }
}

func BiFromT[K, V any](k K, v V) BiSeq[K, V] {
    return func(t func(K, V)) {
        t(k, v)
    }
}

func BiFromSeq[K, V, T any](seq Seq[T], cast func(T) (K, V)) BiSeq[K, V] {
    return func(t func(K, V)) {
        defer stopRecover()
        seq(func(t1 T) { t(cast(t1)) })
    }
}
func BiFromIterator[K, V any](it BiIterator[K, V]) BiSeq[K, V] {
    return func(t func(K, V)) {
        defer stopRecover()
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
        defer stopRecover()
        for _, v := range ts {
            t(v.K, v.V)
        }
    }
}

// BiFromTupleRepeat 重复生成BiSeq,limit为0时无限重复
func BiFromTupleRepeat[K, V any](limit int, ts ...BiTuple[K, V]) BiSeq[K, V] {
    return func(t func(K, V)) {
        defer stopRecover()
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
        defer stopRecover()
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
        defer stopRecover()
        f(k, v)
        getChild(k, v).ForEach(func(k K, v V) { BiFromTreeT(k, v, getChild).ForEach(f) })
    }
}

// BiFromMap 从map生成BiSeq
func BiFromMap[K comparable, V any](m map[K]V) BiSeq[K, V] {
    return func(t func(K, V)) {
        defer stopRecover()
        for k, v := range m {
            t(k, v)
        }
    }
}

// BiFromMapRepeat 从map生成BiSeq
func BiFromMapRepeat[K comparable, V any](m map[K]V, limit ...int) BiSeq[K, V] {
    return func(t func(K, V)) {
        defer stopRecover()
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

func BiToMap[K comparable, V any](seq BiSeq[K, V]) map[K]V {
    m := make(map[K]V)
    seq(func(k K, v V) {
        m[k] = v
    })
    return m
}
