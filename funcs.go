package util

func Ternary[T any](test bool, trueValue, falseValue T) T {
    if test {
        return trueValue
    }
    return falseValue
}

func TernaryF[T any, F func() T](test bool, trueValue, falseValue F) T {
    if test {
        return trueValue()
    }
    return falseValue()
}

func TernaryVF[T any, F func() T](test bool, trueValue T, falseValue F) T {
    if test {
        return trueValue
    }
    return falseValue()
}

func DoMapIf[T any](t T, test func(T) bool, f func(T) T) T {
    if test(t) {
        return f(t)
    }
    return t
}

func DoMapIfZero[T any](t T, f func() T) T {
    if !IsZero(t) {
        return t
    }
    return f()
}

func DoMapIfNoZero[T any](t T, f func(T) T) T {
    if IsZero(t) {
        return t
    }
    return f(t)
}
