package pool

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestObjectPool_GetPut_Basic 基本 Get/Put 循环, 验证对象复用
func Test_ObjectPool_GetPut_Basic(t *testing.T) {
	t.Parallel()

	var resetCalls int32
	pool := NewObjectPool[int](
		func() *int { v := 0; return &v },
		func(i *int) { atomic.AddInt32(&resetCalls, 1); *i = 0 },
	)

	// 首次 Get 应创建新对象
	p := pool.Get()
	if p == nil {
		t.Fatal("Get 不应返回 nil")
	}
	*p = 42

	// Put 应调用 reset
	pool.Put(p)
	if atomic.LoadInt32(&resetCalls) != 1 {
		t.Fatalf("Put 后 reset 应被调用 1 次, 实际 %d 次", resetCalls)
	}

	// 再次 Get 可能复用同一对象, reset 已将其清零
	p2 := pool.Get()
	if *p2 != 0 {
		t.Fatalf("复用对象应已被 reset 清零, got %d", *p2)
	}
	pool.Put(p2)
}

// TestObjectPool_ResetCalled 验证每次 Put 都调用 reset 函数
func Test_ObjectPool_ResetCalled(t *testing.T) {
	t.Parallel()

	type item struct {
		Value int
		Flag  bool
	}

	var resetCount int32
	pool := NewObjectPool[item](
		func() *item { return &item{} },
		func(i *item) {
			atomic.AddInt32(&resetCount, 1)
			i.Value = 0
			i.Flag = false
		},
	)

	const rounds = 10
	for i := 0; i < rounds; i++ {
		obj := pool.Get()
		obj.Value = i
		obj.Flag = true
		pool.Put(obj)
	}

	if got := atomic.LoadInt32(&resetCount); got != rounds {
		t.Fatalf("reset 应被调用 %d 次, 实际 %d 次", rounds, got)
	}
}

// TestObjectPool_Concurrent 并发场景下 Get/Put 安全无竞争
func TestObjectPool_Concurrent(t *testing.T) {
	t.Parallel()

	pool := NewObjectPool[int](
		func() *int { v := 0; return &v },
		func(i *int) { *i = 0 },
	)

	const workers = 50
	const opsPerWorker = 200
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				p := pool.Get()
				*p = j
				pool.Put(p)
			}
		}()
	}

	wg.Wait()
}

// TestObjectPool_MultipleTypes 验证泛型对不同类型都能正常工作
func Test_ObjectPool_MultipleTypes(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		pool := NewObjectPool[string](
			func() *string { s := ""; return &s },
			func(s *string) { *s = "" },
		)
		p := pool.Get()
		*p = "hello"
		pool.Put(p)
		p2 := pool.Get()
		if *p2 != "" {
			t.Fatalf("reset 后应为空字符串, got %q", *p2)
		}
		pool.Put(p2)
	})

	t.Run("slice", func(t *testing.T) {
		t.Parallel()
		pool := NewObjectPool[[]byte](
			func() *[]byte { s := make([]byte, 0, 16); return &s },
			func(s *[]byte) { *s = (*s)[:0] },
		)
		p := pool.Get()
		*p = append(*p, 1, 2, 3)
		if len(*p) != 3 {
			t.Fatalf("append 后长度应为 3, got %d", len(*p))
		}
		pool.Put(p)
		p2 := pool.Get()
		if len(*p2) != 0 {
			t.Fatalf("reset 后长度应为 0, got %d", len(*p2))
		}
		pool.Put(p2)
	})
}
