package seq

import (
	"fmt"
	"github.com/mzzsfy/go-util/helper"
	"strconv"
)

type _stop bool

var stop *_stop

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

// LessT 排序用,小的在前,用法: .Order(LessT[int])
func LessT[T Comparable](i T, i2 T) bool {
	return i < i2
}

// GreatT 排序用,大的在前,用法: .Order(GreatT[int])
func GreatT[T Comparable](i T, i2 T) bool {
	return i > i2
}

// sortSlice 泛型排序,三数取中快排,小区间转插入排序,非稳定,与sort.Slice语义一致
func sortSlice[T any](r []T, less func(T, T) bool) {
	//已排序检测,近有序数据O(n)返回
	sorted := true
	for i := 1; i < len(r); i++ {
		if less(r[i], r[i-1]) {
			sorted = false
			break
		}
	}
	if sorted {
		return
	}
	quickSort(r, less)
}

// insertionSortMaxSize 转入插入排序的小区间上限
const insertionSortMaxSize = 12

// quickSort 较小区间转插入排序,递归较小侧控制栈深
func quickSort[T any](r []T, less func(T, T) bool) {
	for len(r) > insertionSortMaxSize {
		//三数取中,中位数移至首位作基准
		mid := len(r) / 2
		last := len(r) - 1
		if less(r[mid], r[0]) {
			r[mid], r[0] = r[0], r[mid]
		}
		if less(r[last], r[0]) {
			r[last], r[0] = r[0], r[last]
		}
		if less(r[last], r[mid]) {
			r[last], r[mid] = r[mid], r[last]
		}
		r[mid], r[0] = r[0], r[mid]
		pivot := r[0]
		i, j := 1, last
		for {
			for i <= j && less(r[i], pivot) {
				i++
			}
			for i <= j && less(pivot, r[j]) {
				j--
			}
			if i > j {
				break
			}
			r[i], r[j] = r[j], r[i]
			i++
			j--
		}
		r[0], r[j] = r[j], r[0]
		if j < len(r)-j-1 {
			quickSort(r[:j], less)
			r = r[j+1:]
		} else {
			quickSort(r[j+1:], less)
			r = r[:j]
		}
	}
	insertionSort(r, less)
}

func insertionSort[T any](r []T, less func(T, T) bool) {
	for i := 1; i < len(r); i++ {
		for j := i; j > 0 && less(r[j], r[j-1]); j-- {
			r[j], r[j-1] = r[j-1], r[j]
		}
	}
}
