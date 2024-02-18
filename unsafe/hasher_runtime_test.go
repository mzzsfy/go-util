package unsafe

import (
    "testing"
    "unsafe"
)

func Test_GetWithHash(t *testing.T) {
    hasher := getRuntimeHasher[int]()
    i := 1
    p := noescape(unsafe.Pointer(&i))
    t.Log("hash", hasher(p, newHashSeed()))
}
