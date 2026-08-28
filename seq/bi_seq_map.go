package seq

//======转换,添加或修改内部元素========

//// ExchangeKV 交换kv位置,转换为 BiSeq[V, K]
//func (t BiSeq[K, V]) ExchangeKV() BiSeq[V, K] {
//    return func(c func(V, K)) { t(func(k K, v V) { c(v, k) }) }
//}

// Join 合并多个Seq
func (t BiSeq[K, V]) Join(seqs ...BiSeq[K, V]) BiSeq[K, V] {
	return func(c func(K, V)) {
		defer stopRecover()
		t(func(k K, v V) { c(k, v) })
		for _, seq := range seqs {
			seq(func(k K, v V) { c(k, v) })
		}
	}
}

// Add 添加元素
func (t BiSeq[K, V]) Add(k K, v V) BiSeq[K, V] {
	return func(c func(K, V)) {
		t(func(k K, v V) { c(k, v) })
		c(k, v)
	}
}

//// AddTuple 添加元素
//func (t BiSeq[K, V]) AddTuple(vs ...BiTuple[K, V]) BiSeq[K, V] {
//    return func(c func(K, V)) {
//        t(func(k K, v V) { c(k, v) })
//        for _, v := range vs {
//            c(v.K, v.V)
//        }
//    }
//}
//
//// AddBy 添加元素
//func (t BiSeq[K, V]) AddBy(cast func(any, any) (K, V), es ...any) BiSeq[K, V] {
//    if len(es)%2 != 0 {
//        panic("添加的元素个数必须为偶数")
//    }
//    return func(c func(K, V)) {
//        t(func(k K, v V) { c(k, v) })
//        FromIntSeq(0, len(es), 2)(func(i int) { c(cast(es[i], es[i+1])) })
//    }
//}

func (t BiSeq[K, V]) AddIf(condition bool, k K, v V) BiSeq[K, V] {
	if !condition {
		return t
	}
	return t.Add(k, v)
}

func (t BiSeq[K, V]) AddIfF(condition func(BiSeq[K, V]) bool, k K, v V) BiSeq[K, V] {
	if !condition(t) {
		return t
	}
	return t.Add(k, v)
}

//func (t BiSeq[K, V]) AddTupleIf(condition bool, vs ...BiTuple[K, V]) BiSeq[K, V] {
//    if !condition {
//        return t
//    }
//    return t.AddTuple(vs...)
//}
//
//func (t BiSeq[K, V]) AddTupleIfF(condition func(BiSeq[K, V]) bool, vs ...BiTuple[K, V]) BiSeq[K, V] {
//    if !condition(t) {
//        return t
//    }
//    return t.AddTuple(vs...)
//}

//func (t BiSeq[K, V]) AddByIf(condition bool, cast func(any, any) (K, V), es ...any) BiSeq[K, V] {
//    if !condition {
//        return t
//    }
//    return t.AddBy(cast, es...)
//}
//
//func (t BiSeq[K, V]) AddByIfF(condition func(BiSeq[K, V]) bool, cast func(any, any) (K, V), es ...any) BiSeq[K, V] {
//    if !condition(t) {
//        return t
//    }
//    return t.AddBy(cast, es...)
//}
