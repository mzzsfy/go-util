package seq

import (
	"testing"
)

// When MapParallel只传顺序参数order=2,When 并发数缺省,Then 使用缺省并发数不panic且结果有序
func Test_MapParallel_Order2_DefaultConcurrency(t *testing.T) {
	t.Parallel()
	var r []int
	FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).
		MapParallel(func(v int) any { return v }, 2).
		ForEach(func(v any) { r = append(r, v.(int)) })
	if len(r) != 8 {
		t.Fatalf("元素数量错误: %d", len(r))
	}
	for i, v := range r {
		if v != i+1 {
			t.Fatalf("order=2顺序错误: idx=%d v=%d", i, v)
		}
	}
}

// When MapVParallel只传顺序参数order=2,When 并发数缺省,Then 使用缺省并发数不panic
func Test_MapVParallel_Order2_DefaultConcurrency(t *testing.T) {
	t.Parallel()
	count := 0
	BiFromMap(map[string]int{"a": 1, "b": 2, "c": 3}).
		MapVParallel(func(k string, v int) any { return v }, 2).
		ForEach(func(k string, v any) { count++ })
	if count != 3 {
		t.Fatalf("元素数量错误: %d", count)
	}
}

// When 多个异步回调并发panic,Then 错误收集无数据竞争,首个panic中止生产并重抛
func Test_MapParallelCustomize_ConcurrentPanicRace(t *testing.T) {
	t.Parallel()
	for round := 0; round < 200; round++ {
		panicked := false
		func() {
			defer func() {
				if recover() != nil {
					panicked = true
				}
			}()
			FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).
				MapParallelCustomize(func(v int, push func(any)) {
					go push(v)
				}).
				ForEach(func(any) {
					panic("boom")
				})
		}()
		if !panicked {
			t.Fatalf("round=%d 期望panic未发生", round)
		}
	}
}

// When 多个并行回调并发panic,Then 错误收集无数据竞争,首个panic中止生产并重抛
func Test_ParallelCustomize_ConcurrentPanicRace(t *testing.T) {
	t.Parallel()
	for round := 0; round < 200; round++ {
		panicked := false
		func() {
			defer func() {
				if recover() != nil {
					panicked = true
				}
			}()
			FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8}).
				ParallelCustomize(func(v int, f func()) {
					go f()
				}).
				ForEach(func(int) {
					panic("boom")
				})
		}()
		if !panicked {
			t.Fatalf("round=%d 期望panic未发生", round)
		}
	}
}
