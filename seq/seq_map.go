package seq

//======转换,添加或修改内部元素========

// Join 合并多个Seq
func (t Seq[T]) Join(seqs ...Seq[T]) Seq[T] {
    return func(c func(T)) {
        defer stopRecover()
        t(func(t T) { c(t) })
        for _, seq := range seqs {
            seq(func(t T) { c(t) })
        }
    }
}

//// JoinF 合并Seq
//func (t Seq[T]) JoinF(seq Seq[any], cast func(any) T) Seq[T] {
//    return func(c func(T)) {
//        t(func(t T) { c(t) })
//        seq(func(t any) { c(cast(t)) })
//    }
//}

// Add 直接添加元素
func (t Seq[T]) Add(ts ...T) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        for _, e := range ts {
            c(e)
        }
    }
}

//// AddF 直接添加需要转换的元素
//func (t Seq[T]) AddF(cast func(any) T, ts ...any) Seq[T] {
//    return func(c func(T)) {
//        t(func(t T) { c(t) })
//        for _, e := range ts {
//            c(cast(e))
//        }
//    }
//}

// AddIf 满足条件才添加元素
func (t Seq[T]) AddIf(condition bool, ts ...T) Seq[T] {
    if !condition {
        return t
    }
    return t.Add(ts...)
}

// AddIfF 满足条件才添加元素
func (t Seq[T]) AddIfF(condition func(T) bool, ts ...T) Seq[T] {
    return t.Join(FromSlice(ts).Filter(condition))
}

//// AddFIf 满足条件才添加需要转换的元素
//func (t Seq[T]) AddFIf(condition bool, cast func(any) T, ts ...any) Seq[T] {
//    if !condition {
//        return t
//    }
//    return t.AddF(cast, ts...)
//}
//
//// AddFIfF 满足条件才添加需要转换的元素
//func (t Seq[T]) AddFIfF(condition func(Seq[T]) bool, cast func(any) T, ts ...any) Seq[T] {
//    if !condition(t) {
//        return t
//    }
//    return t.AddF(cast, ts...)
//}

// MapSliceN 每n个元素合并为[]T,由于golang泛型问题,不能使用Seq[[]T],使用 CastAny 转换为Seq[[]T]
func MapSliceN[T any](t Seq[T], n int) Seq[any] {
    return MapSliceBy(t, func(t T, ts []T) bool { return len(ts) == n })
}

// MapSliceBy 自定义元素合并为[]T,由于golang泛型问题,不能返回[]Seq[T],使用 CastAny 转换为Seq[[]T]
func MapSliceBy[T any](t Seq[T], f func(T, []T) bool) Seq[any] {
    return func(c func(any)) {
        var ts []T
        t(func(t T) {
            ts = append(ts, t)
            if f(t, ts) {
                c(append([]T(nil), ts...))
                ts = ts[:0]
            }
        })
        if len(ts) > 0 {
            c(ts)
        }
    }
}
