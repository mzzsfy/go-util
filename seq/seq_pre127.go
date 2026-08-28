//go:build !go1.27

package seq

import (
	"runtime"
)

// 本文件收纳 Seq 域在 go1.27 及以下版本使用的 API,这些功能在 go1.27 轨道由泛型方法提供(见 seq_method_go127.go)

//======生成相关========

// CastAny 从any类型的Seq转换为T类型的Seq,强制转换
func CastAny[T any](seq Seq[any]) Seq[T] {
	return func(c func(T)) { seq(func(t any) { c((t).(T)) }) }
}

// CastAnyT 从any类型的Seq转换为T类型的Seq,强制转换,简便写法
func CastAnyT[T any](seq Seq[any], _ T) Seq[T] {
	return func(c func(T)) { seq(func(t any) { c((t).(T)) }) }
}

// Map 每个元素自定义转换
func Map[E, T any](seq Seq[T], cast func(T) E) Seq[E] {
	return func(c func(E)) { seq(func(t T) { c(cast(t)) }) }
}

// Join 合并多个Seq
func Join[T any](seqs ...Seq[T]) Seq[T] {
	return func(c func(T)) {
		defer stopRecover()
		for _, seq := range seqs {
			seq(func(t T) { c(t) })
		}
	}
}

// JoinL 合并2个不同Seq,右边转换为左边的类型
func JoinL[T, E any](seq1 Seq[T], seq2 Seq[E], cast func(E) T) Seq[T] {
	return func(c func(T)) {
		seq1(func(t T) { c(t) })
		seq2(func(t E) { c(cast(t)) })
	}
}

// JoinBy 合并2个不同Seq,统一转换为新类型
func JoinBy[T, E, R any](seq1 Seq[T], cast1 func(T) R, seq2 Seq[E], cast2 func(E) R) Seq[R] {
	return func(c func(R)) {
		seq1(func(t T) { c(cast1(t)) })
		seq2(func(t E) { c(cast2(t)) })
	}
}

//======转换相关========

// MapParallel 每个元素转换为any
// order.1 顺序保证方式,规则如下:
// 0:不保证任务启动顺序,不保证消费顺序,会消费竞争
// 1:尽量保持顺序,优先保证并发数,异步任务完成时,会直接消费,会消费竞争,可以使用 Sync() 保证消费不竞争,其后的Last/OnLast等单次消费操作仍是数据竞争
// 2:异步任务与消费端解偶,在保证顺序的前提下,优先保证并发数,不会消费竞争
// 3:保持异步与消费同步,以消费为准,不消费完成不会开始下一个异步任务,不会消费竞争
//
// order.2 最大并发数,根据第一个参数决定逻辑,缺省为2*GOMAXPROCS
func (t Seq[T]) MapParallel(syncFn func(T) any, order ...int) Seq[any] {
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
		return mapParallelCore(t, syncFn, o, sl)
	} else {
		return t.Parallel(sl).Map(syncFn)
	}
}

func (t Seq[T]) MapParallelCustomize(asyncFn func(T, func(any))) Seq[any] {
	return mapParallelCustomizeCore(t, asyncFn)
}

// Map 每个元素转换为any
func (t Seq[T]) Map(f func(T) any) Seq[any] {
	return func(c func(any)) { t(func(t T) { c(f(t)) }) }
}

// MapString 每个元素转换为 string
func (t Seq[T]) MapString(f func(T) string) Seq[string] {
	return func(c func(string)) { t(func(t T) { c(f(t)) }) }
}

// MapInt 每个元素转换为 int
func (t Seq[T]) MapInt(f func(T) int) Seq[int] {
	return func(c func(int)) { t(func(t T) { c(f(t)) }) }
}

// MapFlat 每个元素转换为Seq,并扁平化
func (t Seq[T]) MapFlat(f func(T) Seq[any]) Seq[any] {
	return func(c func(any)) { t(func(t T) { f(t)(c) }) }
}

// 扁平化特化,泛型方法场景下使用 MapFlat 与显式类型参数替代
// func MapFlatInt 扁平化 → Seq.MapFlat[int]
func MapFlatInt[T any](t Seq[T], f func(T) Seq[int]) Seq[int] {
	return func(c func(int)) { t(func(t T) { f(t)(c) }) }
}

func MapFlatString[T any](t Seq[T], f func(T) Seq[string]) Seq[string] {
	return func(c func(string)) { t(func(t T) { f(t)(c) }) }
}

//======消费相关========

// GroupBy 元素分组,每个组保留所有元素
func (t Seq[T]) GroupBy(f func(T) any) map[any][]T {
	r := make(map[any][]T)
	t(func(t T) {
		k := f(t)
		r[k] = append(r[k], t)
	})
	return r
}

// GroupByFirst 元素分组,每个组只保留第一个元素
func (t Seq[T]) GroupByFirst(f func(T) any) map[any]T {
	r := make(map[any]T)
	t(func(t T) {
		k := f(t)
		if _, ok := r[k]; !ok {
			r[k] = t
		}
	})
	return r
}

// GroupByLast 元素分组,每个组只保留最后一个元素
func (t Seq[T]) GroupByLast(f func(T) any) map[any]T {
	r := make(map[any]T)
	t(func(t T) {
		k := f(t)
		r[k] = t
	})
	return r
}

// Reduce 自定义聚合
func (t Seq[T]) Reduce(f func(T, any) any, init any) any {
	t(func(t T) { init = f(t, init) })
	return init
}
