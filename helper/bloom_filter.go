package helper

import (
    "github.com/mzzsfy/go-util/unsafe"
    "math"
)

type BloomFilter[T any] interface {
    Has(T) bool
    Add(T)
}

type bloomFilter[T comparable] struct {
    hasher []func(T) uint64
    bits   []bool
}

func (b *bloomFilter[T]) Has(t T) bool {
    for _, hash := range b.hasher {
        if !b.bits[hash(t)%uint64(len(b.bits))] {
            return false
        }
    }
    return true
}

func (b *bloomFilter[T]) Add(t T) {
    for _, hash := range b.hasher {
        b.bits[hash(t)%uint64(len(b.bits))] = true
    }
}

var (
    // seeds 为固定哈希种子,目前是随便写的,用于生产哈希函数时固定种子,这样可以保证哈希函数不会变化
    seeds = [10]uintptr{
        0x0,
        0xf1b3d,
        0xfafff231,
        0xfaff32818,
        0xffcff161339,
        0xffcf5358da932,
        0xaefff632426433,
        0xf0e7f6fe2795088,
        0xaa98135fb0971694,
        0xf99ea26399b43751,
    }
)

// NewBloomFilter 创建一个布隆过滤器, expectedCount 为预期元素数量, failProbability 为错误概率
func NewBloomFilter[T comparable](expectedCount uint, failProbability float64) BloomFilter[T] {
    //todo:系统内部hash函数效果不太行,但是这样会导致内存占用增加
    bitsLen := uint(math.Ceil(-1 * float64(expectedCount) * 2 * math.Log(failProbability) / math.Pow(math.Log(2), 2)))
    hashNum := uint(math.Ceil(math.Log(2) * float64(bitsLen) / float64(expectedCount)))
    bitsLen = Max(bitsLen, 1)
    hashNum = Max(hashNum, 1)
    hasher := make([]func(T) uint64, hashNum)
    for i := range hasher {
        hasher[i] = unsafe.NewHasher[T]().WithSeed(
            seeds[i%10] ^ (uintptr(i) << 16) ^ uintptr(i),
            //uintptr(rand.Uint64()),
        ).Hash
    }
    return &bloomFilter[T]{hasher: hasher, bits: make([]bool, bitsLen/8+1)}
}
