package helper

import (
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkTimerCell_ProcessLayerTasks 槽位取出-清空路径微观基准
// processLayerTasks 将槽 slice 置 nil 后, 下次填充需重新扩容分配
const processBenchTasksPerSlot = 8

func BenchmarkTimerCell_ProcessLayerTasks(b *testing.B) {
	w := NewTimerWheel()
	defer w.Stop()
	// 同步执行器消除 goroutine 分配噪声
	w.executor = func(t Task) { t.Run() }
	layer := w.layers[0]
	cell := &layer.cells[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cell.mu.Lock()
		for j := 0; j < processBenchTasksPerSlot; j++ {
			cell.tasks = append(cell.tasks, &task{taskID: int32(j + 1), wheel: w})
		}
		cell.mu.Unlock()
		w.processLayerTasks(layer, 0, true)
	}
	b.StopTimer()
}

// BenchmarkTimerWheel_RepeatingExpire 重复任务到期执行吞吐
// 每 tick 全量到期并重调度, 槽 slice 反复填充-清空, allocs/op 反映槽 slice 复用收益
func BenchmarkTimerWheel_RepeatingExpire(b *testing.B) {
	w := NewTimerWheel(WithTickInterval(5 * time.Millisecond))
	defer w.Stop()
	var count int32
	w.executor = func(t Task) {
		t.Run()
		atomic.AddInt32(&count, 1)
	}
	// worker 数须保证单 tick 内可消化 b.N 量级
	const workers = 64
	for i := 0; i < workers; i++ {
		w.ScheduleRepeating(w.tickInterval, FuncTask(func() {}))
	}
	// 预热一轮
	time.Sleep(w.tickInterval * 3)
	atomic.StoreInt32(&count, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for atomic.LoadInt32(&count) < int32(b.N) {
		time.Sleep(time.Millisecond)
	}
	b.StopTimer()
}
