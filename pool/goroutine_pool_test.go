package pool

import (
    "math/rand"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

const sleepTime = time.Millisecond * 3

func getSleepTime() time.Duration {
    return sleepTime * time.Duration(rand.Intn(100))
}

//
//func TestGopool_Go(t *testing.T) {
//    t.Parallel()
//    n := 200000
//    x := int32(n)
//    maxGoroutine := 0
//    println("start", n)
//    lock := sync.Mutex{}
//    wg := sync.WaitGroup{}
//    wg.Add(n)
//    pool := NewGopool()
//    for i := 0; i < n; i++ {
//        pool.Go(func() {
//            time.Sleep(getSleepTime())
//            goroutine := runtime.NumGoroutine()
//            lock.Lock()
//            if goroutine > maxGoroutine {
//                maxGoroutine = goroutine
//            }
//            lock.Unlock()
//            wg.Done()
//            atomic.AddInt32(&x, -1)
//        })
//    }
//    println("done1", n)
//    wg.Wait()
//    pool.Shutdown()
//    println("done2", n)
//    count := defaultGoPool.TaskCount()
//    if x != 0 {
//        t.Fatal("x != 0", x)
//    }
//    if count != 0 {
//        t.Fatal("count != 0", count)
//    }
//    t.Log("maxGoroutine", maxGoroutine)
//}
//
//func TestGopool_Go1(t *testing.T) {
//    t.Parallel()
//    n := 20000
//    x := int32(n)
//    maxGoroutine := 0
//    println("start", n)
//    lock := sync.Mutex{}
//    wg := sync.WaitGroup{}
//    wg.Add(n)
//    pool := NewGopool(WithMaxSize(1000))
//    for i := 0; i < n; i++ {
//        pool.Go(func() {
//            time.Sleep(getSleepTime())
//            goroutine := runtime.NumGoroutine()
//            lock.Lock()
//            if goroutine > maxGoroutine {
//                maxGoroutine = goroutine
//            }
//            lock.Unlock()
//            wg.Done()
//            atomic.AddInt32(&x, -1)
//        })
//    }
//    println("done1", n)
//    wg.Wait()
//    pool.Shutdown()
//    println("done2", n)
//    if x != 0 {
//        t.Fatal("x != 0", x)
//    }
//    if defaultGoPool.TaskCount() != 0 {
//        t.Fatal("count != 0", defaultGoPool.TaskCount())
//    }
//    t.Log("maxGoroutine", maxGoroutine)
//}
//
//func Benchmark_Go(b *testing.B) {
//    n := b.N
//    x := int32(n)
//    wg := sync.WaitGroup{}
//    wg.Add(n)
//    for i := 0; i < n; i++ {
//        go func() {
//            time.Sleep(getSleepTime())
//            atomic.AddInt32(&x, -1)
//            wg.Done()
//        }()
//    }
//    wg.Wait()
//    if x != 0 {
//        b.Fatal("x != 0", x)
//    }
//    if defaultGoPool.TaskCount() != 0 {
//        b.Fatal("count != 0", defaultGoPool.TaskCount())
//    }
//}
//func BenchmarkGopool_Go(b *testing.B) {
//    wg := sync.WaitGroup{}
//    n := b.N
//    x := int32(n)
//    wg.Add(n)
//    gopool := NewGopool()
//    for i := 0; i < n; i++ {
//        gopool.Go(func() {
//            time.Sleep(getSleepTime())
//            atomic.AddInt32(&x, -1)
//            wg.Done()
//        })
//    }
//    wg.Wait()
//    if x != 0 {
//        b.Fatal("x != 0", x)
//    }
//    if defaultGoPool.TaskCount() != 0 {
//        b.Fatal("count != 0", defaultGoPool.TaskCount())
//    }
//}
//
//func BenchmarkGopool_Go1(b *testing.B) {
//    n := b.N
//    x := int32(n)
//    wg := sync.WaitGroup{}
//    wg.Add(n)
//    gopool := NewGopool(WithLoadProbability(0.4))
//    for i := 0; i < n; i++ {
//        gopool.Go(func() {
//            time.Sleep(getSleepTime())
//            atomic.AddInt32(&x, -1)
//            wg.Done()
//        })
//    }
//    wg.Wait()
//    if x != 0 {
//        b.Fatal("x != 0", x)
//    }
//    if defaultGoPool.TaskCount() != 0 {
//        b.Fatal("count != 0", defaultGoPool.TaskCount())
//    }
//}
//
//func BenchmarkGopool_Go2(b *testing.B) {
//    n := b.N
//    x := int32(n)
//    wg := sync.WaitGroup{}
//    wg.Add(n)
//    gopool := NewGopool(WithCoreSize(math.MaxUint32))
//    for i := 0; i < n; i++ {
//        gopool.Go(func() {
//            time.Sleep(getSleepTime())
//            atomic.AddInt32(&x, -1)
//            wg.Done()
//        })
//    }
//    wg.Wait()
//    if x != 0 {
//        b.Fatal("x != 0", x)
//    }
//    if defaultGoPool.TaskCount() != 0 {
//        b.Fatal("count != 0", defaultGoPool.TaskCount())
//    }
//}
//
//func BenchmarkGopool_Go3(b *testing.B) {
//    n := b.N
//    x := int32(n)
//    wg := sync.WaitGroup{}
//    wg.Add(n)
//    gopool := NewGopool(WithLoadProbability(1))
//    for i := 0; i < n; i++ {
//        gopool.Go(func() {
//            time.Sleep(getSleepTime())
//            atomic.AddInt32(&x, -1)
//            wg.Done()
//        })
//    }
//    wg.Wait()
//    if x != 0 {
//        b.Fatal("x != 0", x)
//    }
//    if defaultGoPool.TaskCount() != 0 {
//        b.Fatal("count != 0", defaultGoPool.TaskCount())
//    }
//}
