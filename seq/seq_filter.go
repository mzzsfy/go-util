package seq

//======控制,删除或者停止========

// ConsumeTillStop 消费直到遇到stop
func (t Seq[T]) ConsumeTillStop(f func(T)) {
    defer func() {
        a := recover()
        if a != nil && a != &Stop {
            panic(a)
        }
    }()
    t(f)
}

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
        t.ConsumeTillStop(func(e T) {
            n--
            if n >= 0 {
                c(e)
            } else {
                panic(&Stop)
            }
        })
    }
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
