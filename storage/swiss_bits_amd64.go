//nolint:all
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

//nolint:all
func metaMatchH2(m *metadata, h h2) bitset {
    b := MatchMetadata((*[16]int8)(m), int8(h))
    return bitset(b)
}

//nolint:all
func metaMatchEmpty(m *metadata) bitset {
    b := MatchMetadata((*[16]int8)(m), empty)
    return bitset(b)
}

//nolint:all
func nextMatch(b *bitset) (s uint32) {
    s = uint32(bits.TrailingZeros16(uint16(*b)))
    *b &= ^(1 << s) // clear bit |s|
    return
}

//nolint:all
// MatchMetadata performs a 16-way probe of |metadata| using SSE instructions
// nb: |metadata| must be an aligned pointer
func MatchMetadata(metadata *[16]int8, hash int8) uint16
