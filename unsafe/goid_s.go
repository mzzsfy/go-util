//go:build arm64 || arm || amd64 || amd64p32 || 386

package unsafe

import (
    "unsafe"
)

const goroutineIDOffset = 152 // more Go1.10

func GoID() int64 {
    p := (*int64)(unsafe.Pointer(getG() + goroutineIDOffset))
    return *p
}

func getG() uintptr
