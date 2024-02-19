package helper

import (
    "reflect"
)

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
