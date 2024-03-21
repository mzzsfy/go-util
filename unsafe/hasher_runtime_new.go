//go:build go1.19

package unsafe

import _ "unsafe"

//go:linkname newHashSeed runtime.fastrand64
func newHashSeed() uintptr
