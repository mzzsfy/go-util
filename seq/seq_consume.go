package seq

import (
    "fmt"
    "strings"
)

//======消费,消耗掉当前seq========

//// Complete 消费所有元素
//func (t Seq[T]) Complete() { t(func(_ T) {}) }

// ForEach 每个元素执行f
func (t Seq[T]) ForEach(f func(T)) { t(f) }

//// AsyncEach 每个元素执行f,并行执行
//func (t Seq[T]) AsyncEach(f func(T)) { t.Parallel()(f) }

// FindFirstBy 找到第一个满足条件的元素
func (t Seq[T]) FindFirstBy(f func(T) bool) *T {
    return t.Filter(f).First()
}

//// FindBestBy 找到最优的元素
//func (t Seq[T]) FindBestBy(f func(previous *T, current T) (replace bool)) *T {
//    var r *T
//    t(func(t T) {
//        if f(r, t) {
//            r = &t
//        }
//    })
//    return r
//}

// First 有则返回第一个元素,无则返回nil
func (t Seq[T]) First() *T {
    var r *T
    t.Take(1)(func(t T) { r = &t })
    return r
}

//// FirstOr 有则返回第一个元素,无则返回默认值
//func (t Seq[T]) FirstOr(d T) T {
//    var r *T
//    exist := false
//    t.Take(1)(func(t T) { r = &t; exist = true })
//    if exist {
//        return *r
//    }
//    return d
//}

// FirstOrF 有则返回第一个元素,无则返回默认值
func (t Seq[T]) FirstOrF(d func() T) T {
    var r *T
    exist := false
    t.Take(1)(func(t T) { r = &t; exist = true })
    if exist {
        return *r
    }
    return d()
}

// Last 有则返回最后一个元素,无则返回nil
func (t Seq[T]) Last() *T {
    var r *T
    t(func(t T) { r = &t })
    return r
}

//// LastOr 有则返回最后一个元素,无则返回默认值
//func (t Seq[T]) LastOr(d T) T {
//    var r *T
//    exist := false
//    t(func(t T) { r = &t; exist = true })
//    if exist {
//        return *r
//    }
//    return d
//}

// LastOrF 有则返回最后一个元素,无则返回默认值
func (t Seq[T]) LastOrF(d func() T) T {
    var r *T
    exist := false
    t(func(t T) { r = &t; exist = true })
    if exist {
        return *r
    }
    return d()
}

// AnyMatch 任意匹配
func (t Seq[T]) AnyMatch(f func(T) bool) bool {
    r := false
    t.Filter(f).Take(1)(func(t T) { r = true })
    return r
}

// AllMatch 全部匹配
func (t Seq[T]) AllMatch(f func(T) bool) bool {
    r := true
    t.Filter(func(t T) bool { return !f(t) }).Take(1)(func(t T) { r = false })
    return r
}

// NonMatch 全部不匹配
func (t Seq[T]) NonMatch(f func(T) bool) bool {
    return !t.AnyMatch(f)
}

// GroupBy 元素分组,每个组保留所有元素
func (t Seq[T]) GroupBy(f func(T) any) map[any][]T {
    r := make(map[any][]T)
    t(func(t T) {
        k := f(t)
        r[k] = append(r[k], t)
    })
    return r
}

// GroupByFirst 元素分组,每个组只保留第一个元素
func (t Seq[T]) GroupByFirst(f func(T) any) map[any]T {
    r := make(map[any]T)
    t(func(t T) {
        k := f(t)
        if _, ok := r[k]; !ok {
            r[k] = t
        }
    })
    return r
}

// GroupByLast 元素分组,每个组只保留最后一个元素
func (t Seq[T]) GroupByLast(f func(T) any) map[any]T {
    r := make(map[any]T)
    t(func(t T) {
        k := f(t)
        r[k] = t
    })
    return r
}

// Reduce 自定义聚合
func (t Seq[T]) Reduce(f func(T, any) any, init any) any {
    t(func(t T) { init = f(t, init) })
    return init
}

// ToSlice 转换为切片
func (t Seq[T]) ToSlice() []T {
    var r []T
    t(func(t T) { r = append(r, t) })
    return r
}

// Count 计数
func (t Seq[T]) Count() int {
    var r int
    t(func(t T) { r++ })
    return r
}

//// Count64 计数
//func (t Seq[T]) Count64() int64 {
//    var r int64
//    t(func(t T) { r++ })
//    return r
//}

// SumBy 求和
func (t Seq[T]) SumBy(f func(T) int) int {
    var r int
    t(func(t T) { r += f(t) })
    return r
}

//// SumBy64 求和
//func (t Seq[T]) SumBy64(f func(T) int64) int64 {
//    var r int64
//    t(func(t T) { r += f(t) })
//    return r
//}

//// SumByFloat32 求和
//func (t Seq[T]) SumByFloat32(f func(T) float32) float32 {
//    var r float32
//    t(func(t T) { r += f(t) })
//    return r
//}

// SumByFloat64 求和
func (t Seq[T]) SumByFloat64(f func(T) float64) float64 {
    var r float64
    t(func(t T) { r += f(t) })
    return r
}

// JoinStringBy 拼接为字符串,自定义转换函数
func (t Seq[T]) JoinStringBy(f func(T) string, delimiter ...string) string {
    sb := strings.Builder{}
    d := ""
    if len(delimiter) > 0 {
        d = delimiter[0]
    }
    t.MapString(f).ForEach(func(s string) {
        if d != "" && sb.Len() > 0 {
            sb.WriteString(d)
        }
        sb.WriteString(s)
    })
    return sb.String()
}

// JoinString 拼接为字符串,使用默认转换函数
func (t Seq[T]) JoinString(delimiter ...string) string {
    var x T
    f := getToStringFn(x)
    if f != nil {
        return t.JoinStringBy(f.(func(T) string), delimiter...)
    }
    return t.JoinStringBy(func(t T) string { return fmt.Sprint(t) }, delimiter...)
}
