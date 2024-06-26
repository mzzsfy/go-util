//go:build amd64 && !nosimd

package storage

import (
    "math/bits"
    _ "unsafe"
)

const (
    groupSize       = 16
    maxAvgGroupLoad = 14
)

type bitset uint16

var supportsSimd = func() bool {
    _, _, ecx1, edx1 := Cpuid(1, 0)
    sse2 := edx1&(1<<26) != 0
    if !sse2 {
        println("you cpu not support SSE2, falling back to slower implementation")
    }
    ssse3 := ecx1&(1<<9) != 0
    if !ssse3 {
        println("you cpu not support SSSE3, falling back to slower implementation")
    }
    return sse2 && ssse3
}()

func metaMatchH2(m *metadata, h loByte) bitset {
    if supportsSimd {
        return bitset(matchMetadata((*[16]int8)(m), int8(h)))
    }
    return bitset(matchMetadataFallback((*[16]int8)(m), int8(h)))
}

func metaMatchEmpty(m *metadata) bitset {
    if supportsSimd {
        return bitset(matchMetadata((*[16]int8)(m), empty))
    }
    return bitset(matchMetadataFallback((*[16]int8)(m), empty))
}

func nextMatch(b *bitset) (s uint32) {
    s = uint32(bits.TrailingZeros16(uint16(*b)))
    *b &= ^(1 << s) // clear bit |s|
    return
}

// matchMetadataFallback is the Go implementation used when SSSE3 is not supported.
func matchMetadataFallback(metadata *[16]int8, hash int8) uint16 {
    var result uint16
    for i := 0; i < 16; i++ {
        if metadata[i] == hash {
            result |= 1 << uint(i)
        }
    }
    return result
}

// matchMetadata performs a 16-way probe of |metadata| using SSE instructions
// nb: |metadata| must be an aligned pointer
// Requires: SSE2, SSSE3
//go:noescape
func matchMetadata(metadata *[16]int8, hash int8) uint16

// Cpuid is implemented in cpu_x86.s for gc compiler
// and in cpu_gccgo.c for gccgo.
// copy from golang.org/x/sys/cpu
func Cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)
