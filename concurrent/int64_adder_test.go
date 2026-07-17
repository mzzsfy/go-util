package concurrent

import (
    "math"
    "runtime"
    "strconv"
    "sync"
    "sync/atomic"
    "testing"
    "time"
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
    // 减少并发数以提高CI稳定性
    goroutineNum := 100
    wg := NewWaitGroup(goroutineNum)
    wg1 := NewWaitGroup(goroutineNum)
    wg2 := NewWaitGroup(1)
    // 使用 channel 添加超时机制
    done := make(chan struct{})
    go func() {
        // 使用 waitGroup 保证所有的 goroutine 都已经启动
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
        close(done)
    }()
    // 添加整体超时机制
    select {
    case <-done:
    case <-time.After(30 * time.Second):
        t.Fatal("测试超时，可能存在死锁或调度延迟")
    }
    if adder.Sum() != 100000 {
        t.Errorf("Expected sum to be 100000, got %d", adder.Sum())
    }
}

func Benchmark1Int64Adder(b *testing.B) {
    // 测试不同的并发场景下,goroutine切换速度不同时的性能对比
    // 减少并发数以提高CI稳定性
    for _, goroutineNum := range []int{50, 100, 200} {
        for _, i := range []int{4, 8, 32, 64, 128} {
            b.Run("Int64Adder_goroutine("+strconv.Itoa(goroutineNum)+")_"+strconv.Itoa(i), func(b *testing.B) {
                adder := &Int64Adder{}
                wg := NewWaitGroup(goroutineNum)
                wg1 := NewWaitGroup(goroutineNum)
                wg2 := NewWaitGroup(1)
                // 使用 waitGroup 保证所有的 goroutine 都已经启动
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
                // 使用 waitGroup 保证所有的 goroutine 都已经启动
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

// Test_Int64Adder_Overflow 测试Int64Adder溢出行为
func Test_Int64Adder_Overflow(t *testing.T) {
    t.Parallel()
    // 测试接近int64最大值时的加法行为
    adder := &Int64Adder{}
    // 先设置为接近最大值
    adder.AddSimple(math.MaxInt64 - 10)
    if adder.Sum() != math.MaxInt64-10 {
        t.Fatalf("初始值应为 %d, 实际 %d", math.MaxInt64-10, adder.Sum())
    }

    // 添加一个值使其溢出
    adder.AddSimple(20)
    // 溢出后应该是负数(int64回绕)
    sum := adder.Sum()
    if sum >= 0 {
        t.Errorf("溢出后应该是负数, 实际 %d", sum)
    }
    // 验证回绕的正确性
    // MaxInt64-10+20 = MaxInt64+10 会溢出回绕到 MinInt64+9
    expected := int64(math.MinInt64) + 9
    if sum != expected {
        t.Errorf("溢出值应为 %d, 实际 %d", expected, sum)
    }
}

// Test_Int64Adder_Underflow 测试Int64Adder下溢行为
func Test_Int64Adder_Underflow(t *testing.T) {
    t.Parallel()
    adder := &Int64Adder{}
    // 从接近最小值开始
    adder.AddSimple(math.MinInt64 + 10)
    if adder.Sum() != math.MinInt64+10 {
        t.Fatalf("初始值应为 %d, 实际 %d", math.MinInt64+10, adder.Sum())
    }

    // 减去一个值使其下溢
    adder.AddSimple(-20)
    sum := adder.Sum()
    // 下溢后应该是正数(int64回绕)
    if sum <= 0 {
        t.Errorf("下溢后应该是正数, 实际 %d", sum)
    }
    // MinInt64+10-20 = MinInt64-10 会下溢回绕到 MaxInt64-9
    expected := int64(math.MaxInt64) - 9
    if sum != expected {
        t.Errorf("下溢值应为 %d, 实际 %d", expected, sum)
    }
}

// Test_Int64Adder_MaxBoundary 测试Int64Adder在边界值附近的行为
func Test_Int64Adder_MaxBoundary(t *testing.T) {
    t.Parallel()
    // 测试到达最大值后再减去值的行为
    adder := &Int64Adder{}
    adder.AddSimple(math.MaxInt64)
    if adder.Sum() != math.MaxInt64 {
        t.Fatalf("值应为 %d, 实际 %d", math.MaxInt64, adder.Sum())
    }

    // 从最大值减去1
    adder.AddSimple(-1)
    if adder.Sum() != math.MaxInt64-1 {
        t.Errorf("减1后应为 %d, 实际 %d", math.MaxInt64-1, adder.Sum())
    }

    // 再加回去
    adder.AddSimple(1)
    if adder.Sum() != math.MaxInt64 {
        t.Errorf("加1后应为 %d, 实际 %d", math.MaxInt64, adder.Sum())
    }
}

// Test_Int64Adder_MinBoundary 测试Int64Adder在最小值附近的行为
func Test_Int64Adder_MinBoundary(t *testing.T) {
    t.Parallel()
    adder := &Int64Adder{}
    adder.AddSimple(math.MinInt64)
    if adder.Sum() != math.MinInt64 {
        t.Fatalf("值应为 %d, 实际 %d", math.MinInt64, adder.Sum())
    }

    // 从最小值加1
    adder.AddSimple(1)
    if adder.Sum() != math.MinInt64+1 {
        t.Errorf("加1后应为 %d, 实际 %d", math.MinInt64+1, adder.Sum())
    }

    // 再减回去
    adder.AddSimple(-1)
    if adder.Sum() != math.MinInt64 {
        t.Errorf("减1后应为 %d, 实际 %d", math.MinInt64, adder.Sum())
    }
}

// Test_Int64Adder_ResetAfterOverflow 测试溢出后重置的行为
func Test_Int64Adder_ResetAfterOverflow(t *testing.T) {
    t.Parallel()
    adder := &Int64Adder{}
    // 制造溢出
    adder.AddSimple(math.MaxInt64)
    adder.AddSimple(100)
    // 确认已溢出
    if adder.Sum() >= 0 {
        t.Fatalf("应该已溢出为负数, 实际 %d", adder.Sum())
    }

    // 重置
    adder.Reset()
    if adder.Sum() != 0 {
        t.Errorf("重置后应为0, 实际 %d", adder.Sum())
    }

    // 重置后正常使用
    adder.AddSimple(100)
    if adder.Sum() != 100 {
        t.Errorf("重置后添加值应为100, 实际 %d", adder.Sum())
    }
}

// Test_Int64Adder_InitRace 反复触发首次并发初始化窗口
// 修复前: ARM 等弱内存模型下 values 可能 nil panic, 或 Sum 漏分片数据
// 修复后: 通过 atomic 三态发布, 保证 values 可见性与数据完整
func Test_Int64Adder_InitRace(t *testing.T) {
    t.Parallel()
    const iters = 200
    const goroutines = 8
    const addsPerG = 100
    want := int64(goroutines * addsPerG)
    for n := 0; n < iters; n++ {
        adder := &Int64Adder{}
        // 多个 goroutine 同时首次 Add, 命中 init 0->1->2 的发布窗口
        start := make(chan struct{})
        wg := sync.WaitGroup{}
        for g := 0; g < goroutines; g++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                id := GoID()
                <-start
                for i := 0; i < addsPerG; i++ {
                    adder.Add(id, 1)
                }
            }()
        }
        close(start)
        wg.Wait()
        if got := adder.Sum(); got != want {
            t.Fatalf("iter %d: sum=%d want=%d (漏数据, values 发布存在 race)", n, got, want)
        }
    }
}
