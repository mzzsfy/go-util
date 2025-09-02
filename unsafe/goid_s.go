//go:build (arm64 || arm || amd64 || amd64p32 || 386) && go1.10

package unsafe

import (
    "unsafe"
)

func GoID() int64 {
    p := (*int64)(unsafe.Pointer(getG() + goroutineIDOffset))
    return *p
}

func getG() uintptr
