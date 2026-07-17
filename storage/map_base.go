package storage

import (
    "github.com/mzzsfy/go-util/unsafe"
    "runtime"
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

var slotNumber, modNumber = func() (int, int) {
    numCPU := runtime.NumCPU()
    if numCPU <= 1 {
        return 1, 1
    }
    modNum := nextPowerOfTwo(numCPU)
    return modNum, modNum
}()
var idxFn = func(hash uint64) int {
    // 使用位运算替代取模,modNumber为2的幂次方
    return int(hash & uint64(modNumber-1))
}

func init() {
    //32位系统,int为4字节,uint64强转int有损失
    if ^uint(0)>>63 == 0 {
        idxFn = func(key uint64) int {
            //高32位和低32位异或,减少hash冲突,使用位运算替代取模
            return int((key ^ (key >> 32)) & uint64(modNumber-1))
        }
    }
}

func NewDefaultHasher[K comparable]() unsafe.Hasher[K] {
    return unsafe.NewHasher[K]()
}
