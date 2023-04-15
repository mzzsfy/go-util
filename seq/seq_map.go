package seq

//======转换========

// AsyncMap 每个元素转换为any,使用 Sync() 保证消费不竞争
// order 是否保持顺序,大于0保持顺序,默认不保持顺序
// order 第二个参数,并发数
func (t Seq[T]) AsyncMap(f func(T) any, order ...int) Seq[any] {
    o := false
    c := 0
    if len(order) > 0 {
        o = order[0] > 0
    }
    if len(order) > 1 {
        c = order[1]
    }
    if o {
        return t.MapBiSerialNumber().Parallel(c).
            MapV(func(i int, t T) any { return f(t) }).SortK(LessT[int]).SeqV()
    } else {
        return t.Parallel(c).Map(f)
    }
}

// Map 每个元素转换为any
func (t Seq[T]) Map(f func(T) any) Seq[any] {
    return func(c func(any)) { t(func(t T) { c(f(t)) }) }
}

// FlatMap 每个元素转换为Seq,并扁平化
func (t Seq[T]) FlatMap(f func(T) Seq[any]) Seq[any] {
    return func(c func(any)) { t(func(t T) { f(t).ForEach(c) }) }
}

// MapString 每个元素转换为字符串
func (t Seq[T]) MapString(f func(T) string) Seq[string] {
    return func(c func(string)) { t(func(t T) { c(f(t)) }) }
}

func (t Seq[T]) MergeBiInt(iterator Iterator[int]) BiSeq[int, T] {
    return BiFrom(func(f1 func(int, T)) {
        t.ConsumeTillStop(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(x, t)
        })
    })
}

func (t Seq[T]) MergeBiIntRight(iterator Iterator[int]) BiSeq[T, int] {
    return BiFrom(func(f1 func(T, int)) {
        t.ConsumeTillStop(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(t, x)
        })
    })
}

func (t Seq[T]) MergeBiAny(iterator Iterator[any]) BiSeq[any, T] {
    return BiFrom(func(f1 func(any, T)) {
        t.ConsumeTillStop(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(x, t)
        })
    })
}

func (t Seq[T]) MergeBiAnyRight(iterator Iterator[any]) BiSeq[any, T] {
    return BiFrom(func(f1 func(any, T)) {
        t.ConsumeTillStop(func(t T) {
            x, exist := iterator()
            if !exist {
                panic(&Stop)
            }
            f1(x, t)
        })
    })
}

// MapBiSerialNumber 为每个元素生成一个序列号,并转换为BiSeq,参数rang为规则,参考 IteratorInt
func (t Seq[T]) MapBiSerialNumber(rang ...int) BiSeq[int, T] {
    it := IteratorInt(rang...)
    return BiFrom(func(f1 func(int, T)) {
        t.ConsumeTillStop(func(t T) {
            i, exist := it()
            if !exist {
                panic(&Stop)
            }
            f1(i, t)
        })
    })
}

// MapBiInt 每个元素获取一个int,并转换为BiSeq
func (t Seq[T]) MapBiInt(f func(T) int) BiSeq[int, T] {
    return BiFrom(func(f1 func(int, T)) { t(func(t T) { f1(f(t), t) }) })
}

// MapBiString 每个元素获取一个String,并转换为BiSeq
func (t Seq[T]) MapBiString(f func(T) string) BiSeq[string, T] {
    return BiFrom(func(f1 func(string, T)) { t(func(t T) { f1(f(t), t) }) })
}

// MapBiAny 每个元素获取一个值,并转换为BiSeq
func (t Seq[T]) MapBiAny(f func(T) any) BiSeq[any, T] {
    return BiFrom(func(f1 func(any, T)) { t(func(t T) { f1(f(t), t) }) })
}

// MapBiAnyRight 每个元素获取一个值,并转换为BiSeq
func (t Seq[T]) MapBiAnyRight(f func(T) any) BiSeq[T, any] {
    return BiFrom(func(f1 func(T, any)) { t(func(t T) { f1(t, f(t)) }) })
}

// MapInt 每个元素转换为int
func (t Seq[T]) MapInt(f func(T) int) Seq[int] {
    return func(c func(int)) { t(func(t T) { c(f(t)) }) }
}

// MapFloat32 每个元素转换为float32
func (t Seq[T]) MapFloat32(f func(T) float32) Seq[float32] {
    return func(c func(float32)) { t(func(t T) { c(f(t)) }) }
}

// MapFloat64 每个元素转换为float64
func (t Seq[T]) MapFloat64(f func(T) float64) Seq[float64] {
    return func(c func(float64)) { t(func(t T) { c(f(t)) }) }
}

// Join 合并多个Seq
func (t Seq[T]) Join(seqs ...Seq[T]) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        for _, seq := range seqs {
            seq(func(t T) { c(t) })
        }
    }
}

// Add 添加单个元素
func (t Seq[T]) Add(ts ...T) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        for _, e := range ts {
            c(e)
        }
    }
}

// AddF 添加单个元素
func (t Seq[T]) AddF(cast func(any) T, ts ...any) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        for _, e := range ts {
            c(cast(e))
        }
    }
}

// JoinF 合并Seq
func (t Seq[T]) JoinF(seq Seq[any], cast func(any) T) Seq[T] {
    return func(c func(T)) {
        t(func(t T) { c(t) })
        seq(func(t any) { c(cast(t)) })
    }
}
