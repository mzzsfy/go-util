package concurrent

import (
	"sync"
)

type Locker interface {
	sync.Locker
	TryLock() bool
}

// ReentrantLock 可重入锁
//
// 语义: Unlock责任在调用方, 与sync.Mutex一致——持锁goroutine退出未Unlock将导致锁永久泄漏
// (runtime无goroutine退出钩子, holder存活校验需遍历allgs, 检测代价与收益不成比例, 不做自动回收)
type ReentrantLock struct {
	cond      sync.Cond
	recursion int32
	_         [56]byte
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
	return &ReentrantLock{cond: *sync.NewCond(&sync.Mutex{}), goId: -1}
}
