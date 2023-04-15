package seq

//======控制========

// ConsumeTillStop 消费直到遇到stop
func (t BiSeq[K, V]) ConsumeTillStop(f func(K, V)) {
    defer func() {
        a := recover()
        if a != nil && a != &Stop {
            panic(a)
        }
    }()
    t(f)
}

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
    return func(c func(K, V)) {
        t.ConsumeTillStop(func(k K, v V) {
            n--
            if n >= 0 {
                c(k, v)
            } else {
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
