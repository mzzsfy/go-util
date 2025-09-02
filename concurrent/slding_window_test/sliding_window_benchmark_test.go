package slding_window_test

import (
    "fmt"
    "testing"
    "time"

    c "github.com/mzzsfy/go-util/concurrent"
)

// 基准测试参数
var benchmarkParams = []struct {
    timeWindow   int64
    allowNumber  int32
    windowNumber int32
    name         string
}{
    {100, 100, 5, "默认配置"},
    {100, 1000, 10, "高吞吐量"},
    {50, 50, 5, "短时间窗口"},
    {500, 500, 10, "长时间窗口"},
    {100, 10, 10, "低速率"},
    {100, 10000, 10, "高速率"},
}

func BenchmarkSlidingWindow(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            sw := c.NewSlidingWindow(param.timeWindow, param.allowNumber, param.windowNumber)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                sw.CanDo()
            }
        })
    }
}

func BenchmarkSlidingWindow_RateLimiter(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            // Limiter使用每秒允许的事件数和桶大小
            limiter := NewLimiter(Limit(param.allowNumber), int(param.allowNumber))
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                limiter.Allow()
            }
        })
    }
}

func BenchmarkSlidingWindow_Parallel(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            sw := c.NewSlidingWindow(param.timeWindow, param.allowNumber, param.windowNumber)
            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    sw.CanDo()
                }
            })
        })
    }
}

func BenchmarkSlidingWindow_RateLimiter_Parallel(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            limiter := NewLimiter(Limit(param.allowNumber), int(param.allowNumber))
            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    limiter.Allow()
                }
            })
        })
    }
}

// 测试在限制范围内的处理能力
func BenchmarkSlidingWindow_WithinLimit(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            sw := c.NewSlidingWindow(param.timeWindow, param.allowNumber, param.windowNumber)

            // 预热
            for i := 0; i < 100; i++ {
                sw.CanDo()
            }

            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    sw.CanDo()
                }
            })
        })
    }
}

func BenchmarkSlidingWindow_WithinLimit_RateLimiter(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            limiter := NewLimiter(Limit(param.allowNumber), int(param.allowNumber))

            // 预热
            for i := 0; i < 100; i++ {
                limiter.Allow()
            }

            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    limiter.Allow()
                }
            })
        })
    }
}

// 测试在限制外的处理能力（令牌桶已满）
func BenchmarkSlidingWindow_BeyondLimit(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            sw := c.NewSlidingWindow(param.timeWindow, param.allowNumber, param.windowNumber)

            // 先填满窗口
            for i := int32(0); i < param.windowNumber; i++ {
                for j := int32(0); j < param.allowNumber; j++ {
                    sw.CanDo()
                }
                // 等待窗口滑动
                time.Sleep(time.Duration(10+param.timeWindow/int64(param.windowNumber)) * time.Millisecond)
            }

            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    sw.CanDo()
                }
            })
        })
    }
}

func BenchmarkSlidingWindow_BeyondLimit_RateLimiter(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            limiter := NewLimiter(Limit(param.allowNumber), int(param.allowNumber))

            // 先消耗完所有令牌并稍微超出
            for i := 0; i < int(param.allowNumber)+20; i++ {
                limiter.Allow()
            }

            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    limiter.Allow()
                }
            })
        })
    }
}

// 测试在高并发情况下的性能
func BenchmarkSlidingWindow_HighConcurrency(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            // 使用更高的速率进行高并发测试
            sw := c.NewSlidingWindow(param.timeWindow, param.allowNumber*10, param.windowNumber)

            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    sw.CanDo()
                }
            })
        })
    }
}

func BenchmarkSlidingWindow_HighConcurrency_RateLimiter(b *testing.B) {
    for _, param := range benchmarkParams {
        b.Run(param.name, func(b *testing.B) {
            // 使用更高的速率进行高并发测试
            limiter := NewLimiter(Limit(param.allowNumber*10), int(param.allowNumber*10))

            b.ResetTimer()
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    limiter.Allow()
                }
            })
        })
    }
}

// 复杂度分析测试
func BenchmarkComplexityAnalysis(b *testing.B) {
    // 测试不同窗口数量对性能的影响
    windowNumbers := []int32{3, 5, 10, 20, 50}

    for _, wn := range windowNumbers {
        b.Run(fmt.Sprintf("窗口数量_%d", wn), func(b *testing.B) {
            sw := c.NewSlidingWindow(1000, 1000, wn)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                sw.CanDo()
            }
        })
    }
}
