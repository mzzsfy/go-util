package concurrent

import (
    "runtime"
    "sync/atomic"
)

type RwLocker interface {
    Locker
    RLock()
    RUnlock()
    TryRLock() bool
}

// CasRwLocker 基于CAS实现的读写锁,性能不如sync.RWMutex,仅保留用于基准测试对比
// Deprecated: 使用sync.RWMutex替代,CAS实现在高并发场景下性能较差
type CasRwLocker struct {
    lock int32
}

func (c *CasRwLocker) Lock() {
    for i := 0; ; i++ {
        if atomic.CompareAndSwapInt32(&c.lock, 0, -1) {
            return
        }
        if i > 10 {
            runtime.Gosched()
        }
    }
}

func (c *CasRwLocker) Unlock() {
    if !atomic.CompareAndSwapInt32(&c.lock, -1, 0) {
        panic("unlock of unlocked lock")
    }
}

func (c *CasRwLocker) TryLock() bool {
    return atomic.CompareAndSwapInt32(&c.lock, 0, -1)
}

func (c *CasRwLocker) RLock() {
    for i := 0; ; i++ {
        v := atomic.LoadInt32(&c.lock)
        if v != -1 && atomic.CompareAndSwapInt32(&c.lock, v, v+1) {
            return
        }
        if i > 10 {
            runtime.Gosched()
        }
    }
}

func (c *CasRwLocker) RUnlock() {
    for i := 0; ; i++ {
        v := atomic.LoadInt32(&c.lock)
        if v < 1 {
            panic("unlock of unlocked lock")
        }
        if atomic.CompareAndSwapInt32(&c.lock, v, v-1) {
            return
        }
        if i > 10 {
            runtime.Gosched()
        }
    }
}

func (c *CasRwLocker) TryRLock() bool {
    for i := 0; ; i++ {
        v := atomic.LoadInt32(&c.lock)
        if v != -1 && atomic.CompareAndSwapInt32(&c.lock, v, v+1) {
            return true
        }
        if i > 10 {
            return false
        }
    }
}

// NoLock 一个空锁,用于站位
type NoLock struct{}

func (l NoLock) Lock()         {}
func (l NoLock) Unlock()       {}
func (l NoLock) TryLock() bool { return true }

func (l NoLock) RLock()         {}
func (l NoLock) RUnlock()       {}
func (l NoLock) TryRLock() bool { return true }
