package seq

import (
    "strings"
)

//======消费========

// Complete 消费所有元素
func (t Seq[T]) Complete() { t(func(_ T) {}) }

// ForEach 每个元素执行f
func (t Seq[T]) ForEach(f func(T)) { t(f) }

// AsyncEach 每个元素执行f,并行执行
func (t Seq[T]) AsyncEach(f func(T)) { t.Parallel().ForEach(f) }

// First 有则返回第一个元素,无则返回nil
func (t Seq[T]) First() *T {
    var r *T
    t.Take(1)(func(t T) { r = &t })
    return r
}

// FirstOr 有则返回第一个元素,无则返回默认值
func (t Seq[T]) FirstOr(d T) T {
    var r *T
    exist := false
    t.Take(1)(func(t T) { r = &t; exist = true })
    if exist {
        return *r
    }
    return d
}

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

// GroupBy 元素分组,每个组保留所有元素
func (t Seq[T]) GroupBy(f func(T) any) map[any][]T {
    r := make(map[any][]T)
    t.ForEach(func(t T) {
        k := f(t)
        r[k] = append(r[k], t)
    })
    return r
}

// GroupByFirst 元素分组,每个组只保留第一个元素
func (t Seq[T]) GroupByFirst(f func(T) any) map[any]T {
    r := make(map[any]T)
    t.ForEach(func(t T) {
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
    t.ForEach(func(t T) {
        k := f(t)
        r[k] = t
    })
    return r
}

// Reduce 求值
func (t Seq[T]) Reduce(f func(T, any) any, init any) any {
    t.ForEach(func(t T) { init = f(t, init) })
    return init
}

// ToSlice 转换为切片
func (t Seq[T]) ToSlice() []T {
    var r []T
    t(func(t T) { r = append(r, t) })
    return r
}

// Count 计数
func (t Seq[T]) Count() int64 {
    var r int64
    t(func(t T) { r++ })
    return r
}

// JoinStringF 拼接为字符串,自定义转换函数
func (t Seq[T]) JoinStringF(f func(T) string, delimiter ...string) string {
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

// JoinString 拼接为字符串
func (t Seq[T]) JoinString(delimiter ...string) string {
    var x T
    return t.JoinStringF(getToStringFn(x), delimiter...)
}
