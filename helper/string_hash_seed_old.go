//go:build !go1.19

package helper

import (
	"math/rand"
	"unsafe"
)

//go:linkname runtimeStrhash runtime.strhash
func runtimeStrhash(p unsafe.Pointer, h uintptr) uintptr

func newStringHashSeed() uintptr {
	return uintptr(rand.Int())
}

func init() {
	strhashFunc = runtimeStrhash
}
