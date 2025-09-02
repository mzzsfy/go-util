package storage

import (
    "github.com/mzzsfy/go-util/unsafe"
    "runtime"
)

var slotNumber, modNumber = func() (int, int) {
    numCPU := runtime.NumCPU()
    if numCPU <= 1 {
        return 1, 1
    }
    return numCPU, numCPU - 1
}()
var idxFn = func(hash uint64) int {
    //去除符号位,避免可能的负数
    return int(hash&0x7FFFFFFFFFFFFFFF) % modNumber
}

func init() {
    //32位系统,int为4字节,uint64强转int有损失
    if ^uint(0)>>63 == 0 {
        idxFn = func(key uint64) int {
            //高32位和低32位异或,减少hash冲突
            return int((key^(key>>32))&0x7FFFFFFF) % modNumber
        }
    }
}

func NewDefaultHasher[K comparable]() unsafe.Hasher[K] {
    return unsafe.NewHasher[K]()
}
