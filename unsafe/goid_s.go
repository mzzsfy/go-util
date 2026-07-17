//go:build (arm64 || arm || amd64 || amd64p32 || 386) && go1.10

package unsafe

import (
    "unsafe"
)

//go:nocheckptr 读取 runtime 内部 g 结构体字段, 非堆对象, checkptr 无法合规
func GoID() int64 {
    p := (*int64)(unsafe.Pointer(getG() + goroutineIDOffset))
    return *p
}

func getG() uintptr
