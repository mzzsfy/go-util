package helper

import (
	"reflect"
)

// IsZero 判断值是否为零值
// 支持指针类型,会递归判断指针指向的值
func IsZero(test any) bool {
	of := reflect.ValueOf(test)
	return isZero(of)
}

func isZero(of reflect.Value) bool {
	//nil
	if of.Kind() == reflect.Invalid {
		return true
	}
	//指针
	if of.Kind() == reflect.Pointer {
		return isZero(of.Elem())
	}
	//非指针
	if of.IsZero() {
		return true
	}
	return false
}

// New 创建类型T的新实例
// 对于指针类型(*T),返回新分配的*T;对于非指针类型(T),返回新分配的T
func New[T any](a T) T {
	t := reflect.TypeOf(a)
	isPtr := t.Kind() == reflect.Ptr
	if isPtr {
		t = t.Elem()
	}
	value := reflect.New(t)
	if isPtr {
		return value.Interface().(T)
	} else {
		r := value.Interface().(*T)
		return *r
	}
}

// Ptr 返回指向参数的指针
func Ptr[T any](t T) *T {
	return &t
}
