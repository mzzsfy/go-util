package seq

import (
    "math"
    "math/rand"
)

//参考来源: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw

// Seq 一种特殊的集合,可以用于链式操作
type Seq[T any] func(t func(T))

//======生成,产生新的seq========

// From 从函数生成Seq,是一个便捷方法
func From[T any](f Seq[T]) Seq[T] {
    return f
}

// FromSlice 从数组生成Seq
func FromSlice[T any](arr []T) Seq[T] {
    return func(t func(T)) {
        for _, v := range arr {
            t(v)
        }
    }
}

// FromSliceRepeat 从数组生成Seq,可以指定重复次数
func FromSliceRepeat[T any](arr []T, limit ...int) Seq[T] {
    return func(t func(T)) {
        if len(limit) > 0 && limit[0] > 0 {
            l := limit[0]
            for i := 0; i < l; i++ {
                for _, e := range arr {
                    t(e)
                }
            }
        } else {
            for {
                for _, e := range arr {
                    t(e)
                }
            }
        }

    }
}

// FromChan 从Chan生成Seq
func FromChan[T any](c <-chan T) Seq[T] {
    return func(t func(T)) {
        for v := range c {
            t(v)
        }
    }
}

// FromIterator 从Iterator生成Seq
func FromIterator[T any](it Iterator[T]) Seq[T] {
    return func(t func(T)) {
        for {
            item, ok := it()
            if !ok {
                break
            }
            t(item)
        }
    }
}

// FromT 从元素生成Seq
func FromT[T any](ts ...T) Seq[T] {
    return FromSlice(ts)
}

func FromTRepeat[T any](ts ...T) Seq[T] {
    return FromSliceRepeat(ts)
}
func FromTRepeatN[T any](limit int, ts ...T) Seq[T] {
    return FromSliceRepeat(ts, limit)
}

// FromRandIntSeq 生成随机整数序列,可以自定义随机数范围
// 如果不指定范围,则生成的随机数为int类型的最大值[0,MaxInt)
// 参数1: 生成数量
// 参数2: 随机数范围:[0,n)
func FromRandIntSeq(i ...int) Seq[int] {
    l := math.MaxInt
    r := 0
    if len(i) > 0 {
        l = i[0]
    }
    if len(i) > 1 {
        r = i[1]
    }
    return func(t func(int)) {
        if r > 0 {
            for st := 0; st <= l; st++ {
                t(rand.Intn(r))
            }
        } else {
            for st := 0; st <= l; st++ {
                t(rand.Int())
            }
        }
    }
}

// FromIntSeq 生成整数序列,可以自定义起始值,结束值,步长
// 参数1,起始值,默认为0
// 参数2,结束值,默认为int类型的最大值
// 参数3,步长,默认为1
func FromIntSeq(Range ...int) Seq[int] {
    return func(f func(int)) {
        r := makeRange(Range...)
        defer func() {
            a := recover()
            if a != nil && a != &Stop {
                panic(a)
            }
        }()
        for {
            f(r())
        }
    }
}

func FromTreeT[T any](t T, getChild func(T) []T) Seq[T] {
    return func(f func(T)) {
        f(t)
        for _, c := range getChild(t) {
            FromTreeT(c, getChild).ForEach(f)
        }
    }
}
func FromTreeTV[T, V any](p T, getChild func(T) []T, getValue func(T) V) Seq[V] {
    return func(f func(V)) {
        f(getValue(p))
        for _, c := range getChild(p) {
            FromTreeTV(c, getChild, getValue)(f)
        }
    }
}

func FromTreeAny(o any, getChild func(any) []any) Seq[any] {
    return func(f func(any)) {
        f(o)
        for _, c := range getChild(o) {
            FromTreeAny(c, getChild)(f)
        }
    }
}

func FromTreeAnyTV[V any](o any, getChild func(any) []any, getValue func(any) V) Seq[V] {
    return func(f func(V)) {
        f(getValue(o))
        for _, c := range getChild(o) {
            FromTreeAnyTV(c, getChild, getValue)(f)
        }
    }
}

// CastAny 从any类型的Seq转换为T类型的Seq,强制转换
func CastAny[T any](seq Seq[any]) Seq[T] {
    return func(c func(T)) { seq(func(t any) { c((t).(T)) }) }
}

// CastAnyT 从any类型的Seq转换为T类型的Seq,强制转换,简便写法
func CastAnyT[T any](seq Seq[any], _ T) Seq[T] {
    return func(c func(T)) { seq(func(t any) { c((t).(T)) }) }
}

// Map 每个元素自定义转换
func Map[T, E any](seq Seq[T], cast func(T) E) Seq[E] {
    return func(c func(E)) { seq(func(t T) { c(cast(t)) }) }
}

// Join 合并多个Seq
func Join[T any](seqs ...Seq[T]) Seq[T] {
    return func(c func(T)) {
        for _, seq := range seqs {
            seq(func(t T) { c(t) })
        }
    }
}

// JoinL 合并2个不同Seq,右边转换为左边的类型
func JoinL[T, E any](seq1 Seq[T], seq2 Seq[E], cast func(E) T) Seq[T] {
    return func(c func(T)) {
        seq1(func(t T) { c(t) })
        seq2(func(t E) { c(cast(t)) })
    }
}

// JoinBy 合并2个不同Seq,统一转换为新类型
func JoinBy[T, E, R any](seq1 Seq[T], cast1 func(T) R, seq2 Seq[E], cast2 func(E) R) Seq[R] {
    return func(c func(R)) {
        seq1(func(t T) { c(cast1(t)) })
        seq2(func(t E) { c(cast2(t)) })
    }
}
