package helper

import (
    "github.com/mzzsfy/go-util/unsafe"
    "math"
)

type bloomFilter[T comparable] struct {
    hasher []func(T) uint64
    bits   []byte
}

func newBloomFilter[T comparable](expectedCount uint, failProbability float64) *bloomFilter[T] {
    bitsLen := uint(float64(-expectedCount) * math.Log(failProbability) / (math.Log(2) * math.Log(2)))
    hashNum := uint(math.Round(float64(bitsLen) / float64(expectedCount) * math.Log(2)))
    hasher := make([]func(T) uint64, hashNum)
    for i := range hasher {
        //fixme 如何实现多个hash函数,使用系统的hash函数,可能会有问题
        hasher[i] = unsafe.NewHasher[T]().NewSeed().Hash
    }
    return &bloomFilter[T]{hasher: hasher, bits: make([]byte, bitsLen/8+1)}
}
