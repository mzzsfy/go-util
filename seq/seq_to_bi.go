package seq

//========转换为 BiSeq ================

// MergeBiInt 与一个Iterator合并,参数为一个迭代器,返回 BiSeq[int, T]
func (t Seq[T]) MergeBiInt(iterator Iterator[int]) BiSeq[int, T] {
    return BiFrom(func(f1 func(int, T)) {
        t.Stoppable()(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(x, t)
        })
    })
}

// MergeBiIntRight 与一个Iterator合并,参数为一个迭代器,返回 BiSeq[T, int]
func (t Seq[T]) MergeBiIntRight(iterator Iterator[int]) BiSeq[T, int] {
    return BiFrom(func(f1 func(T, int)) {
        t.Stoppable()(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(t, x)
        })
    })
}

// MergeBiString 与一个Iterator合并,参数为一个迭代器,返回 BiSeq[string, T]
func (t Seq[T]) MergeBiString(iterator Iterator[string]) BiSeq[string, T] {
    return BiFrom(func(f1 func(string, T)) {
        t.Stoppable()(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(x, t)
        })
    })
}

// MergeBiStringRight 与一个Iterator合并,参数为一个迭代器,返回 BiSeq[T, string]
func (t Seq[T]) MergeBiStringRight(iterator Iterator[string]) BiSeq[T, string] {
    return BiFrom(func(f1 func(T, string)) {
        t.Stoppable()(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(t, x)
        })
    })
}

// MergeBiAny 与一个Iterator合并,参数为一个迭代器,返回 BiSeq
func (t Seq[T]) MergeBiAny(iterator Iterator[any]) BiSeq[any, T] {
    return BiFrom(func(f1 func(any, T)) {
        t.Stoppable()(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(x, t)
        })
    })
}

// MergeBiAnyRight 与一个Iterator合并,参数为一个迭代器,返回 BiSeq
func (t Seq[T]) MergeBiAnyRight(iterator Iterator[any]) BiSeq[any, T] {
    return BiFrom(func(f1 func(any, T)) {
        t.Stoppable()(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(x, t)
        })
    })
}

// MapBiSerialNumber 为每个元素生成一个序列号,并转换为 BiSeq,参数Range为规则,参考 IteratorInt
func (t Seq[T]) MapBiSerialNumber(Range ...int) BiSeq[int, T] {
    return BiFrom(func(f1 func(int, T)) {
        r := makeRange(Range...)
        t.Stoppable()(func(t T) {
            f1(r(), t)
        })
    })
}

// MapBiInt 每个元素获取一个int,并转换为 BiSeq
func (t Seq[T]) MapBiInt(f func(T) int) BiSeq[int, T] {
    return BiFrom(func(f1 func(int, T)) { t(func(t T) { f1(f(t), t) }) })
}

// MapBiString 每个元素获取一个String,并转换为 BiSeq
func (t Seq[T]) MapBiString(f func(T) string) BiSeq[string, T] {
    return BiFrom(func(f1 func(string, T)) { t(func(t T) { f1(f(t), t) }) })
}

// MapBiAny 每个元素获取一个值,并转换为 BiSeq
func (t Seq[T]) MapBiAny(f func(T) any) BiSeq[any, T] {
    return BiFrom(func(f1 func(any, T)) { t(func(t T) { f1(f(t), t) }) })
}

// MapBiAnyRight 每个元素获取一个值,并转换为 BiSeq
func (t Seq[T]) MapBiAnyRight(f func(T) any) BiSeq[T, any] {
    return BiFrom(func(f1 func(T, any)) { t(func(t T) { f1(t, f(t)) }) })
}
