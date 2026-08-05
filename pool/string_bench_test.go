package pool

import (
	"sync/atomic"
	"testing"
)

// BenchmarkStringPool_UseHit 单goroutine Use命中路径
func BenchmarkStringPool_UseHit(b *testing.B) {
	p := NewStringPool()
	s := "bench-key"
	p.Use(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Use(s)
	}
}

// BenchmarkStringPool_Peek 单goroutine Peek
func BenchmarkStringPool_Peek(b *testing.B) {
	p := NewStringPool()
	s := "bench-key"
	p.Use(s)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Peek(s)
	}
}

// BenchmarkStringPool_UnUse 单goroutine UnUse, 配对Use
func BenchmarkStringPool_UnUse(b *testing.B) {
	p := NewStringPool()
	s := "bench-key"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Use(s)
		p.UnUse(s)
	}
}

// BenchmarkStringPool_ConcurrentUseHit 并发Use命中
func BenchmarkStringPool_ConcurrentUseHit(b *testing.B) {
	p := NewStringPool()
	s := "bench-key"
	p.Use(s)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Use(s)
		}
	})
}

// BenchmarkStringPool_ConcurrentUseUnUse 并发Use+UnUse配对
func BenchmarkStringPool_ConcurrentUseUnUse(b *testing.B) {
	p := NewStringPool()
	s := "bench-key"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Use(s)
			p.UnUse(s)
		}
	})
}

// BenchmarkStringPool_PeekParallel 并发Peek
func BenchmarkStringPool_PeekParallel(b *testing.B) {
	p := NewStringPool()
	s := "bench-key"
	p.Use(s)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Peek(s)
		}
	})
}

// BenchmarkStringPool_ConcurrentDistinctKeys 并发不同key的Use/UnUse
func BenchmarkStringPool_ConcurrentDistinctKeys(b *testing.B) {
	p := NewStringPool()
	keys := make([]string, 256)
	for i := 0; i < 256; i++ {
		keys[i] = "key-" + string(rune('A'+i%26)) + string(rune('0'+i/26))
	}
	var idx int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := atomic.AddInt64(&idx, 1)
			k := keys[int(i%int64(len(keys)))]
			p.Use(k)
			p.UnUse(k)
		}
	})
	// 防止编译器优化
	_ = keys[0]
}
