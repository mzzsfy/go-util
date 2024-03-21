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

func Test_Hash(t *testing.T) {
    h := NewHasher[int]()
    h1 := h.NewSeed()
    h2 := h.NewSeed()
    h3 := h.NewSeed()
    for j := 0; j < 10; j++ {
        t.Logf("hash%d:%b", j, h1.Hash(j))
        if h1.Hash(j) != h1.Hash(j) {
            t.Error("hash1 not equal")
        }
        //}
        //for j := 0; j < 10; j++ {
        t.Logf("hash%d:%b", j, h2.Hash(j))
        //}
        //for j := 0; j < 10; j++ {
        t.Logf("hash%d:%b", j, h3.Hash(j))
    }
}
