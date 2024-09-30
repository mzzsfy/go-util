package seq

import (
    "fmt"
    "github.com/mzzsfy/go-util/helper"
    "strconv"
)

type _stop bool

var (
    Stop *_stop
)

func getToStringFn[T any](i T) any {
    switch any(i).(type) {
    case string:
        return func(t string) string { return t }
    case bool:
        return func(t bool) string { return strconv.FormatBool(t) }
    case float64:
        return func(t float64) string { return strconv.FormatFloat(t, 'f', -1, 64) }
    case float32:
        return func(t float32) string { return strconv.FormatFloat(float64(t), 'f', -1, 32) }
    case int:
        return func(t int) string { return helper.NumberToString(t) }
    case int64:
        return func(t int64) string { return helper.NumberToString(t) }
    case int32:
        return func(t int32) string { return helper.NumberToString(t) }
    case int16:
        return func(t int16) string { return helper.NumberToString(t) }
    case int8:
        return func(t int8) string { return helper.NumberToString(t) }
    case uint:
        return func(t uint) string { return helper.NumberToString(t) }
    case uint64:
        return func(t uint64) string { return helper.NumberToString(t) }
    case uint32:
        return func(t uint32) string { return helper.NumberToString(t) }
    case uint16:
        return func(t uint16) string { return helper.NumberToString(t) }
    case uint8:
        return func(t uint8) string { return helper.NumberToString(t) }
    case []byte:
        return func(t []byte) string { return string(t) }
    case []rune:
        return func(t []rune) string { return string(t) }
    case fmt.Stringer:
        return func(t fmt.Stringer) string { return t.String() }
    case error:
        return func(t error) string { return t.Error() }
    default:
        return nil
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
