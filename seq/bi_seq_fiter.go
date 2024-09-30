package seq

//======控制========

// Filter 过滤元素,只保留满足条件的元素,即f() == true保留
func (t BiSeq[K, V]) Filter(f func(K, V) bool) BiSeq[K, V] {
    return func(c func(K, V)) {
        t(func(k K, v V) {
            if f(k, v) {
                c(k, v)
            }
        })
    }
}

// Take 保留前n个元素
func (t BiSeq[K, V]) Take(n int) BiSeq[K, V] {
    if n <= 0 {
        return func(k func(K, V)) {}
    }
    return func(c func(K, V)) {
        n := n
        t.Stoppable()(func(k K, v V) {
            c(k, v)
            n--
            if n == 0 {
                panic(&Stop)
            }
        })
    }
}

// Drop 跳过前n个元素
func (t BiSeq[K, V]) Drop(n int) BiSeq[K, V] {
    return func(c func(K, V)) {
        i := n
        t(func(k K, v V) {
            if i <= 0 {
                c(k, v)
            } else {
                i--
            }
        })
    }
}

// Distinct 去重
func (t BiSeq[K, V]) Distinct(equals func(K, V, K, V) bool) BiSeq[K, V] {
    var r []BiTuple[K, V]
    t(func(k K, v V) {
        for _, x := range r {
            if equals(k, v, x.K, x.V) {
                return
            }
        }
        r = append(r, BiTuple[K, V]{k, v})
    })
    return BiFrom(func(k func(K, V)) {
        for _, v := range r {
            k(v.K, v.V)
        }
    })
}

//// DistinctK 使用K去重
//func (t BiSeq[K, V]) DistinctK(equals func(K, K) bool) BiSeq[K, V] {
//    var r []BiTuple[K, V]
//    t(func(k K, v V) {
//        for _, x := range r {
//            if equals(k, x.K) {
//                return
//            }
//        }
//        r = append(r, BiTuple[K, V]{k, v})
//    })
//    return BiFrom(func(k func(K, V)) {
//        for _, v := range r {
//            k(v.K, v.V)
//        }
//    })
//}

//// DistinctV 使用V去重
//func (t BiSeq[K, V]) DistinctV(equals func(V, V) bool) BiSeq[K, V] {
//    var r []BiTuple[K, V]
//    t(func(k K, v V) {
//        for _, x := range r {
//            if equals(v, x.V) {
//                return
//            }
//        }
//        r = append(r, BiTuple[K, V]{k, v})
//    })
//    return BiFrom(func(k func(K, V)) {
//        for _, v := range r {
//            k(v.K, v.V)
//        }
//    })
//}
