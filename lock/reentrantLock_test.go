package lock

import (
    "testing"
    "time"
)

func Test_ReentrantLock_Lock_Unlock(t *testing.T) {
    lock := NewReentrantLock()
    go func() {
        lock.Lock()
        time.Sleep(100 * time.Millisecond)
        lock.Unlock()
    }()
    time.Sleep(50 * time.Millisecond)
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
    lock := NewReentrantLock()
    if lock.TryLock() {
        lock.Unlock()
    } else {
        t.Fatal("TryLock should return true")
    }
    go func() {
        lock.Lock()
        time.Sleep(50 * time.Millisecond)
        lock.Unlock()
    }()
    time.Sleep(10 * time.Millisecond)
    if lock.TryLock() {
        lock.Unlock()
        t.Fatal("TryLock should return false")
    }
}
