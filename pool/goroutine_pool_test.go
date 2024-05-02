package pool

import (
    "github.com/mzzsfy/go-util/unsafe"
    "math/rand"
    "runtime"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

const sleepTime = time.Millisecond * 3

func getSleepTime() time.Duration {
    return sleepTime * time.Duration(rand.Intn(100))
}

func TestGopool_Go(t *testing.T) {
    t.Parallel()
    n := 200000
    x := int32(n)
    maxGoroutine := 0
    t.Log("start", n)
    lock := sync.Mutex{}
    wg := sync.WaitGroup{}
    wg.Add(n)
    pool := NewGopool()
    for i := 0; i < n; i++ {
        pool.Go(func() {
            time.Sleep(getSleepTime())
            goroutine := runtime.NumGoroutine()
            lock.Lock()
            if goroutine > maxGoroutine {
                maxGoroutine = goroutine
            }
            lock.Unlock()
            wg.Done()
            atomic.AddInt32(&x, -1)
        })
    }
    t.Log("done1", n)
    wg.Wait()
    pool.Shutdown()
    t.Log("done2", n)
    count := defaultGoPool.TaskCount()
    if x != 0 {
        t.Fatal("x != 0", x)
    }
    if count != 0 {
        t.Fatal("count != 0", count)
    }
    t.Log("maxGoroutine", maxGoroutine)
}

func TestGopool_Go1(t *testing.T) {
    t.Parallel()
    n := 20000
    x := int32(n)
    maxGoroutine := 0
    println("start", n)
    lock := sync.Mutex{}
    wg := sync.WaitGroup{}
    wg.Add(n)
    pool := NewGopool()
    for i := 0; i < n; i++ {
        pool.Go(func() {
            time.Sleep(getSleepTime())
            goroutine := runtime.NumGoroutine()
            lock.Lock()
            if goroutine > maxGoroutine {
                maxGoroutine = goroutine
            }
            lock.Unlock()
            wg.Done()
            atomic.AddInt32(&x, -1)
        })
    }
    t.Log("done1", n)
    wg.Wait()
    pool.Shutdown()
    t.Log("done2", n)
    if x != 0 {
        t.Fatal("x != 0", x)
    }
    if defaultGoPool.TaskCount() != 0 {
        t.Fatal("count != 0", defaultGoPool.TaskCount())
    }
    t.Log("maxGoroutine", maxGoroutine)
}

func Benchmark_Go(b *testing.B) {
    n := b.N
    x := int32(n)
    wg := sync.WaitGroup{}
    wg.Add(n)
    var goid int64
    go func() {
        goid = unsafe.GoID()
    }()
    for i := 0; i < n; i++ {
        go func() {
            time.Sleep(getSleepTime())
            atomic.AddInt32(&x, -1)
            wg.Done()
        }()
    }
    wg.Wait()
    if b.N > 600000 {
        time.Sleep(time.Millisecond * 50)
        go func() {
            b.Log("goid start", goid, "end", unsafe.GoID(), "new", unsafe.GoID()-goid)
        }()
    }
    if x != 0 {
        b.Fatal("x != 0", x)
    }
    if defaultGoPool.TaskCount() != 0 {
        b.Fatal("count != 0", defaultGoPool.TaskCount())
    }
    time.Sleep(time.Millisecond * 50)
}

func BenchmarkGopool_Go(b *testing.B) {
    wg := sync.WaitGroup{}
    n := b.N
    x := int32(n)
    wg.Add(n)
    gopool := NewGopool()
    var goid int64
    go func() {
        goid = unsafe.GoID()
    }()
    for i := 0; i < n; i++ {
        gopool.Go(func() {
            time.Sleep(getSleepTime())
            atomic.AddInt32(&x, -1)
            wg.Done()
        })
    }
    wg.Wait()
    if b.N > 600000 {
        time.Sleep(time.Millisecond * 50)
        go func() {
            b.Log("goid start", goid, "end", unsafe.GoID(), "new", unsafe.GoID()-goid)
        }()
    }
    if x != 0 {
        b.Fatal("x != 0", x)
    }
    if defaultGoPool.TaskCount() != 0 {
        b.Fatal("count != 0", defaultGoPool.TaskCount())
    }
    time.Sleep(time.Millisecond * 50)
}
