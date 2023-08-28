package biMapper

import "github.com/mzzsfy/go-util/seq"

// 由于产生的泛型太多,大部分情况下用不到这么多自定义转换,拆分包后提升编译速度

type ValueMapper[K, V any] func(func(K, V))

// MapV 每个元素自定义转换V为any,用于连续转换操作,使用 BiCastAny 进行恢复泛型
func (t ValueMapper[K, V]) MapV(f func(K, V) any) seq.BiSeq[K, any] {
    return func(c func(K, any)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// MapVInt 转换V的类型
func (t ValueMapper[K, V]) MapVInt(f func(K, V) int) seq.BiSeq[K, int] {
    return func(c func(K, int)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// MapVInt32 转换V的类型
func (t ValueMapper[K, V]) MapVInt32(f func(K, V) int32) seq.BiSeq[K, int32] {
    return func(c func(K, int32)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// MapVInt64 转换V的类型
func (t ValueMapper[K, V]) MapVInt64(f func(K, V) int64) seq.BiSeq[K, int64] {
    return func(c func(K, int64)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// MapVFloat32 转换V的类型
func (t ValueMapper[K, V]) MapVFloat32(f func(K, V) float32) seq.BiSeq[K, float32] {
    return func(c func(K, float32)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// MapVFloat64 转换V的类型
func (t ValueMapper[K, V]) MapVFloat64(f func(K, V) float64) seq.BiSeq[K, float64] {
    return func(c func(K, float64)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// MapVString 转换V的类型
func (t ValueMapper[K, V]) MapVString(f func(K, V) string) seq.BiSeq[K, string] {
    return func(c func(K, string)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// MapVBytes 转换V的类型
func (t ValueMapper[K, V]) MapVBytes(f func(K, V) []byte) seq.BiSeq[K, []byte] {
    return func(c func(K, []byte)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// MapFlatV V扁平化
func (t ValueMapper[K, V]) MapFlatV(f func(K, V) seq.Seq[any]) seq.BiSeq[K, any] {
    return func(c func(K, any)) {
        t(func(k K, v V) {
            s := f(k, v)
            s(func(a any) {
                c(k, a)
            })
        })
    }
}

// MapFlatVInt V扁平化
func (t ValueMapper[K, V]) MapFlatVInt(f func(K, V) seq.Seq[int]) seq.BiSeq[K, int] {
    return func(c func(K, int)) {
        t(func(k K, v V) {
            s := f(k, v)
            s(func(a int) {
                c(k, a)
            })
        })
    }
}

// MapFlatVInt32 V扁平化
func (t ValueMapper[K, V]) MapFlatVInt32(f func(K, V) seq.Seq[int32]) seq.BiSeq[K, int32] {
    return func(c func(K, int32)) {
        t(func(k K, v V) {
            s := f(k, v)
            s(func(a int32) {
                c(k, a)
            })
        })
    }
}

// MapFlatVInt64 V扁平化
func (t ValueMapper[K, V]) MapFlatVInt64(f func(K, V) seq.Seq[int64]) seq.BiSeq[K, int64] {
    return func(c func(K, int64)) {
        t(func(k K, v V) {
            s := f(k, v)
            s(func(a int64) {
                c(k, a)
            })
        })
    }
}

// MapFlatVFloat32 V扁平化
func (t ValueMapper[K, V]) MapFlatVFloat32(f func(K, V) seq.Seq[float32]) seq.BiSeq[K, float32] {
    return func(c func(K, float32)) {
        t(func(k K, v V) {
            s := f(k, v)
            s(func(a float32) {
                c(k, a)
            })
        })
    }
}

// MapFlatVFloat64 V扁平化
func (t ValueMapper[K, V]) MapFlatVFloat64(f func(K, V) seq.Seq[float64]) seq.BiSeq[K, float64] {
    return func(c func(K, float64)) {
        t(func(k K, v V) {
            s := f(k, v)
            s(func(a float64) {
                c(k, a)
            })
        })
    }
}

// MapFlatVString V扁平化
func (t ValueMapper[K, V]) MapFlatVString(f func(K, V) seq.Seq[string]) seq.BiSeq[K, string] {
    return func(c func(K, string)) {
        t(func(k K, v V) {
            s := f(k, v)
            s(func(a string) {
                c(k, a)
            })
        })
    }
}

// MapFlatVBytes V扁平化
func (t ValueMapper[K, V]) MapFlatVBytes(f func(K, V) seq.Seq[[]byte]) seq.BiSeq[K, []byte] {
    return func(c func(K, []byte)) {
        t(func(k K, v V) {
            s := f(k, v)
            s(func(a []byte) {
                c(k, a)
            })
        })
    }
}
