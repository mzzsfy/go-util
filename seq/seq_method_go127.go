//go:build go1.27

package seq

// 本文件在 go1.27 轨道提供泛型方法形态的强类型 API,替代 go1.27 以下版本的 any 妥协函数(见 *_pre127.go)

//======转换相关========

// MapParallel 每个元素转换为 R
// order.1 顺序保证方式,规则如下:
// 0:不保证任务启动顺序,不保证消费顺序,会消费竞争
// 1:尽量保持顺序,优先保证并发数,异步任务完成时,会直接消费,会消费竞争,可以使用 Sync() 保证消费不竞争
// 2:异步任务与消费端解偶,在保证顺序的前提下,优先保证并发数,不会消费竞争
// 3:保持异步与消费同步,以消费为准,不消费完成不会开始下一个异步任务,不会消费竞争
//
// order.2 最大并发数,根据第一个参数决定逻辑
func (t Seq[T]) MapParallel[R any](syncFn func(T) R, order ...int) Seq[R] {
    o := 0
    sl := 0
    if len(order) > 0 {
        o = order[0]
    }
    if len(order) > 1 {
        sl = order[1]
    }
    if o > 0 {
        return mapParallelCore(t, syncFn, o, sl)
    } else {
        return t.Parallel(sl).Map(syncFn)
    }
}

func (t Seq[T]) MapParallelCustomize[R any](asyncFn func(T, func(R))) Seq[R] {
    return mapParallelCustomizeCore(t, asyncFn)
}

// Map 每个元素自定义转换,E 通常可由 cast 函数推断
func (t Seq[T]) Map[E any](cast func(T) E) Seq[E] {
    return func(c func(E)) { t(func(t T) { c(cast(t)) }) }
}

// MapFlat 每个元素转换为Seq,并扁平化
func (t Seq[T]) MapFlat[E any](f func(T) Seq[E]) Seq[E] {
    return func(c func(E)) { t(func(t T) { f(t)(c) }) }
}

// MapFlatBi 每个元素转换为BiSeq,并扁平化
func (t Seq[T]) MapFlatBi[A, B any](f func(T) BiSeq[A, B]) BiSeq[A, B] {
    return func(c func(A, B)) { t(func(t T) { f(t)(c) }) }
}

// MapFlatK 每个元素展开为多个值并与原元素配对,展开值占K位
func (t Seq[T]) MapFlatK[E any](f func(T) Seq[E]) BiSeq[E, T] {
    return func(c func(E, T)) {
        t(func(t T) {
            f(t).ForEach(func(e E) {
                c(e, t)
            })
        })
    }
}

// MapFlatV 每个元素展开为多个值并与原元素配对,展开值占V位
func (t Seq[T]) MapFlatV[E any](f func(T) Seq[E]) BiSeq[T, E] {
    return func(c func(T, E)) {
        t(func(t T) {
            f(t).ForEach(func(e E) {
                c(t, e)
            })
        })
    }
}

// JoinL 合并2个不同Seq,右边转换为左边的类型
func (t Seq[T]) JoinL[E any](seq2 Seq[E], cast func(E) T) Seq[T] {
    return func(c func(T)) {
        defer stopRecover()
        t(func(t T) { c(t) })
        seq2(func(t E) { c(cast(t)) })
    }
}

// JoinBy 合并2个不同Seq,统一转换为新类型
func (t Seq[T]) JoinBy[E, R any](seq2 Seq[E], cast1 func(T) R, cast2 func(E) R) Seq[R] {
    return func(c func(R)) {
        defer stopRecover()
        t(func(t T) { c(cast1(t)) })
        seq2(func(t E) { c(cast2(t)) })
    }
}

// Cast 强制断言每个元素为N类型,用于恢复经过any化的强类型链
func (t Seq[T]) Cast[N any]() Seq[N] {
    return func(c func(N)) { t(func(t T) { c(any(t).(N)) }) }
}

//======消费相关========

// GroupBy 元素分组,每个组保留所有元素,K 为分组键的类型
func (t Seq[T]) GroupBy[K comparable](f func(T) K) map[K][]T {
    r := make(map[K][]T)
    t(func(t T) {
        k := f(t)
        r[k] = append(r[k], t)
    })
    return r
}

// GroupByFirst 元素分组,每个组只保留第一个元素
func (t Seq[T]) GroupByFirst[K comparable](f func(T) K) map[K]T {
    r := make(map[K]T)
    t(func(t T) {
        k := f(t)
        if _, ok := r[k]; !ok {
            r[k] = t
        }
    })
    return r
}

// GroupByLast 元素分组,每个组只保留最后一个元素
func (t Seq[T]) GroupByLast[K comparable](f func(T) K) map[K]T {
    r := make(map[K]T)
    t(func(t T) {
        k := f(t)
        r[k] = t
    })
    return r
}

// Reduce 自定义聚合,R 由 init 的类型确定
func (t Seq[T]) Reduce[R any](f func(T, R) R, init R) R {
    r := init
    t(func(t T) { r = f(t, r) })
    return r
}

//======转换为 BiSeq========

// MergeBi 与一个Iterator合并,K 由迭代器元素类型推断
func (t Seq[T]) MergeBi[K any](iterator Iterator[K]) BiSeq[K, T] {
    return BiFrom(func(f1 func(K, T)) {
        defer stopRecover()
        t(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(x, t)
        })
    })
}

// MergeBiR 与一个Iterator合并,V 由迭代器元素类型推断
func (t Seq[T]) MergeBiR[V any](iterator Iterator[V]) BiSeq[T, V] {
    return BiFrom(func(f1 func(T, V)) {
        defer stopRecover()
        t(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(t, x)
        })
    })
}

// MapBi 以 f 的返回值作为K
func (t Seq[T]) MapBi[K any](f func(T) K) BiSeq[K, T] {
    return BiFrom(func(c func(K, T)) { t(func(t T) { c(f(t), t) }) })
}

// MapBiR 以 f 的返回值作为V
func (t Seq[T]) MapBiR[V any](f func(T) V) BiSeq[T, V] {
    return BiFrom(func(c func(T, V)) { t(func(t T) { c(t, f(t)) }) })
}

// Enumerate 为每个元素生成一个序列号,参数Range规则参考 FromIntSeq
func (t Seq[T]) Enumerate(Range ...int) BiSeq[int, T] {
    return BiFrom(func(c func(int, T)) {
        defer stopRecover()
        r := makeRange(Range...)
        t(func(t T) {
            c(r(), t)
        })
    })
}
