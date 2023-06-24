package pool

import (
    "math"
    "math/rand"
    "sync"
    "testing"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

const sleepTime = time.Millisecond * 3

func getSleepTime() time.Duration {
    return sleepTime * time.Duration(rand.Intn(100))
}

func TestGopool_Go(t *testing.T) {
    n := 20000
    println("start", n)
    wg := sync.WaitGroup{}
    wg.Add(n)
    for i := 0; i < n; i++ {
        Go(func() {
            time.Sleep(getSleepTime())
            wg.Done()
        })
    }
    println("done1", n)
    wg.Wait()
    println("done2", n)
}

func Benchmark_Go(b *testing.B) {
    n := b.N
    wg := sync.WaitGroup{}
    wg.Add(n)
    for i := 0; i < n; i++ {
        go func() {
            time.Sleep(getSleepTime())
            wg.Done()
        }()
    }
    wg.Wait()
}
func BenchmarkGopool_Go(b *testing.B) {
    n := b.N
    wg := sync.WaitGroup{}
    wg.Add(n)
    gopool := NewGopool(math.MaxInt32)
    for i := 0; i < n; i++ {
        gopool.Go(func() {
            time.Sleep(getSleepTime())
            wg.Done()
        })
    }
    wg.Wait()
}
