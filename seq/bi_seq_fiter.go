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
		t(func(k K, v V) {
			if n <= 0 {
				panic(&stop)
			}
			c(k, v)
			n--
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

// BiDistinctByKey 使用key函数提取键值,基于map进行O(n)去重,与Seq.DistinctByKey对齐
// K,V 为不可比较类型时,可通过key函数提取comparable的键完成去重
func BiDistinctByKey[K, V any, CK comparable](t BiSeq[K, V], key func(K, V) CK) BiSeq[K, V] {
	return func(c func(K, V)) {
		seen := make(map[CK]struct{})
		t(func(k K, v V) {
			ck := key(k, v)
			if _, ok := seen[ck]; !ok {
				seen[ck] = struct{}{}
				c(k, v)
			}
		})
	}
}

// distinctKey BiSeq去重的组合键,K,V经any装箱后依赖运行时可比较
type distinctKey struct {
	k any
	v any
}

// Distinct 去重,基于K,V的 == 语义O(n)去重
// K,V 为不可比较类型(含slice/map等字段)时会panic
// 急切消费整个流做去重,无限流必须先 Take/Limit,否则会挂起
func (t BiSeq[K, V]) Distinct() BiSeq[K, V] {
	var r []BiTuple[K, V]
	seen := make(map[distinctKey]struct{})
	t(func(k K, v V) {
		key := distinctKey{any(k), any(v)}
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			r = append(r, BiTuple[K, V]{k, v})
		}
	})
	return BiFrom(func(k func(K, V)) {
		for _, v := range r {
			k(v.K, v.V)
		}
	})
}
