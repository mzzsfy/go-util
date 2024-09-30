package seq

//// MapBi2AnyAny 每个元素获取两个值,并转换为 BiSeq
//func (t Seq[T]) MapBi2AnyAny(f func(T) (any, any)) BiSeq[any, any] {
//    return BiFrom(func(f1 func(any, any)) {
//        t(func(t T) {
//            k, v := f(t)
//            f1(k, v)
//        })
//    })
//}
//
//// MapBi2StringString 每个元素获取两个值,并转换为 BiSeq
//func (t Seq[T]) MapBi2StringString(f func(T) (string, string)) BiSeq[string, string] {
//    return BiFrom(func(f1 func(string, string)) {
//        t(func(t T) {
//            k, v := f(t)
//            f1(k, v)
//        })
//    })
//}
//
//// MapBi2IntAny 每个元素获取两个值,并转换为 BiSeq
//func (t Seq[T]) MapBi2IntAny(f func(T) (int, any)) BiSeq[int, any] {
//    return BiFrom(func(f1 func(int, any)) {
//        t(func(t T) {
//            k, v := f(t)
//            f1(k, v)
//        })
//    })
//}
//
//// MapBi2StringAny 每个元素获取两个值,并转换为 BiSeq
//func (t Seq[T]) MapBi2StringAny(f func(T) (string, any)) BiSeq[string, any] {
//    return BiFrom(func(f1 func(string, any)) {
//        t(func(t T) {
//            k, v := f(t)
//            f1(k, v)
//        })
//    })
//}
