package unsafe

import (
    "testing"
    "unsafe"
)

func Test_GetWithHash(t *testing.T) {
    h := getRuntimeHasher[int]()
    i := 1
    p := noescape(unsafe.Pointer(&i))
    t.Log("hash", h(p, newHashSeed()))
}
