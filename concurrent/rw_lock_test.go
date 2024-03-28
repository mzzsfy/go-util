package concurrent

import (
    "runtime"
    "sync"
    "testing"
)

func TestCasRwLocker(t *testing.T) {
    var lock RwLocker = &CasRwLocker{}
    n := 10
    n1 := 10000
    wg := NewWaitGroup(n)
    x := 0
    for i := 0; i < n; i++ {
        go func() {
            defer wg.Done()
            for i := 0; i < n1; i++ {
                lock.RLock()
                x1 := x
                lock.RUnlock()
                lock.Lock()
                x1 = x
                x = x1 + 1
                lock.Unlock()
                if i%100 == 0 {
                    runtime.Gosched()
                }
            }
        }()
    }
    wg.Wait()
    if x != n*n1 {
        t.Fatal("x != n * n1")
    }
}

func BenchmarkCasRwLocker_Lock(b *testing.B) {
    nr := 100
    for _, lock := range []struct {
        name string
        RwLocker
    }{
        {"noLock", NoLock{}},
        {"casRwLocker", &CasRwLocker{}},
        {"sync.RWMutex", &sync.RWMutex{}},
    } {
        b.Run("lock_"+lock.name, func(b *testing.B) {
            x := 0
            b.SetParallelism(2)
            b.RunParallel(
                func(pb *testing.PB) {
                    for pb.Next() {
                        for i := 0; i < nr; i++ {
                            lock.RLock()
                            x1 := x
                            x1++
                            lock.RUnlock()
                        }
                        lock.Lock()
                        x++
                        lock.Unlock()
                    }
                })
        })
        b.Run("lock_"+lock.name+"_sync", func(b *testing.B) {
            x := 0
            for i := 0; i < b.N; i++ {
                for i := 0; i < nr; i++ {
                    lock.RLock()
                    x1 := x
                    x1++
                    lock.RUnlock()
                }
                lock.Lock()
                x++
                lock.Unlock()
            }
        })
    }
}
