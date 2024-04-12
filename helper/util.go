package helper

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

func Default[T any](test, defaultValue T) T {
    if !IsZero(test) {
        return test
    }
    return defaultValue
}

func Defaults[T any](defaultValue T, tests ...T) T {
    for _, t := range tests {
        if !IsZero(t) {
            return t
        }
    }
    return defaultValue
}

func NotZero(test any) bool {
    return !IsZero(test)
}

func AnyArray(vars ...any) []any {
    return vars
}

func AnyArrayT[T any](vars ...T) []T {
    return vars
}

func Must[T any](data T, err error) T {
    if err != nil {
        panic(err)
    }
    return data
}

func MustR[T any](err error, data T) T {
    if err != nil {
        panic(err)
    }
    return data
}
