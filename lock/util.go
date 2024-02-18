package lock

import "sync"

type Helper struct {
    sync.Locker
}

func (h Helper) RunWithLock(f func()) {
    h.Lock()
    defer h.Unlock()
    f()
}

// Lock1 锁包装,用法
// defer lock.Lock1()()
func (h Helper) Lock1() func() {
    h.Locker.Lock()
    return h.Locker.Unlock
}
