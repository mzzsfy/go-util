package seq

import (
    "fmt"
    "math"
    "sort"
    "strings"
    "sync"
)

//参考: https://mp.weixin.qq.com/s/v-HMKBWxtz1iakxFL09PDw

// Seq 一种特殊的集合,可以用于链式操作
type Seq[T any] func(t func(T))

//======生成========

// FromSlice 从数组生成Seq
func FromSlice[T any](arr []T) Seq[T] {
    return func(t func(T)) {
        for _, v := range arr {
            t(v)
        }
    }
}

func From[T any](f Seq[T]) Seq[T] {
    return f
}

// FromIntSeq 生成整数序列,可以自定义起始值,结束值,步长
func FromIntSeq(rang ...int) Seq[int] {
    start := 0
    end := math.MaxInt
    step := 1
    if len(rang) > 0 {
        start = rang[0]
    }
    if len(rang) > 1 {
        end = rang[1]
    }
    if len(rang) > 2 {
        step = rang[2]
        if step == 0 {
            panic("step can not be 0")
        }
    }
    if step > 0 {
        if start > end {
            panic(fmt.Sprintf("step is %d ,start(%d) can not be greater than end(%d)", step, start, end))
        }
    } else {
        if start < end {
            panic(fmt.Sprintf("step is %d ,start(%d) can not be less than end(%d)", step, start, end))
        }
    }
    return func(t func(int)) {
        if step > 0 {
            for i := start; i <= end; i += step {
                t(i)
            }
        } else {
            for i := start; i >= end; i += step {
                t(i)
            }
        }
    }
}

// Unit 生成单元素的Seq
func Unit[T any](e T) Seq[T] {
    return func(t func(T)) { t(e) }
}

// UnitRepeat 生成重复产生单元素的Seq
func UnitRepeat[T any](e T, limit ...int) Seq[T] {
    return func(t func(T)) {
        if len(limit) > 0 && limit[0] > 0 {
            l := limit[0]
            for i := 0; i < l; i++ {
                t(e)
            }
        } else {
            for {
                t(e)
            }
        }

    }
}

// MapAny 从any类型的Seq转换为T类型的Seq,强制转换
func MapAny[T any](seq Seq[any]) Seq[T] {
    return func(c func(T)) { seq(func(t any) { c((t).(T)) }) }
}

// Map 每个元素自定义转换
func Map[T, E any](seq Seq[T], cast func(T) E) Seq[E] {
    return func(c func(E)) { seq(func(t T) { c(cast(t)) }) }
}

//======转换========

// Map 每个元素转换为any
func (t Seq[T]) Map(f func(T) any) Seq[any] {
    return func(c func(any)) { t(func(t T) { c(f(t)) }) }
}

// FlatMap 每个元素转换为Seq,并扁平化
func (t Seq[T]) FlatMap(f func(T) Seq[any]) Seq[any] {
    return func(c func(any)) {
        t(func(t T) {
            s := f(t)
            s.DoEach(c)
        })
    }
}

// MapString 每个元素转换为字符串
func (t Seq[T]) MapString(f func(T) string) Seq[string] {
    return func(c func(string)) { t(func(t T) { c(f(t)) }) }
}

// MapInt 每个元素转换为int
func (t Seq[T]) MapInt(f func(T) int) Seq[int] {
    return func(c func(int)) { t(func(t T) { c(f(t)) }) }
}

//======HOOK========

// OnEach 每个元素额外执行
func (t Seq[T]) OnEach(f func(T)) Seq[T] {
    return func(c func(T)) {
        t(func(t T) {
            f(t)
            c(t)
        })

    }
}

// Parallel 对后续操作启用并行执行
func (t Seq[T]) Parallel() Seq[T] {
    return func(c func(T)) {
        t.Map(func(t T) any {
            lock := sync.Mutex{}
            lock.Lock()
            go func() {
                defer lock.Unlock()
                c(t)
            }()
            return &lock
        }).Cache()(func(t any) {
            lock := t.(sync.Locker)
            lock.Lock()
        })

    }
}

//======消费========

// DoEach 每个元素执行f
func (t Seq[T]) DoEach(f func(T)) { t(f) }

// AsyncEach 每个元素执行f,并行执行
func (t Seq[T]) AsyncEach(f func(T)) { t.Parallel().DoEach(f) }

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
    t.DoEach(func(t T) {
        k := f(t)
        r[k] = append(r[k], t)
    })
    return r
}

// GroupByFirst 元素分组,每个组只保留第一个元素
func (t Seq[T]) GroupByFirst(f func(T) any) map[any]T {
    r := make(map[any]T)
    t.DoEach(func(t T) {
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
    t.DoEach(func(t T) {
        k := f(t)
        r[k] = t
    })
    return r
}

// Reduce 求值
func (t Seq[T]) Reduce(f func(T, any) any, init any) any {
    t.DoEach(func(t T) { init = f(t, init) })
    return init
}

// Sort 排序
func (t Seq[T]) Sort(less func(T, T) bool) Seq[T] {
    var r []T
    t(func(t T) { r = append(r, t) })
    sort.Slice(r, func(i, j int) bool { return less(r[i], r[j]) })
    return FromSlice(r)
}

// Distinct 去重
func (t Seq[T]) Distinct(eq func(T, T) bool) any {
    var r []T
    t(func(t T) {
        for _, v := range r {
            if eq(t, v) {
                return
            }
        }
        r = append(r, t)
    })
    return FromSlice(r)
}

// ToSlice 转换为切片
func (t Seq[T]) ToSlice() []T {
    var r []T
    t(func(t T) { r = append(r, t) })
    return r
}

// Cache 缓存Seq,使该Seq可以多次重复消费,并保证前面内容不会重复执行
func (t Seq[T]) Cache() Seq[T] {
    var r []T
    t(func(t T) { r = append(r, t) })
    return FromSlice(r)
}

// Complete 消费所有元素
func (t Seq[T]) Complete() { t(func(_ T) {}) }

// JoinString 拼接为字符串
func (t Seq[T]) JoinString(f func(T) string, delimiter ...string) string {
    sb := strings.Builder{}
    d := ""
    if len(delimiter) > 0 {
        d = delimiter[0]
    }
    t.MapString(f).DoEach(func(s string) {
        if d != "" && sb.Len() > 0 {
            sb.WriteString(d)
        }
        sb.WriteString(s)
    })
    return sb.String()
}

//======控制========

// consumeTillStop 消费直到遇到stop
func (t Seq[T]) consumeTillStop(f func(T)) {
    defer func() {
        a := recover()
        if a != nil && a != &stop {
            panic(a)
        }
    }()
    t(f)
}

// Filter 过滤元素,只保留满足条件的元素,即f(t) == true保留
func (t Seq[T]) Filter(f func(T) bool) Seq[T] {
    return func(c func(T)) {
        t(func(t T) {
            if f(t) {
                c(t)
            }
        })

    }
}

// Take 保留前n个元素
func (t Seq[T]) Take(n int) Seq[T] {
    return func(c func(T)) {
        t.consumeTillStop(func(e T) {
            n--
            if n >= 0 {
                c(e)
            } else {
                panic(&stop)
            }
        })

    }
}

// Drop 跳过前n个元素
func (t Seq[T]) Drop(n int) Seq[T] {
    return func(c func(T)) {
        t(func(e T) {
            if n <= 0 {
                c(e)
            } else {
                n--
            }
        })
    }
}

var (
    stop *bool
)
