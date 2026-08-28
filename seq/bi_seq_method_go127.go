//go:build go1.27

package seq

import (
	"runtime"
)

// 本文件在 go1.27 轨道提供 BiSeq 泛型方法形态的强类型 API,替代 go1.27 以下版本的 any 妥协函数(见 *_pre127.go)

//======转换相关========

// MapVParallel 每个元素转换V为R
// order 语义见 Seq.MapParallel
func (t BiSeq[K, V]) MapVParallel[R any](f func(k K, v V) R, order ...int) BiSeq[K, R] {
	o := 0
	sl := 0
	if len(order) > 0 {
		o = order[0]
	}
	if len(order) > 1 {
		sl = order[1]
	}
	if sl <= 0 {
		//缺省并发数与Parallel一致
		sl = runtime.GOMAXPROCS(0) * 2
	}
	if o > 0 {
		return biMapVParallelCore(t, f, o, sl)
	}

	return biMapVCore(t.Parallel(sl), f)
}

// Map 每个元素自定义转换为新的K,V,用于连续强类型转换
func (t BiSeq[K, V]) Map[A, B any](f func(K, V) (A, B)) BiSeq[A, B] {
	return func(c func(A, B)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// MapK 每个元素自定义转换K为R,V保持不变
func (t BiSeq[K, V]) MapK[R any](f func(K, V) R) BiSeq[R, V] {
	return func(c func(R, V)) { t(func(k K, v V) { c(f(k, v), v) }) }
}

// MapV 每个元素自定义转换V为R,K保持不变
func (t BiSeq[K, V]) MapV[R any](f func(K, V) R) BiSeq[K, R] {
	return biMapVCore(t, f)
}

// MapFlat 每个元素转换为BiSeq[A,B],并扁平化
func (t BiSeq[K, V]) MapFlat[A, B any](f func(K, V) BiSeq[A, B]) BiSeq[A, B] {
	return func(c func(A, B)) { t(func(k K, v V) { f(k, v)(c) }) }
}

// MapFlatK K扁平化,每个元素展开为多个新K,V保持不变
func (t BiSeq[K, V]) MapFlatK[R any](f func(K, V) Seq[R]) BiSeq[R, V] {
	return func(c func(R, V)) {
		t(func(k K, v V) {
			s := f(k, v)
			s.ForEach(func(a R) {
				c(a, v)
			})
		})
	}
}

// MapFlatV V扁平化,每个元素展开为多个新V,K保持不变
func (t BiSeq[K, V]) MapFlatV[R any](f func(K, V) Seq[R]) BiSeq[K, R] {
	return func(c func(K, R)) {
		t(func(k K, v V) {
			s := f(k, v)
			s.ForEach(func(a R) {
				c(k, a)
			})
		})
	}
}

// MapFlatSingle 扁平化为 Seq[E]
func (t BiSeq[K, V]) MapFlatSingle[E any](f func(K, V) Seq[E]) Seq[E] {
	return func(c func(E)) {
		t(func(k K, v V) {
			s := f(k, v)
			s.ForEach(func(a E) {
				c(a)
			})
		})
	}
}

// JoinL 合并2个不同Seq,右边转换为左边的类型
func (t BiSeq[K, V]) JoinL[K1, V1 any](seq2 BiSeq[K1, V1], cast func(K1, V1) (K, V)) BiSeq[K, V] {
	return func(c func(K, V)) {
		defer stopRecover()
		t(func(k K, v V) { c(k, v) })
		seq2(func(k K1, v V1) { c(cast(k, v)) })
	}
}

// JoinBy 合并2个不同Seq,统一转换为新类型
func (t BiSeq[K, V]) JoinBy[A, B, K2, V2 any](seq2 BiSeq[K2, V2], cast1 func(K, V) (A, B), cast2 func(K2, V2) (A, B)) BiSeq[A, B] {
	return func(c func(A, B)) {
		defer stopRecover()
		t(func(k K, v V) { c(cast1(k, v)) })
		seq2(func(k K2, v V2) { c(cast2(k, v)) })
	}
}

// Cast 强制断言每个元素为NK,NV类型,用于恢复经过any化的强类型链
func (t BiSeq[K, V]) Cast[NK, NV any]() BiSeq[NK, NV] {
	return func(c func(NK, NV)) { t(func(k K, v V) { c(any(k).(NK), any(v).(NV)) }) }
}

//======消费相关========

// Reduce 自定义聚合,R 由 init 的类型确定
func (t BiSeq[K, V]) Reduce[R any](f func(K, V, R) R, init R) R {
	r := init
	t(func(k K, v V) { r = f(k, v, r) })
	return r
}
