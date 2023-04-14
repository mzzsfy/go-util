package seq

import (
    "sort"
    "strings"
    "sync"
)

type biTuple[K, V any] struct {
    k K
    v V
}

// BiSeq 一种特殊的双元素集合,可以用于链式操作
type BiSeq[K, V any] func(k func(K, V))

//======生成========

// BiFrom 从BiSeq生成BiSeq
func BiFrom[K, V any](f BiSeq[K, V]) BiSeq[K, V] {
    return f
}

// BiFromMap 从map生成BiSeq
func BiFromMap[K comparable, V any](m map[K]V) BiSeq[K, V] {
    return func(t func(K, V)) {
        for k, v := range m {
            t(k, v)
        }
    }
}

// BiUnit 生成单元素的BiSeq
func BiUnit[K, V any](k K, v V) BiSeq[K, V] {
    return func(t func(K, V)) { t(k, v) }
}

// BiUnitRepeat 生成重复的单元素的BiSeq
func BiUnitRepeat[K, V any](k K, v V, limit ...int) BiSeq[K, V] {
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

// BiCastAny 从BiSeq[any,any]强制转换为BiSeq[K,V]
func BiCastAny[K, V any](seq BiSeq[any, any]) BiSeq[K, V] {
    return func(c func(K, V)) { seq(func(t, x any) { c(t.(K), x.(V)) }) }
}

// BiMap 从BiSeq[K,V]自定义转换为BiSeq[RK,RV]
func BiMap[K, V, RK, RV any](seq BiSeq[K, V], cast func(K, V) (RK, RV)) BiSeq[RK, RV] {
    return func(c func(RK, RV)) { seq(func(k K, v V) { c(cast(k, v)) }) }
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

// BiJoinF 合并2个不同Seq,统一转换为新类型
func BiJoinF[K1, V1, K2, V2, K, V any](seq1 BiSeq[K1, V1], cast1 func(K1, V1) (K, V), seq2 BiSeq[K2, V2], cast2 func(K2, V2) (K, V), ) BiSeq[K, V] {
    return func(c func(K, V)) {
        seq1(func(k K1, v V1) { c(cast1(k, v)) })
        seq2(func(k K2, v V2) { c(cast2(k, v)) })
    }
}

//======转换========

// Map 每个元素自定义转换为any,用于连续转换操作,使用 BiCastAny 进行恢复泛型
func (t BiSeq[K, V]) Map(f func(K, V) (any, any)) BiSeq[any, any] {
    return func(c func(any, any)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// KMap 每个元素自定义转换K为any,用于连续转换操作,使用 BiCastAny 进行恢复泛型
func (t BiSeq[K, V]) KMap(f func(K, V) any) BiSeq[any, V] {
    return func(c func(any, V)) { t(func(k K, v V) { c(f(k, v), v) }) }
}

// VMap 每个元素自定义转换V为any,用于连续转换操作,使用 BiCastAny 进行恢复泛型
func (t BiSeq[K, V]) VMap(f func(K, V) any) BiSeq[K, any] {
    return func(c func(K, any)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// KSeq 转换为只保留K的Seq
func (t BiSeq[K, V]) KSeq() Seq[K] {
    return func(c func(K)) { t(func(k K, v V) { c(k) }) }
}

// VSeq 转换为只保留V的Seq
func (t BiSeq[K, V]) VSeq() Seq[V] {
    return func(c func(V)) { t(func(k K, v V) { c(v) }) }
}

// KSeqF 转换为只保留K的Seq,并自定义转换
func (t BiSeq[K, V]) KSeqF(f func(K, V) K) Seq[K] {
    return func(c func(K)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// VSeqF 转换为只保留V的Seq,并自定义转换
func (t BiSeq[K, V]) VSeqF(f func(K, V) V) Seq[V] {
    return func(c func(V)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// SeqF 转换为Seq[any],自定义转换
func (t BiSeq[K, V]) SeqF(f func(K, V) V) Seq[any] {
    return func(c func(any)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// FlatMap 每个元素转换为BiSeq[any,any],并扁平化
func (t BiSeq[K, V]) FlatMap(f func(K, V) BiSeq[any, any]) BiSeq[any, any] {
    return func(c func(any, any)) {
        t(func(k K, v V) {
            s := f(k, v)
            s.DoEach(c)
        })
    }
}

// StringMap 转换为Seq[string]
func (t BiSeq[K, V]) StringMap(f func(K, V) string) Seq[string] {
    return func(c func(string)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// Join 合并多个Seq
func (t BiSeq[K, V]) Join(seqs ...BiSeq[K, V]) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) { c(k, v) })
        for _, seq := range seqs {
            seq(func(k K, v V) { c(k, v) })
        }
    }
}

// JoinF 合并Seq
func (t BiSeq[K, V]) JoinF(seq BiSeq[any, any], cast func(any, any) (K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) { c(k, v) })
        seq(func(k any, v any) { c(cast(k, v)) })
    }
}

//======HOOK========

// OnEach 每个元素执行f
func (t BiSeq[K, V]) OnEach(f func(K, V)) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) {
            f(k, v)
            c(k, v)
        })
    }
}

// Parallel 对后续操作启用并行执行
func (t BiSeq[K, V]) Parallel() BiSeq[K, V] {
    return func(c func(K, V)) {
        t.Map(func(k K, v V) (any, any) {
            lock := sync.Mutex{}
            lock.Lock()
            go func() {
                defer lock.Unlock()
                c(k, v)
            }()
            return &lock, nil
        }).Cache()(func(t, _ any) {
            lock := t.(sync.Locker)
            lock.Lock()
        })
    }
}

//======消费========

// DoEach 每个元素执行f
func (t BiSeq[K, V]) DoEach(f func(K, V)) {
    t(f)
}

// AsyncEach 每个元素执行f,并行执行
func (t BiSeq[K, V]) AsyncEach(f func(K, V)) { t.Parallel().DoEach(f) }

// First 获取第一个元素,无则返回nil
func (t BiSeq[K, V]) First() (*K, *V) {
    var rk *K
    var rv *V
    t.Take(1)(func(k K, v V) {
        rk = &k
        rv = &v
    })
    return rk, rv
}

// FirstOr 获取第一个元素,无则返回默认值
func (t BiSeq[K, V]) FirstOr(k K, v V) (K, V) {
    var rk *K
    var rv *V
    exist := false
    t.Take(1)(func(k K, v V) {
        rk = &k
        rv = &v
        exist = true
    })
    if exist {
        return *rk, *rv
    }
    return k, v
}

// FirstOrF 获取第一个元素,无则返回f的值
func (t BiSeq[K, V]) FirstOrF(f func() (K, V)) (K, V) {
    var rk *K
    var rv *V
    exist := false
    t.Take(1)(func(k K, v V) {
        rk = &k
        rv = &v
        exist = true
    })
    if exist {
        return *rk, *rv
    }
    return f()
}

// AnyMatch 任意匹配
func (t BiSeq[K, V]) AnyMatch(f func(K, V) bool) bool {
    r := false
    t.Filter(f).Take(1)(func(K, V) { r = true })
    return r
}

// AllMatch 全部匹配
func (t BiSeq[K, V]) AllMatch(f func(K, V) bool) bool {
    r := true
    t.Filter(func(k K, v V) bool { return !f(k, v) }).Take(1)(func(K, V) { r = false })
    return r
}

// Keys 获取所有K
func (t BiSeq[K, V]) Keys() []K {
    var r []K
    t(func(t K, _ V) { r = append(r, t) })
    return r
}

// Values 获取所有V
func (t BiSeq[K, V]) Values() []V {
    var r []V
    t(func(_ K, t V) { r = append(r, t) })
    return r
}

// Cache 缓存Seq,使该Seq可以多次消费,并保证前面内容不会重复执行
func (t BiSeq[K, V]) Cache() BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) { r = append(r, biTuple[K, V]{k, v}) })
    return func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    }
}

// Complete 消费所有元素
func (t BiSeq[K, V]) Complete() { t(func(_ K, _ V) {}) }

// JoinString 拼接为字符串
func (t BiSeq[K, V]) JoinString(f func(K, V) string, delimiter ...string) string {
    sb := strings.Builder{}
    d := ""
    if len(delimiter) > 0 {
        d = delimiter[0]
    }
    t.StringMap(f).DoEach(func(s string) {
        if d != "" && sb.Len() > 0 {
            sb.WriteString(d)
        }
        sb.WriteString(s)
    })
    return sb.String()
}

// Reduce 求值
func (t BiSeq[K, V]) Reduce(f func(K, V, any) any, init any) any {
    t.DoEach(func(k K, v V) { init = f(k, v, init) })
    return init
}

// Sort 排序
func (t BiSeq[K, V]) Sort(less func(K, V, K, V) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) { r = append(r, biTuple[K, V]{k, v}) })
    sort.Slice(r, func(i, j int) bool { return less(r[i].k, r[i].v, r[j].k, r[j].v) })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// SortK 根据K排序
func (t BiSeq[K, V]) SortK(less func(K, K) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) { r = append(r, biTuple[K, V]{k, v}) })
    sort.Slice(r, func(i, j int) bool { return less(r[i].k, r[j].k) })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// SortV 根据V排序
func (t BiSeq[K, V]) SortV(less func(V, V) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) { r = append(r, biTuple[K, V]{k, v}) })
    sort.Slice(r, func(i, j int) bool { return less(r[i].v, r[j].v) })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// Distinct 去重
func (t BiSeq[K, V]) Distinct(eq func(K, V, K, V) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) {
        for _, x := range r {
            if eq(k, v, x.k, x.v) {
                return
            }
        }
        r = append(r, biTuple[K, V]{k, v})
    })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// DistinctK 使用K去重
func (t BiSeq[K, V]) DistinctK(eq func(K, K) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) {
        for _, x := range r {
            if eq(k, x.k) {
                return
            }
        }
        r = append(r, biTuple[K, V]{k, v})
    })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

// DistinctV 使用V去重
func (t BiSeq[K, V]) DistinctV(eq func(V, V) bool) BiSeq[K, V] {
    var r []biTuple[K, V]
    t(func(k K, v V) {
        for _, x := range r {
            if eq(v, x.v) {
                return
            }
        }
        r = append(r, biTuple[K, V]{k, v})
    })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.k, v.v)
        }
    })
}

//======控制========

// consumeTillStop 消费直到遇到stop
func (t BiSeq[K, V]) consumeTillStop(f func(K, V)) {
    defer func() {
        a := recover()
        if a != nil && a != &stop {
            panic(a)
        }
    }()
    t(f)
}

// Filter 过滤元素,只保留满足条件的元素,即f() == true保留
func (t BiSeq[K, V]) Filter(f func(K, V) bool) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) {
            if f(k, v) {
                c(k, v)
            }
        })
    }
}

// Take 保留前n个元素
func (t BiSeq[K, V]) Take(n int) BiSeq[K, V] {
    return func(c func(K, V)) {
        t.consumeTillStop(func(k K, v V) {
            n--
            if n >= 0 {
                c(k, v)
            } else {
                panic(&stop)
            }
        })
    }
}

// Drop 跳过前n个元素
func (t BiSeq[K, V]) Drop(n int) BiSeq[K, V] {
    return func(c func(K, V)) {
        i := n
        t(func(k K, v V) {
            if i <= 0 {
                c(k, v)
            } else {
                i--
            }
        })
    }
}
