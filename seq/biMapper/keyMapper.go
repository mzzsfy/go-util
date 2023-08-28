package biMapper

import "github.com/mzzsfy/go-util/seq"

// 由于产生的泛型太多,大部分情况下用不到这么多自定义转换,拆分包提升编译速度

type KeyMapper[K, V any] func(func(K, V))

// MapK 每个元素自定义转换K为any,用于连续转换操作,使用 BiCastAny 进行恢复泛型
func (t KeyMapper[K, V]) MapK(f func(K, V) any) seq.BiSeq[any, V] {
    return func(c func(any, V)) { t(func(k K, v V) { c(f(k, v), v) }) }
}

// MapKInt 转换K的类型
func (t KeyMapper[K, V]) MapKInt(f func(K, V) int) seq.BiSeq[int, V] {
    return func(c func(int, V)) { t(func(k K, v V) { c(f(k, v), v) }) }
}

// MapKInt64 转换K的类型
func (t KeyMapper[K, V]) MapKInt64(f func(K, V) int64) seq.BiSeq[int64, V] {
    return func(c func(int64, V)) { t(func(k K, v V) { c(f(k, v), v) }) }
}

// MapKString 转换K的类型
func (t KeyMapper[K, V]) MapKString(f func(K, V) string) seq.BiSeq[string, V] {
    return func(c func(string, V)) { t(func(k K, v V) { c(f(k, v), v) }) }
}

// MapFlatK K扁平化
func (t KeyMapper[K, V]) MapFlatK(f func(K, V) seq.Seq[any]) seq.BiSeq[any, V] {
    return func(c func(any, V)) {
        t(func(k K, v V) {
            s := f(k, v)
            s(func(a any) {
                c(a, v)
            })
        })
    }
}
