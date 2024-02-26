package seq

import "strings"

//======消费========

// Complete 消费所有元素
func (t BiSeq[K, V]) Complete() { t(func(_ K, _ V) {}) }

// ForEach 每个元素执行f
func (t BiSeq[K, V]) ForEach(f func(K, V)) { t(f) }

// AsyncEach 每个元素执行f,并行执行
func (t BiSeq[K, V]) AsyncEach(f func(K, V)) { t.Parallel().ForEach(f) }

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

// Last 获取最后一个元素,无则返回nil
func (t BiSeq[K, V]) Last() (*K, *V) {
    var rk *K
    var rv *V
    t(func(k K, v V) {
        rk = &k
        rv = &v
    })
    return rk, rv
}

// LastOr 获取最后一个元素,无则返回默认值
func (t BiSeq[K, V]) LastOr(k K, v V) (K, V) {
    var rk *K
    var rv *V
    exist := false
    t(func(k K, v V) {
        rk = &k
        rv = &v
        exist = true
    })
    if exist {
        return *rk, *rv
    }
    return k, v
}

// LastOrF 获取最后一个元素,无则返回f的值
func (t BiSeq[K, V]) LastOrF(f func() (K, V)) (K, V) {
    var rk *K
    var rv *V
    exist := false
    t(func(k K, v V) {
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

// Count 计数
func (t BiSeq[K, V]) Count() int {
    var r int
    t(func(k K, v V) { r++ })
    return r
}

// Count64 计数
func (t BiSeq[K, V]) Count64() int64 {
    var r int64
    t(func(k K, v V) { r++ })
    return r
}

// SumBy 求和
func (t BiSeq[K, V]) SumBy(f func(K, V) int) int {
    var r int
    t(func(k K, v V) { r += f(k, v) })
    return r
}

// SumBy64 求和
func (t BiSeq[K, V]) SumBy64(f func(K, V) int64) int64 {
    var r int64
    t(func(k K, v V) { r += f(k, v) })
    return r
}

// SumByFloat32 求和
func (t BiSeq[K, V]) SumByFloat32(f func(K, V) float32) float32 {
    var r float32
    t(func(k K, v V) { r += f(k, v) })
    return r
}

// SumByFloat64 求和
func (t BiSeq[K, V]) SumByFloat64(f func(K, V) float64) float64 {
    var r float64
    t(func(k K, v V) { r += f(k, v) })
    return r
}

// JoinStringBy 拼接为字符串
func (t BiSeq[K, V]) JoinStringBy(f func(K, V) string, delimiter ...string) string {
    sb := strings.Builder{}
    d := ""
    if len(delimiter) > 0 {
        d = delimiter[0]
    }
    t.MapStringBy(f)(func(s string) {
        if d != "" && sb.Len() > 0 {
            sb.WriteString(d)
        }
        sb.WriteString(s)
    })
    return sb.String()
}

// Reduce 求值
func (t BiSeq[K, V]) Reduce(f func(K, V, any) any, init any) any {
    t(func(k K, v V) { init = f(k, v, init) })
    return init
}
