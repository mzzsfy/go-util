package seq

//========转换为 Seq ================

//请使用: seq.From[T]

//// SeqK 转换为只保留K的Seq
//func (t BiSeq[K, V]) SeqK() Seq[K] {
//    return func(c func(K)) { t(func(k K, v V) { c(k) }) }
//}
//
//// SeqV 转换为只保留V的Seq
//func (t BiSeq[K, V]) SeqV() Seq[V] {
//    return func(c func(V)) { t(func(k K, v V) { c(v) }) }
//}
//
//// SeqBy 转换为Seq[any],自定义转换
//func (t BiSeq[K, V]) SeqBy(f func(K, V) any) Seq[any] {
//    return func(c func(any)) { t(func(k K, v V) { c(f(k, v)) }) }
//}
//
//// SeqKBy 转换为只保留K的Seq,并自定义转换
//func (t BiSeq[K, V]) SeqKBy(f func(K, V) K) Seq[K] {
//    return func(c func(K)) { t(func(k K, v V) { c(f(k, v)) }) }
//}
//
//// SeqVBy 转换为只保留V的Seq,并自定义转换
//func (t BiSeq[K, V]) SeqVBy(f func(K, V) V) Seq[V] {
//    return func(c func(V)) { t(func(k K, v V) { c(f(k, v)) }) }
//}
//
//// MapStringBy 转换为Seq[string]
//func (t BiSeq[K, V]) MapStringBy(f func(K, V) string) Seq[string] {
//    return func(c func(string)) { t(func(k K, v V) { c(f(k, v)) }) }
//}
//
//// MapIntBy 转换为Seq[int]
//func (t BiSeq[K, V]) MapIntBy(f func(K, V) int) Seq[int] {
//    return func(c func(int)) { t(func(k K, v V) { c(f(k, v)) }) }
//}
//
//// MapSliceN 每n个元素合并为[]T,由于golang泛型问题,不能使用[]BiTuple[K,V],使用 BiCastAny 进行恢复泛型[]BiTuple[K,V]
//func (t BiSeq[K, V]) MapSliceN(n int) Seq[any] {
//    return t.MapSliceBy(func(k K, v V, ts any) bool { return len(ts.([]BiTuple[K, V])) == n })
//}
//
////MapSliceBy 自定义元素合并为[]T,由于golang泛型问题,不能使用[]BiTuple[K,V],使用 BiCastAny 进行恢复泛型[]BiTuple[K,V]
//func (t BiSeq[K, V]) MapSliceBy(f func(K, V, any) bool) Seq[any] {
//    return func(c func(any)) {
//        var ts []BiTuple[K, V]
//        t(func(k K, v V) {
//            ts = append(ts, BiTuple[K, V]{k, v})
//            if f(k, v, ts) {
//                c(ts)
//                ts = nil
//            }
//        })
//        if len(ts) > 0 {
//            c(ts)
//        }
//    }
//}
