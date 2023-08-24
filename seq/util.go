package seq

import (
    "fmt"
    "strconv"
)

type _stop bool

var (
    Stop *_stop
)

func getToStringFn[T any](i T) func(T) string {
    switch r := any(i).(type) {
    case string:
        return func(t T) string {
            return r
        }
    case bool:
        return func(t T) string {
            return strconv.FormatBool(r)
        }
    case float64:
        return func(t T) string {
            return strconv.FormatFloat(r, 'f', -1, 64)
        }
    case float32:
        return func(t T) string {
            return strconv.FormatFloat(float64(r), 'f', -1, 64)
        }
    case int:
        return func(t T) string {
            return strconv.Itoa(r)
        }
    case int64:
        return func(t T) string {
            return strconv.FormatInt(r, 10)
        }
    case int32:
        return func(t T) string {
            return strconv.Itoa(int(r))
        }
    case int16:
        return func(t T) string {
            return strconv.Itoa(int(r))
        }
    case int8:
        return func(t T) string {
            return strconv.Itoa(int(r))
        }
    case uint:
        return func(t T) string {
            return strconv.FormatUint(uint64(r), 10)
        }
    case uint64:
        return func(t T) string {
            return strconv.FormatUint(r, 10)
        }
    case uint32:
        return func(t T) string {
            return strconv.FormatUint(uint64(r), 10)
        }
    case uint16:
        return func(t T) string {
            return strconv.FormatUint(uint64(r), 10)
        }
    case uint8:
        return func(t T) string {
            return strconv.FormatUint(uint64(r), 10)
        }
    case []byte:
        return func(t T) string {
            return string(r)
        }
    case []rune:
        return func(t T) string {
            return string(r)
        }
    case fmt.Stringer:
        return func(t T) string {
            return r.String()
        }
    case error:
        return func(t T) string {
            return r.Error()
        }
    default:
        return func(t T) string {
            return fmt.Sprint(t)
        }
    }
}

type Comparable interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}

// EqualsT 是否相等,用法: .Distinct(EqualsT[int])
func EqualsT[T comparable](i T, i2 T) bool {
    return i == i2
}

// LessT 排序用,小的在前,用法: .Order(LessT[int])
func LessT[T Comparable](i T, i2 T) bool {
    return i < i2
}

// GreatT 排序用,大的在前,用法: .Order(GreatT[int])
func GreatT[T Comparable](i int, i2 int) bool {
    return i > i2
}

func AnyT[T any](t T) any {
    return any(t)
}
func AnyBiT[K, V any](k K, v V) (any, any) {
    return any(k), any(v)
}
func AnyBiTK[T any](t T, a any) (any, any) {
    return any(t), a
}
func AnyBiTV[T any](a any, t T) (any, any) {
    return a, any(t)
}
