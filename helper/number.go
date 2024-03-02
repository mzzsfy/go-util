package helper

import "sync"

type Signed interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}
type Unsigned interface {
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}
type Integer interface {
    Signed | Unsigned
}
type Float interface {
    ~float32 | ~float64
}
type Complex interface {
    ~complex64 | ~complex128
}
type Number interface {
    Integer | Float
}

type Ordered interface {
    Number | ~string
}

func Max[N Number](n1, n2 N) N {
    if n1 < n2 {
        return n2
    }
    return n1
}
func MaxN[N Number](ns ...N) N {
    if len(ns) == 0 {
        panic("Max中最少提供一个值")
    }
    r := ns[0]
    for _, n := range ns {
        if r < n {
            r = n
        }
    }
    return r
}

func Min[N Number](n1, n2 N) N {
    if n1 > n2 {
        return n2
    }
    return n1
}
func MinN[N Number](ns ...N) N {
    if len(ns) == 0 {
        panic("Min中最少提供一个值")
    }
    r := ns[0]
    for _, n := range ns {
        if r < n {
            r = n
        }
    }
    return r
}

func Abs[N Number](n N) N {
    if n < 0 {
        n = -n
        if n < 0 {
            n += 1
            n = -n
        }
    }
    return n
}

var buf = sync.Pool{
    New: func() any {
        return &[20]byte{}
    },
}

func NumberToString[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) string {
    bs := buf.Get().(*[20]byte)
    defer buf.Put(bs)
    r := bs[:]
    negative := false
    i := len(r) - 1
    if n < 0 {
        n = -n
        //溢出,特殊处理
        // math.MinInt8
        // math.MinInt16
        // math.MinInt32
        // math.MinInt64
        if n < 0 {
            next := n / 10
            r[i] = byte('0' + -(n - next*10))
            n = -next
            i--
        }
        negative = true
    }
    for n > 9 {
        next := n / 10
        r[i] = byte('0' + n - next*10)
        n = next
        i--
    }
    r[i] = byte('0' + n)
    if negative {
        i--
        r[i] = '-'
    }
    return string(r[i:])
}
