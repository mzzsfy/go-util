package seq

//========转换为 BiSeq ================

func (t Seq[T]) MapFlatBiInt(f func(T) Seq[int]) BiSeq[int, T] {
    return BiFrom(func(f1 func(int, T)) { t(func(t T) { f(t)(func(v int) { f1(v, t) }) }) })
}

func (t Seq[T]) MapFlatBiString(f func(T) Seq[string]) BiSeq[string, T] {
    return BiFrom(func(f1 func(string, T)) { t(func(t T) { f(t)(func(v string) { f1(v, t) }) }) })
}

func (t Seq[T]) MapFlatBiAny(f func(T) Seq[any]) BiSeq[any, T] {
    return BiFrom(func(f1 func(any, T)) { t(func(t T) { f(t)(func(v any) { f1(v, t) }) }) })
}

func (t Seq[T]) MapFlatBiAnyRight(f func(T) Seq[any]) BiSeq[T, any] {
    return BiFrom(func(f1 func(T, any)) { t(func(t T) { f(t)(func(v any) { f1(t, v) }) }) })
}
