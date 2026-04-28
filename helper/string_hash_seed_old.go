//go:build !go1.19

package helper

import (
    "math/rand"
    "unsafe"
)

func newStringHashSeed() uintptr {
    return uintptr(rand.Int())
}

func init() {
    // Go 1.18 下使用从 map 提取 hasher 的方式
    a := any(make(map[string]struct{}))
    type maptype struct {
        _      [48]byte
        hasher func(unsafe.Pointer, uintptr) uintptr
    }
    strhashFunc = (**(**maptype)(unsafe.Pointer(&a))).hasher
}
