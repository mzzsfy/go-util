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
