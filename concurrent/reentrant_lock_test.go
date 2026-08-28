package concurrent

import (
	"testing"
	"time"
)

func Test_ReentrantLock_Lock_Unlock(t *testing.T) {
	t.Parallel()
	lock := NewReentrantLock()
	locked := make(chan struct{})
	go func() {
		lock.Lock()
		close(locked)
		time.Sleep(100 * time.Millisecond)
		lock.Unlock()
	}()
	// 事件同步: 确认对方已持锁, 主 goroutine Lock 必然阻塞等其释放
	<-locked
	lock.Lock()
	lock.Unlock()
}

func Test_ReentrantLock_DoubleLock_Unlock(t *testing.T) {
	lock := NewReentrantLock()
	lock.Lock()
	lock.Lock()
	lock.Unlock()
	lock.Unlock()
}

func Test_ReentrantLock_UnlockWithoutLock(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("The code did not panic")
		}
	}()
	lock := NewReentrantLock()
	lock.Unlock()
}

func Test_ReentrantLock_DoubleUnlock(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	lock := NewReentrantLock()
	lock.Lock()
	lock.Unlock()
	lock.Unlock()
}

func Test_ReentrantLock_TryLock(t *testing.T) {
	t.Parallel()
	lock := NewReentrantLock()
	if lock.TryLock() {
		lock.Unlock()
	} else {
		t.Fatal("TryLock should return true")
	}
	locked := make(chan struct{})
	go func() {
		lock.Lock()
		close(locked)
		time.Sleep(50 * time.Millisecond)
		lock.Unlock()
	}()
	// 事件同步: 确认对方已持锁后再尝试, 消除调度时序竞态
	<-locked
	if lock.TryLock() {
		lock.Unlock()
		t.Fatal("TryLock should return false")
	}
}
