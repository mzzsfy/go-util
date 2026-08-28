package storage

import (
	"github.com/mzzsfy/go-util/unsafe"
	"runtime"
	"sync"
)

// nextPowerOfTwo 返回大于等于n的最小2的幂次方
func nextPowerOfTwo(n int) int {
	if n <= 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n + 1
}

// slotNumber 分片数, slotMask 分片位掩码, slotIdx 通过位运算定位分片
var (
	slotNumber = func() int {
		n := runtime.NumCPU()
		if n <= 1 {
			return 1
		}
		return nextPowerOfTwo(n)
	}()
	slotMask = uint64(slotNumber - 1)
)

// slotIdx 用位掩码替代取模, 编译器可内联
func slotIdx(hash uint64) int {
	return int(hash & slotMask)
}

// paddedLock 填充至 cache line 大小(64B), 避免并发场景下相邻锁 false sharing
type paddedLock struct {
	sync.RWMutex
	_ [40]byte // sizeof(sync.RWMutex)=24 + 40 = 64
}

func NewDefaultHasher[K comparable]() unsafe.Hasher[K] {
	return unsafe.NewHasher[K]()
}
