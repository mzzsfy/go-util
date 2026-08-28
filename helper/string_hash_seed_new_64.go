//go:build go1.19 && (amd64 || arm64 || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64 || s390x || wasm)

package helper

import _ "unsafe"

//go:linkname newStringHashSeed runtime.fastrand64
func newStringHashSeed() uintptr
