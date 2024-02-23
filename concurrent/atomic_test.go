package concurrent

import (
    "github.com/mzzsfy/go-util/helper"
    "runtime"
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
    wg := helper.NewWaitGroup(100)
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

//go test -bench=Benchmark1.+ -count=3 ./concurrent
func Benchmark1Int64Adder4(b *testing.B) {
    testInt64Adder(8, b)
}

func Benchmark1Atomic4(b *testing.B) {
    testAtomic(8, b)
}

func Benchmark1Int64Adder32(b *testing.B) {
    testInt64Adder(32, b)
}

func Benchmark1Atomic32(b *testing.B) {
    testAtomic(32, b)
}

func Benchmark1Int64Adder128(b *testing.B) {
    testInt64Adder(128, b)
}

func Benchmark1Atomic128(b *testing.B) {
    testAtomic(128, b)
}

func testInt64Adder(interval int, b *testing.B) {
    adder := &Int64Adder{}
    adder.AddSimple(1)
    adder.Reset()
    wg := helper.NewWaitGroup(b.N)
    wg1 := helper.NewWaitGroup(b.N)
    wg2 := helper.NewWaitGroup(1)
    for i := 0; i < b.N; i++ {
        go func() {
            wg1.Done()
            defer wg.Done()
            wg2.Wait()
            id := GoID()
            for i := 0; i < 1000; i++ {
                adder.Increment(id)
                if i%interval == 0 {
                    runtime.Gosched()
                }
            }
        }()
    }
    wg1.Wait()
    b.ResetTimer()
    wg2.Done()
    wg.Wait()
}

func testAtomic(interval int, b *testing.B) {
    r := int64(0)
    wg := helper.NewWaitGroup(b.N)
    wg1 := helper.NewWaitGroup(b.N)
    wg2 := helper.NewWaitGroup(1)
    for i := 0; i < b.N; i++ {
        go func() {
            wg1.Done()
            defer wg.Done()
            wg2.Wait()
            for i := 0; i < 1000; i++ {
                atomic.AddInt64(&r, 1)
                if i%interval == 0 {
                    runtime.Gosched()
                }
            }
        }()
    }
    wg1.Wait()
    b.ResetTimer()
    wg2.Done()
    wg.Wait()
}
