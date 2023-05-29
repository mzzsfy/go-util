package util

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
func MaxA[N Number](ns ...N) N {
    if len(ns) == 0 {
        panic("Max中需要有一个值")
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
func MinA[N Number](ns ...N) N {
    if len(ns) == 0 {
        panic("Min中需要有一个值")
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
        return -n
    }
    return n
}
