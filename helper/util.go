package helper

import (
    "sync"
    "time"
)

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

func OneOfL[L, R any](data L, _ R) L {
    return data
}
func OneOfR[L, R any](_ L, data R) R {
    return data
}
func OneOf3L[L, M, R any](data L, _ M, _ R) L {
    return data
}
func OneOf3M[L, M, R any](_ L, data M, _ R) M {
    return data
}
func OneOf3R[L, M, R any](_ L, _ M, data R) R {
    return data
}

func Debounce(call func(), duration time.Duration) func() {
    var lastCall *time.Time
    return func() {
        if lastCall == nil {
            call()
            t := time.Now()
            lastCall = &t
        } else {
            now := time.Now()
            if now.Sub(*lastCall) > duration {
                lastCall = &now
                call()
            }
        }
    }
}

func DebounceConcurrent(call func(), duration time.Duration) func() {
    f := Debounce(call, duration)
    var lock sync.Mutex
    return func() {
        lock.Lock()
        defer lock.Unlock()
        f()
    }
}
