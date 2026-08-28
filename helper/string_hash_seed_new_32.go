//go:build go1.19 && (386 || arm || mips || mipsle)

package helper

import _ "unsafe"

// 32位平台uintptr宽度为4, fastrand64按uint64写返回值会越界破坏调用栈, 改用fastrand
//
//go:linkname newStringHashSeed runtime.fastrand
func newStringHashSeed() uintptr
