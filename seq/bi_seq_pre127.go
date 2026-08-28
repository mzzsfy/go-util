//go:build !go1.27

package seq

import (
	"runtime"
)

// 本文件收纳 BiSeq 域在 go1.27 及以下版本使用的 API,这些功能在 go1.27 轨道由泛型方法提供(见 bi_seq_method_go127.go)

//======静态转换方法,xxxK 表示操作左侧参数========

// BiCastAny 从BiSeq[any,any]强制转换为BiSeq[K,V]
func BiCastAny[K, V any](seq BiSeq[any, any]) BiSeq[K, V] {
	return func(c func(K, V)) { seq(func(k, v any) { c(k.(K), v.(V)) }) }
}

// BiCastAnyK 从BiSeq[any,V]强制转换K为指定类型
func BiCastAnyK[K, V any](seq BiSeq[any, V]) BiSeq[K, V] {
	return func(c func(K, V)) { seq(func(k any, v V) { c(k.(K), v) }) }
}

// BiCastAnyV 从BiSeq[K,any]强制转换V为指定类型
func BiCastAnyV[V, K any](seq BiSeq[K, any]) BiSeq[K, V] {
	return func(c func(K, V)) { seq(func(k K, v any) { c(k, v.(V)) }) }
}

// BiCastAnyT 从BiSeq[any,any]强制转换为BiSeq[K,V],简便写法
func BiCastAnyT[K, V any](seq BiSeq[any, any], _ K, _ V) BiSeq[K, V] {
	return func(c func(K, V)) { seq(func(k, v any) { c(k.(K), v.(V)) }) }
}

// BiCastAnyVT 从BiSeq[K,any]强制转换V为指定类型,简便写法
func BiCastAnyVT[V, K any](seq BiSeq[K, any], _ V) BiSeq[K, V] {
	return func(c func(K, V)) { seq(func(k K, v any) { c(k, v.(V)) }) }
}

// BiCastAnyKT 从BiSeq[any,V]强制转换K为指定类型,简便写法
func BiCastAnyKT[K, V any](seq BiSeq[any, V], _ K) BiSeq[K, V] {
	return func(c func(K, V)) { seq(func(k any, v V) { c(k.(K), v) }) }
}

// BiJoin 合并多个Seq
func BiJoin[K, V any](seqs ...BiSeq[K, V]) BiSeq[K, V] {
	return func(c func(K, V)) {
		defer stopRecover()
		for _, seq := range seqs {
			seq(func(k K, v V) { c(k, v) })
		}
	}
}

// BiJoinL 合并2个不同Seq,右边转换为左边的类型
func BiJoinL[K, V, K1, V1 any](seq1 BiSeq[K, V], seq2 BiSeq[K1, V1], cast func(K1, V1) (K, V)) BiSeq[K, V] {
	return func(c func(K, V)) {
		seq1(func(k K, v V) { c(k, v) })
		seq2(func(k K1, v V1) { c(cast(k, v)) })
	}
}

// BiJoinBy 合并2个不同Seq,统一转换为新类型
func BiJoinBy[K1, V1, K2, V2, K, V any](seq1 BiSeq[K1, V1], cast1 func(K1, V1) (K, V), seq2 BiSeq[K2, V2], cast2 func(K2, V2) (K, V)) BiSeq[K, V] {
	return func(c func(K, V)) {
		seq1(func(k K1, v V1) { c(cast1(k, v)) })
		seq2(func(k K2, v V2) { c(cast2(k, v)) })
	}
}

//======转换相关========

// MapVParallel 每个元素转换为any
// order 语义见 MapParallel
func (t BiSeq[K, V]) MapVParallel(f func(k K, v V) any, order ...int) BiSeq[K, any] {
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
	} else {
		return biMapVCore(t.Parallel(sl), f)
	}
}

// Map 每个元素自定义转换为any,用于连续转换操作,使用 Cast() 进行恢复泛型
func (t BiSeq[K, V]) Map(f func(K, V) (any, any)) BiSeq[any, any] {
	return func(c func(any, any)) { t(func(k K, v V) { c(f(k, v)) }) }
}

// MapV 每个元素自定义转换V为any,用于连续转换操作,使用 Cast() 进行恢复泛型
func (t BiSeq[K, V]) MapV(f func(K, V) any) BiSeq[K, any] {
	return func(c func(K, any)) { t(func(k K, v V) { c(k, f(k, v)) }) }
}

// MapFlat 每个元素转换为BiSeq[any,any],并扁平化
func (t BiSeq[K, V]) MapFlat(f func(K, V) BiSeq[any, any]) BiSeq[any, any] {
	return func(c func(any, any)) { t(func(k K, v V) { f(k, v)(c) }) }
}

// JoinBy 合并Seq,右侧转换为当前类型
func (t BiSeq[K, V]) JoinBy(seq BiSeq[any, any], cast func(any, any) (K, V)) BiSeq[K, V] {
	return func(c func(K, V)) {
		t(func(k K, v V) { c(k, v) })
		seq(func(k any, v any) { c(cast(k, v)) })
	}
}

// BiMap 从BiSeq[K,V]自定义转换为BiSeq[RK,RV]
func BiMap[K, V, RK, RV any](seq BiSeq[K, V], cast func(K, V) (RK, RV)) BiSeq[RK, RV] {
	return func(c func(RK, RV)) { seq(func(k K, v V) { c(cast(k, v)) }) }
}

// BiMapK 从BiSeq[K,V]自定义转换K
func BiMapK[K, V, RK any](seq BiSeq[K, V], cast func(K, V) RK) BiSeq[RK, V] {
	return func(c func(RK, V)) { seq(func(k K, v V) { c(cast(k, v), v) }) }
}

// BiMapV 从BiSeq[K,V]自定义转换V
func BiMapV[V, RV, K any](seq BiSeq[K, V], cast func(K, V) RV) BiSeq[K, RV] {
	return func(c func(K, RV)) { seq(func(k K, v V) { c(k, cast(k, v)) }) }
}

func BiMapExchangeKV[K, V any](f BiSeq[K, V]) BiSeq[V, K] {
	return func(t func(V, K)) { f(func(k K, v V) { t(v, k) }) }
}

func BiMapFlatK[K, V any](t BiSeq[K, V], f func(K, V) Seq[any]) BiSeq[any, V] {
	return func(c func(any, V)) {
		t(func(k K, v V) {
			s := f(k, v)
			s.ForEach(func(a any) {
				c(a, v)
			})
		})
	}
}

func BiMapFlatV[K, V any](t BiSeq[K, V], f func(K, V) Seq[any]) BiSeq[K, any] {
	return func(c func(K, any)) {
		t(func(k K, v V) {
			s := f(k, v)
			s.ForEach(func(a any) {
				c(k, a)
			})
		})
	}
}

// BiMapFlatSingle 扁平化为 Seq[T]
func BiMapFlatSingle[K, V any](t BiSeq[K, V], f func(K, V) Seq[any]) Seq[any] {
	return func(c func(any)) {
		t(func(k K, v V) {
			s := f(k, v)
			s.ForEach(func(a any) {
				c(a)
			})
		})
	}
}

//======消费相关========

// Reduce 求值
func (t BiSeq[K, V]) Reduce(f func(K, V, any) any, init any) any {
	t(func(k K, v V) { init = f(k, v, init) })
	return init
}
