//go:build !go1.19

package unsafe

import "math/rand"

func newHashSeed() uintptr {
    return uintptr(rand.Int())
}
