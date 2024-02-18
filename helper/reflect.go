package helper

import (
    "reflect"
)

// Eq 判断是否相等,包装了一下,也许能优化
func Eq[T any](test1, test2 T) bool {
    of1 := reflect.ValueOf(test1)
    if of1.Kind() == reflect.Invalid {
        of2 := reflect.ValueOf(test2)
        return of2.Kind() == reflect.Invalid
    }
    //指针等,仅判断指针指向地址是否一致
    switch of1.Kind() {
    case reflect.Pointer, reflect.Func, reflect.Chan:
        of2 := reflect.ValueOf(test2)
        return of1.Pointer() == of2.Pointer()
    }
    //其他
    return reflect.DeepEqual(test1, test2)
}
func EqPtrDeep[T any](test1, test2 *T) bool {
    return EqPtr(test1, test2, true)
}
func EqPtrNotDeep[T any](test1, test2 *T) bool {
    return EqPtr(test1, test2, false)
}

// EqPtr 比较指针,指针比较失败则比较值是否相等,deepEqual为true时,额外使用reflect.DeepEqual比较
func EqPtr[T any](test1, test2 *T, deepEqual ...bool) bool {
    if len(deepEqual) == 0 || !deepEqual[0] {
        return test1 == test2
    }
    of1 := reflect.ValueOf(test1)
    if of1.Kind() == reflect.Invalid {
        of2 := reflect.ValueOf(test2)
        return of2.Kind() == reflect.Invalid
    }
    return reflect.DeepEqual(reflect.ValueOf(test1).Elem().Interface(), reflect.ValueOf(test2).Elem().Interface())
}

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
