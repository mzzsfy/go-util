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
        return func() int {
            start++
            return start
        }
    }
    if len(Range) == 2 {
        if Range[0] > Range[1] {
            start := Range[0]
            end := Range[1]
            return func() int {
                start--
                if start < end {
                    panic(&Stop)
                }
                return start
            }
        }
        start := Range[0]
        end := Range[1]
        return func() int {
            start++
            if start > end {
                panic(&Stop)
            }
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
    return func() int {
        if step > 0 {
            if start > end {
                panic(&Stop)
            }
        } else {
            if start < end {
                panic(&Stop)
            }
        }
        start += step
        return start - step
    }
}
func IteratorInt(Range ...int) Iterator[int] {
    f := makeRange(Range...)
    if len(Range) > 1 {
        return func() (i int, b bool) {
            defer func() {
                a := recover()
                if a != nil && a != &Stop {
                    panic(a)
                }
            }()
            return f(), true
        }
    } else {
        return func() (i int, b bool) {
            return f(), true
        }
    }
}
