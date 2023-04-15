package seq

import (
    "fmt"
    "strconv"
)

func getToStringFn[T any](i T) func(T) string {
    switch any(i).(type) {
    case string:
        return func(t T) string {
            return any(t).(string)
        }
    case bool:
        return func(t T) string {
            return strconv.FormatBool(any(t).(bool))
        }
    case float64:
        return func(t T) string {
            return strconv.FormatFloat(any(t).(float64), 'f', -1, 64)
        }
    case float32:
        return func(t T) string {
            return strconv.FormatFloat(float64(any(t).(float32)), 'f', -1, 64)
        }
    case int:
        return func(t T) string {
            return strconv.Itoa(any(t).(int))
        }
    case int64:
        return func(t T) string {
            return strconv.FormatInt(any(t).(int64), 10)
        }
    case int32:
        return func(t T) string {
            return strconv.Itoa(int(any(t).(int32)))
        }
    case int16:
        return func(t T) string {
            return strconv.Itoa(int(any(t).(int16)))
        }
    case int8:
        return func(t T) string {
            return strconv.Itoa(int(any(t).(int8)))
        }
    case uint:
        return func(t T) string {
            return strconv.FormatUint(uint64(any(t).(uint)), 10)
        }
    case uint64:
        return func(t T) string {
            return strconv.FormatUint(any(t).(uint64), 10)
        }
    case uint32:
        return func(t T) string {
            return strconv.FormatUint(uint64(any(t).(uint32)), 10)
        }
    case uint16:
        return func(t T) string {
            return strconv.FormatUint(uint64(any(t).(uint16)), 10)
        }
    case uint8:
        return func(t T) string {
            return strconv.FormatUint(uint64(any(t).(uint8)), 10)
        }
    case []byte:
        return func(t T) string {
            return string(any(t).([]byte))
        }
    case []rune:
        return func(t T) string {
            return string(any(t).([]rune))
        }
    case fmt.Stringer:
        return func(t T) string {
            return any(t).(fmt.Stringer).String()
        }
    case error:
        return func(t T) string {
            return any(t).(error).Error()
        }
    default:
        return func(t T) string {
            return fmt.Sprint(t)
        }
    }
}

type stop bool

var (
    Stop *stop
)

type Comparable interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
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
func AnyBiTR[T any](t T, a any) (any, any) {
    return any(t), a
}
func AnyBiTL[T any](a any, t T) (any, any) {
    return a, any(t)
}
