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
                //t.Logf("hash%d:%b", j, h.Hash(j))
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
                //t.Logf("hash%d:%b", j, h.Hash(strconv.Itoa(j)))
                if h.Hash(strconv.Itoa(j)) != h.Hash(strconv.Itoa(j)) {
                    t.Error("hash not equal")
                }
            }
        }
    })

}

// BenchmarkNewHasher 测试 Hasher 创建性能
func BenchmarkNewHasher(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        NewHasher[int]()
    }
}

// BenchmarkHash_Int 测试 int 类型哈希性能
func BenchmarkHash_Int(b *testing.B) {
    b.ReportAllocs()
    h := NewHasher[int]()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        h.Hash(i)
    }
}

// BenchmarkHash_String 测试 string 类型哈希性能
func BenchmarkHash_String(b *testing.B) {
    b.ReportAllocs()
    h := NewHasher[string]()
    data := "benchmark test string"
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        h.Hash(data)
    }
}

// BenchmarkHash_String_Different 测试不同字符串的哈希性能
func BenchmarkHash_String_Different(b *testing.B) {
    b.ReportAllocs()
    h := NewHasher[string]()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        h.Hash(strconv.Itoa(i))
    }
}
