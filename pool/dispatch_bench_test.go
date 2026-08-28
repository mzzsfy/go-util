package pool

import (
	"runtime"
	"sync"
	"testing"
)

// BenchmarkGoPool_DispatchRoundtrip 测量单任务提交+执行往返延迟(无sleep)
func BenchmarkGoPool_DispatchRoundtrip(b *testing.B) {
	p := NewGoPool()
	// 预热: 确保 worker 已创建, 排除冷启动
	var wg sync.WaitGroup
	wg.Add(1)
	_ = p.Go(func() { wg.Done() })
	wg.Wait()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			wg.Add(1)
			_ = p.Go(func() { wg.Done() })
			wg.Wait()
		}
	})
}

// BenchmarkGo_DispatchRoundtrip 原生go对照
func BenchmarkGo_DispatchRoundtrip(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() { wg.Done() }()
			wg.Wait()
		}
	})
}

// BenchmarkGoPool_Throughput 纯吞吐: 并发提交+执行大量任务
func BenchmarkGoPool_Throughput(b *testing.B) {
	p := NewGoPool()
	_ = runtime.NumCPU
	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Go(func() { wg.Done() })
		}
	})
	wg.Wait()
}

// BenchmarkGo_Throughput 原生go吞吐对照
func BenchmarkGo_Throughput(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			go func() { wg.Done() }()
		}
	})
	wg.Wait()
}
