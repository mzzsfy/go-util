package unsafe

import (
    "unsafe"
)

type hashFn func(unsafe.Pointer, uintptr) uintptr

func getRuntimeHasher[K comparable]() (h hashFn) {
    a := any(make(map[K]struct{}))
    h = (**(**maptype)(unsafe.Pointer(&a))).hasher
    return
}

var hashSeed = newHashSeed()

//go:linkname newHashSeed runtime.fastrand64
func newHashSeed() uintptr

// noescape hides a pointer from escape analysis. It is the identity function
// but escape analysis doesn't think the output depends on the input.
// noescape is inlined and currently compiles down to zero instructions.
// USE CAREFULLY!
// This was copied from the runtime (via pkg "strings"); see issues 23382 and 7921.
//
//go:nosplit
//go:nocheckptr
func noescape(p unsafe.Pointer) unsafe.Pointer {
    x := uintptr(p)
    return unsafe.Pointer(x ^ 0)
}
