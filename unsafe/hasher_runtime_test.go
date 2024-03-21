package unsafe

import (
    "strconv"
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
    t.Run("hash_int", func(t *testing.T) {
        h0 := NewHasher[int]()
        h1 := h0.NewSeed()
        h2 := h0.NewSeed()
        h3 := h0.NewSeed()
        for j := 0; j < 10; j++ {
            for _, h := range []Hasher[int]{h1, h2, h3} {
                t.Logf("hash%d:%b", j, h.Hash(j))
                if h.Hash(j) != h.Hash(j) {
                    t.Error("hash not equal")
                }
            }
        }
    })

    t.Run("hash_str", func(t *testing.T) {
        h0 := NewHasher[string]()
        h1 := h0.NewSeed()
        h2 := h0.NewSeed()
        h3 := h0.NewSeed()
        for j := 0; j < 10; j++ {
            for _, h := range []Hasher[string]{h1, h2, h3} {
                t.Logf("hash%d:%b", j, h.Hash(strconv.Itoa(j)))
                if h.Hash(strconv.Itoa(j)) != h.Hash(strconv.Itoa(j)) {
                    t.Error("hash not equal")
                }
            }
        }
    })

}
