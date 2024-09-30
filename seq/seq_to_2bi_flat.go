package seq

//========转换为 BiSeq ================

//func (t Seq[T]) MapFlatBi2AnyAny(f func(T) BiSeq[any, any]) BiSeq[any, any] {
//    return BiFrom(func(f1 func(any, any)) { t(func(t T) { f(t)(func(k, v any) { f1(k, v) }) }) })
//}
//func (t Seq[T]) MapFlatBi2StringString(f func(T) BiSeq[string, string]) BiSeq[string, string] {
//    return BiFrom(func(f1 func(string, string)) { t(func(t T) { f(t)(func(k, v string) { f1(k, v) }) }) })
//}
//
//func (t Seq[T]) MapFlatBi2StringAny(f func(T) BiSeq[string, any]) BiSeq[string, any] {
//    return BiFrom(func(f1 func(string, any)) { t(func(t T) { f(t)(func(k string, v any) { f1(k, v) }) }) })
//}
//
//func (t Seq[T]) MapFlatBi2IntAny(f func(T) BiSeq[int, any]) BiSeq[int, any] {
//    return BiFrom(func(f1 func(int, any)) { t(func(t T) { f(t)(func(k int, v any) { f1(k, v) }) }) })
//}

func MapFlatBi[K, V, T any](t Seq[T], f func(T) BiSeq[K, V]) BiSeq[K, V] {
    return BiFrom(func(f1 func(K, V)) { t(func(t T) { f(t)(func(k K, v V) { f1(k, v) }) }) })
}

func MapFlatBiK[K, T any](t Seq[T], f func(T) BiSeq[K, T]) BiSeq[K, T] {
    return BiFrom(func(f1 func(K, T)) { t(func(t T) { f(t)(func(k K, v T) { f1(k, v) }) }) })
}

func MapFlatBiV[V, T any](t Seq[T], f func(T) BiSeq[T, V]) BiSeq[T, V] {
    return BiFrom(func(f1 func(T, V)) { t(func(t T) { f(t)(func(k T, v V) { f1(k, v) }) }) })
}
