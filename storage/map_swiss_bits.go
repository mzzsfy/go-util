//go:build !amd64 || nosimd

package storage

import (
    "math/bits"
    "unsafe"
)

const (
    groupSize       = 8
    maxAvgGroupLoad = 7

    loBits uint64 = 0x0101010101010101
    hiBits uint64 = 0x8080808080808080
)

type bitset uint64

func metaMatchH2(m *metadata, h loByte) bitset {
    // https://graphics.stanford.edu/~seander/bithacks.html##ValueInWord
    return hasZeroByte(castUint64(m) ^ (loBits * uint64(h)))
}

func metaMatchEmpty(m *metadata) bitset {
    return hasZeroByte(castUint64(m) ^ hiBits)
}

func nextMatch(b *bitset) uint32 {
    s := uint32(bits.TrailingZeros64(uint64(*b)))
    *b &= ^(1 << s) // clear bit |s|
    return s >> 3   // div by 8
}

func hasZeroByte(x uint64) bitset {
    return bitset(((x - loBits) & ^(x)) & hiBits)
}

func castUint64(m *metadata) uint64 {
    return *(*uint64)((unsafe.Pointer)(m))
}
