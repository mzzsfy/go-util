//go:build go1.19

package helper

import "unsafe"

//go:linkname newStringHashSeed runtime.fastrand64
func newStringHashSeed() uintptr

//go:linkname runtimeStrhash runtime.strhash
func runtimeStrhash(p unsafe.Pointer, h uintptr) uintptr

func init() {
    strhashFunc = runtimeStrhash
}
