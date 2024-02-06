package lock

import (
    "sync"
)

type Locker interface {
    sync.Locker
    TryLock() bool
}

type ReentrantLock struct {
    cond      *sync.Cond
    recursion int32
    _         [7]int64
    goId      int64
}

func (r *ReentrantLock) Lock() {
    r.cond.L.Lock()
    defer r.cond.L.Unlock()
    goId := GoID()
    if r.goId == goId {
        r.recursion++
        return
    }
    for r.recursion != 0 {
        r.cond.Wait()
    }
    r.goId = goId
    r.recursion = 1
}
func (r *ReentrantLock) TryLock() bool {
    r.cond.L.Lock()
    defer r.cond.L.Unlock()
    goId := GoID()
    if r.goId == goId {
        r.recursion++
        return true
    }
    if r.recursion != 0 {
        return false
    }
    r.goId = goId
    r.recursion = 1
    return true
}

func (r *ReentrantLock) Unlock() {
    r.cond.L.Lock()
    defer r.cond.L.Unlock()
    goId := GoID()
    if r.recursion == 0 || r.goId != goId {
        panic("unlock of unlocked lock")
    }
    r.recursion--
    if r.recursion == 0 {
        r.goId = -1
        r.cond.Signal()
    }
}

func NewReentrantLock() Locker {
    return &ReentrantLock{cond: sync.NewCond(&sync.Mutex{}), goId: -1}
}
