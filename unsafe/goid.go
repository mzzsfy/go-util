package unsafe

import (
    "unsafe"
)

const gGoroutineIDOffset = 152 // more Go1.10

func GoID() int64 {
    p := (*int64)(unsafe.Pointer(getG() + gGoroutineIDOffset))
    return *p
}

func getG() uintptr
