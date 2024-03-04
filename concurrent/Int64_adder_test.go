package concurrent

import (
    "runtime"
    "strconv"
    "sync"
    "sync/atomic"
    "testing"
)

func Test_Int64AdderAdd(t *testing.T) {
    adder := &Int64Adder{}
    adder.Add(1, 10)
    if adder.Sum() != 10 {
        t.Errorf("Expected sum to be 10, got %d", adder.Sum())
    }
    t.Logf("sum: %d", adder.Sum())
}

func Test_Int64AdderIncrementSimple(t *testing.T) {
    adder := &Int64Adder{}
    adder.IncrementSimple()
    if adder.Sum() != 1 {
        t.Errorf("Expected sum to be 1, got %d", adder.Sum())
    }
}

func Test_Int64AdderDecrement(t *testing.T) {
    adder := &Int64Adder{}
    adder.Add(1, 10)
    adder.Decrement(1)
    if adder.Sum() != 9 {
        t.Errorf("Expected sum to be 9, got %d", adder.Sum())
    }
}

func Test_Int64AdderIncrement(t *testing.T) {
    adder := &Int64Adder{}
    adder.Increment(1)
    if adder.Sum() != 1 {
        t.Errorf("Expected sum to be 1, got %d", adder.Sum())
    }
}

func Test_Int64AdderDecrementSimple(t *testing.T) {
    adder := &Int64Adder{}
    adder.AddSimple(10)
    adder.DecrementSimple()
    if adder.Sum() != 9 {
        t.Errorf("Expected sum to be 9, got %d", adder.Sum())
    }
}

func Test_Int64AdderAddSimple(t *testing.T) {
    adder := &Int64Adder{}
    adder.AddSimple(10)
    if adder.Sum() != 10 {
        t.Errorf("Expected sum to be 10, got %d", adder.Sum())
    }
}

func Test_Int64AdderSum(t *testing.T) {
    adder := &Int64Adder{}
    adder.AddSimple(10)
    adder.AddSimple(20)
    if adder.Sum() != 30 {
        t.Errorf("Expected sum to be 30, got %d", adder.Sum())
    }
}

func Test_Int64AdderSumInt(t *testing.T) {
    adder := &Int64Adder{}
    adder.AddSimple(10)
    adder.AddSimple(20)
    if adder.SumInt() != 30 {
        t.Errorf("Expected sum to be 30, got %d", adder.SumInt())
    }
}

func Test_Int64AdderReset(t *testing.T) {
    adder := &Int64Adder{}
    adder.AddSimple(10)
    adder.AddSimple(20)
    adder.Reset()
    if adder.Sum() != 0 {
        t.Errorf("Expected sum to be 0, got %d", adder.Sum())
    }
}

func Test_Int64Adder(t *testing.T) {
    adder := &Int64Adder{}
    wg := sync.WaitGroup{}
    wg.Add(100)
    for i := 0; i < 100; i++ {
        go func() {
            defer wg.Done()
            for i := 0; i < 100; i++ {
                adder.IncrementSimple()
                runtime.Gosched()
            }
        }()
    }
    wg.Wait()
    if adder.Sum() != 10000 {
        t.Errorf("Expected sum to be 10000, got %d", adder.Sum())
    }
}

func Benchmark1Int64Adder(b *testing.B) {
    goroutineNum := 1000
    for _, i := range []int{2, 4, 8, 32, 64, 128} {
        b.Run("Int64Adder_"+strconv.Itoa(i), func(b *testing.B) {
            adder := &Int64Adder{}
            wg := NewWaitGroup(goroutineNum)
            wg1 := NewWaitGroup(goroutineNum)
            wg2 := NewWaitGroup(1)
            //使用 waitGroup 保证所有的 goroutine 都已经启动
            for x := 0; x < goroutineNum; x++ {
                go func() {
                    wg1.Done()
                    defer wg.Done()
                    wg2.Wait()
                    id := GoID()
                    for j := 0; j < b.N; j++ {
                        adder.Add(id, 1)
                        if j%i == 0 {
                            runtime.Gosched()
                        }
                    }
                }()
            }
            wg1.Wait()
            b.ResetTimer()
            wg2.Done()
            wg.Wait()
        })
        b.Run("Atomic_"+strconv.Itoa(i), func(b *testing.B) {
            r := int64(0)
            wg := NewWaitGroup(b.N)
            wg1 := NewWaitGroup(b.N)
            wg2 := NewWaitGroup(1)
            //使用 waitGroup 保证所有的 goroutine 都已经启动
            for x := 0; x < b.N; x++ {
                go func() {
                    wg1.Done()
                    defer wg.Done()
                    wg2.Wait()
                    for j := 0; j < 1000; j++ {
                        atomic.AddInt64(&r, 1)
                        if j%i == 0 {
                            runtime.Gosched()
                        }
                    }
                }()
            }
            wg1.Wait()
            b.ResetTimer()
            wg2.Done()
            wg.Wait()
        })
    }
}
