package seq

import (
	"sync"
	"testing"
)

// Test_Seq_OnLastRace OnLast位于Parallel之前,last写入在发射方goroutine,无数据竞争
// 并行分支内记录last(OnLast在Parallel之后)为禁止组合,见OnLast文档
func Test_Seq_OnLastRace(t *testing.T) {
	for i := 0; i < 50; i++ {
		FromSlice(benchSlice).OnLast(func(last *int) {
			if last == nil || *last != len(benchSlice)-1 {
				t.Error("last值不正确")
			}
		}).Parallel(8).ForEach(func(int) {})
	}
}

// Test_Seq_OnEachNXRace OnEachNX位于Parallel之前,与OnLast同理
func Test_Seq_OnEachNXRace(t *testing.T) {
	for i := 0; i < 50; i++ {
		FromSlice(benchSlice).OnEachNX(16, func(int) {}).Parallel(8).ForEach(func(int) {})
	}
}

// Test_Seq_ParallelFinally 并行流的收尾钩子使用Finally,在所有任务完成后于调用方goroutine执行
func Test_Seq_ParallelFinally(t *testing.T) {
	for i := 0; i < 50; i++ {
		var mu sync.Mutex
		total := 0
		FromSlice(benchSlice).Parallel(8).Finally(func() {
			mu.Lock()
			defer mu.Unlock()
			if total != len(benchSlice) {
				t.Error("Finally时元素未消费完")
			}
		}).ForEach(func(int) {
			mu.Lock()
			total++
			mu.Unlock()
		})
	}
}
