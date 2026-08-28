package seq

//======控制========

// Filter 过滤元素,只保留满足条件的元素,即f(t) == true保留
func (t Seq[T]) Filter(f func(T) bool) Seq[T] {
	return func(c func(T)) {
		t(func(t T) {
			if f(t) {
				c(t)
			}
		})
	}
}

// Take 保留前n个元素
func (t Seq[T]) Take(n int) Seq[T] {
	if n <= 0 {
		return func(t func(T)) {}
	}
	return func(c func(T)) {
		t(func(e T) {
			if n <= 0 {
				panic(&stop)
			}
			c(e)
			n--
		})
	}
}

// TakeWhile 保留表达式返回true前的所有元素
func (t Seq[T]) TakeWhile(f func(T) bool) Seq[T] {
	return func(c func(T)) {
		t(func(e T) {
			if !f(e) {
				panic(&stop)
			}
			c(e)
		})
	}
}

// Limit 保留前n个元素,Take的别名
func (t Seq[T]) Limit(n int) Seq[T] {
	return t.Take(n)
}

// Drop 跳过前n个元素
func (t Seq[T]) Drop(n int) Seq[T] {
	return func(c func(T)) {
		t(func(e T) {
			if n <= 0 {
				c(e)
			} else {
				n--
			}
		})
	}
}

// DropWhile 保留表达式首次返回true后的所有元素
func (t Seq[T]) DropWhile(f func(T) bool) Seq[T] {
	return func(c func(T)) {
		ok := false
		t(func(e T) {
			if ok {
				c(e)
			} else {
				ok = f(e)
				if ok {
					c(e)
				}
			}
		})
	}
}

// Skip 跳过前n个元素,Drop的别名
func (t Seq[T]) Skip(n int) Seq[T] {
	return t.Drop(n)
}

// Distinct 去重,基于 == 语义O(n)去重
// T 为不可比较类型(含slice/map等字段)时会panic,该场景使用 DistinctByKey 或 DistinctCustomize
// 急切消费整个流做去重,无限流必须先 Take/Limit,否则会挂起
func (t Seq[T]) Distinct() Seq[T] {
	var r []T
	seen := make(map[any]struct{})
	t(func(t T) {
		k := any(t)
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			r = append(r, t)
		}
	})
	return FromSlice(r)
}

// DistinctCustomize 自定义去重
func (t Seq[T]) DistinctCustomize(contains func(T) bool) Seq[T] {
	return t.Filter(func(t T) bool { return !contains(t) })
}

// DistinctByKey 使用key函数提取键值,基于map进行O(n)去重
// 对于comparable类型,可使用 key=自身 作为键
func DistinctByKey[K comparable, T any](t Seq[T], key func(T) K) Seq[T] {
	return func(c func(T)) {
		seen := make(map[K]struct{})
		t(func(e T) {
			k := key(e)
			if _, ok := seen[k]; !ok {
				seen[k] = struct{}{}
				c(e)
			}
		})
	}
}

// DistinctComparable 对comparable类型使用map进行O(n)去重
func DistinctComparable[T comparable](t Seq[T]) Seq[T] {
	return DistinctByKey(t, func(e T) T { return e })
}
