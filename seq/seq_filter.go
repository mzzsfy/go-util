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
    return func(c func(T)) {
        n := n
        t.Stoppable()(func(e T) {
            n--
            if n >= 0 {
                c(e)
            } else {
                panic(&Stop)
            }
        })
    }
}

// TakeWhile 保留表达式返回true前的所有元素
func (t Seq[T]) TakeWhile(f func(T) bool) Seq[T] {
    return func(c func(T)) {
        t(func(e T) {
            if f(e) {
                panic(&Stop)
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
        n := n
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

// Distinct 去重
func (t Seq[T]) Distinct(equals func(T, T) bool) Seq[T] {
    var r []T
    t(func(t T) {
        //如何优化?
        for _, v := range r {
            if equals(t, v) {
                return
            }
        }
        r = append(r, t)
    })
    return FromSlice(r)
}
