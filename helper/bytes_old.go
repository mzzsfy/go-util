//go:build !go1.21

// 代表含义为 go version not in (1.1~1.21)

package helper

import (
    "unsafe"
)

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(
        &struct {
            string
            int
        }{s, len(s)},
    ))
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}
