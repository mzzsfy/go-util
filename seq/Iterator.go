package seq

import (
    "fmt"
    "math"
)

// Iterator 迭代器
type Iterator[T any] func() (T, bool)
type BiIterator[K, V any] func() (K, V, bool)

func IteratorInt(rang ...int) Iterator[int] {
    start := 0
    end := math.MaxInt
    step := 1
    if len(rang) > 0 {
        start = rang[0]
    }
    if len(rang) > 1 {
        end = rang[1]
    }
    if len(rang) > 2 {
        step = rang[2]
        if step == 0 {
            panic("step can not be 0")
        }
    }
    if step > 0 {
        if start > end {
            panic(fmt.Sprintf("step is %d ,start(%d) can not be greater than end(%d)", step, start, end))
        }
    } else {
        if start < end {
            panic(fmt.Sprintf("step is %d ,start(%d) can not be less than end(%d)", step, start, end))
        }
    }
    return func() (int, bool) {
        if step > 0 {
            if start > end {
                return 0, false
            }
        } else {
            if start < end {
                return 0, false
            }
        }
        start += step
        return start - step, true
    }
}
