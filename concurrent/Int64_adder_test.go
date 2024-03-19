package concurrent

import (
    "runtime"
    "strconv"
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
    if adder.Sum() != 10 {
        t.Errorf("Expected sum to be 10, got %d", adder.Sum())
    }
    adder.AddSimple(20)
    if adder.Sum() != 30 {
        t.Errorf("Expected sum to be 30, got %d", adder.Sum())
    }
    adder.Reset()
    if adder.Sum() != 0 {
        t.Errorf("Expected sum to be 0, got %d", adder.Sum())
    }
}

func Test_Int64Adder(t *testing.T) {
    adder := &Int64Adder{}
    goroutineNum := 500
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
            for j := 0; j < 1000; j++ {
                adder.Add(id, 1)
                if j%128 == 0 {
                    runtime.Gosched()
                }
            }
        }()
    }
    wg1.Wait()
    wg2.Done()
    wg.Wait()
    if adder.Sum() != 500000 {
        t.Errorf("Expected sum to be 500000, got %d", adder.Sum())
    }
    if adder.init == 0 {
        t.Errorf("Expected init to be 1, got %d", adder.init)
    }
    //if adder.base == 0 {
    //    t.Errorf("Expected base to be 0, got %d", adder.base)
    //}
}

func Benchmark1Int64Adder(b *testing.B) {
    //测试不同的并发场景下,goroutine切换速度不同时的性能对比
    for _, goroutineNum := range []int{100, 1000, 10000} {
        for _, i := range []int{4, 8, 32, 64, 128} {
            b.Run("Int64Adder_goroutine("+strconv.Itoa(goroutineNum)+")_"+strconv.Itoa(i), func(b *testing.B) {
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
            b.Run("Atomic_goroutine("+strconv.Itoa(goroutineNum)+")_"+strconv.Itoa(i), func(b *testing.B) {
                r := int64(0)
                wg := NewWaitGroup(goroutineNum)
                wg1 := NewWaitGroup(goroutineNum)
                wg2 := NewWaitGroup(1)
                //使用 waitGroup 保证所有的 goroutine 都已经启动
                for x := 0; x < goroutineNum; x++ {
                    go func() {
                        wg1.Done()
                        defer wg.Done()
                        wg2.Wait()
                        for j := 0; j < b.N; j++ {
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
}
