package concurrent

import (
	"sync/atomic"
	"testing"
	"time"
)

// failThenDequeue 模拟无锁队列的 CAS 竞争失败: 前 fails 次 TryDequeue 返回 false,
// 此时底层队列始终非空
type failThenDequeue[T any] struct {
	Queue[T]
	fails int32
}

func (f *failThenDequeue[T]) TryDequeue() (T, bool) {
	if atomic.AddInt32(&f.fails, -1) >= 0 {
		var zero T
		return zero, false
	}
	return f.Queue.Dequeue()
}

// 非空队列上 TryDequeue 连续失败时, 超时出队必须继续重试而不是进入等待
func Test_BlockQueue_NonEmptyQueueKeepsRetrying(t *testing.T) {
	t.Parallel()
	base := newSegQueue[int]()
	base.Enqueue(7)
	// 前 spinTotalMax 次在自旋阶段失败, 之后 slow path 再失败 100 次
	q := BlockQueueWrapper[int](&failThenDequeue[int]{Queue: base, fails: spinTotalMax + 100})

	got, ok := q.DequeueBlock(2 * time.Second)
	if !ok || got != 7 {
		t.Fatalf("队列非空时应重试取到元素, got=%d ok=%v", got, ok)
	}
}

// 无超时出队同样不能把竞争失败当空队列, 否则永久阻塞
func Test_BlockQueue_ForeverDequeueNoLostWakeup(t *testing.T) {
	t.Parallel()
	base := newSegQueue[int]()
	base.Enqueue(9)
	q := BlockQueueWrapper[int](&failThenDequeue[int]{Queue: base, fails: spinTotalMax + 100})

	done := make(chan int, 1)
	go func() {
		v, _ := q.DequeueBlock()
		done <- v
	}()
	select {
	case v := <-done:
		if v != 9 {
			t.Fatalf("取到错误元素: %d", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("无超时出队永久阻塞: 队列非空但未继续重试")
	}
}

// 空队列上的超时出队必须由 timer 唤醒, 不能永久等待
func Test_BlockQueue_TimeoutFires(t *testing.T) {
	t.Parallel()
	q := BlockQueueWrapper[int](newSegQueue[int]())

	start := time.Now()
	done := make(chan bool, 1)
	go func() {
		_, ok := q.DequeueBlock(50 * time.Millisecond)
		done <- ok
	}()
	select {
	case ok := <-done:
		if ok {
			t.Fatal("空队列不应取到元素")
		}
		if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
			t.Fatalf("超时提前生效: %v", elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("出队超时未生效, timer 唤醒丢失")
	}
}
