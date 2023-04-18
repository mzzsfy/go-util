package seq

// Iterator 迭代器
type Iterator[T any] func() (T, bool)
type BiIterator[K, V any] func() (K, V, bool)

func makeRange(Range ...int) func() int {
    if len(Range) <= 1 {
        start := 0
        if len(Range) == 1 {
            start = Range[0]
        }
        start--
        return func() int {
            start++
            return start
        }
    }
    if len(Range) == 2 {
        if Range[0] > Range[1] {
            start := Range[0] + 1
            end := Range[1] + 1
            return func() int {
                if start < end {
                    panic(&Stop)
                }
                start--
                return start
            }
        }
        start := Range[0] - 1
        end := Range[1] - 1
        return func() int {
            if start > end {
                panic(&Stop)
            }
            start++
            return start
        }
    }
    start := Range[0]
    end := Range[1]
    step := Range[2]
    if step == 0 {
        panic("step can not be 0")
    }
    if step > 0 {
        if start > end {
            start = end
            end = Range[0]
        }
    } else {
        if start < end {
            start = end
            end = Range[0]
        }
    }
    if step > 0 {
        start -= step
        end -= step
        return func() int {
            if start > end {
                panic(&Stop)
            }
            start += step
            return start - step
        }
    }
    start += step
    end += step
    return func() int {
        if start < end {
            panic(&Stop)
        }
        start += step
        return start - step
    }
}

// IteratorInt 生成整数迭代器,可以自定义起始值,结束值,步长
// 参数1,起始值,默认为0 (包括)
// 参数2,结束值 (包括),比起始值小递减,比起始值大递增
// 参数3,步长,默认为1,为正递增,为负递减,不关心起始值和结束值的大小
func IteratorInt(Range ...int) Iterator[int] {
    if len(Range) > 1 {
        if len(Range) == 2 {
            if Range[0] > Range[1] {
                start := Range[0] + 1
                end := Range[1] + 1
                return func() (int, bool) {
                    if start < end {
                        return 0, false
                    }
                    start--
                    return start, true
                }
            }
            start := Range[0] - 1
            end := Range[1] - 1
            return func() (int, bool) {
                start++
                if start > end {
                    return 0, false
                }
                return start, true
            }
        }
        start := Range[0]
        end := Range[1]
        step := Range[2]
        if step == 0 {
            panic("step can not be 0")
        }
        if step > 0 {
            if start > end {
                start = end
                end = Range[0]
            }
        } else {
            if start < end {
                start = end
                end = Range[0]
            }
        }
        if step > 0 {
            start -= step
            end -= step
            return func() (int, bool) {
                if start > end {
                    return 0, false
                }
                start += step
                return start - step, true
            }
        }
        start += step
        end += step
        return func() (int, bool) {
            if start < end {
                return 0, false
            }
            start += step
            return start, true
        }
    } else {
        f := makeRange(Range...)
        return func() (i int, b bool) {
            return f(), true
        }
    }
}
