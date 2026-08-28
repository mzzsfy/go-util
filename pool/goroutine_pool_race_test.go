package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestGoPool_Shutdown_RejectsLateDispatch 验证 Shutdown 完成后 dispatch 不得再入队任务
// 复现缺陷: CtxGo 检查 shutDown==0 后被调度走, Shutdown 完成(wg.Wait+drainQueue)后
// 该 goroutine 恢复并调用 dispatch 入队 -> 任务滞留队列, 无 worker 消费, 任务永久丢失
func TestGoPool_Shutdown_RejectsLateDispatch(t *testing.T) {
	t.Parallel()
	p := NewGoPool(WithIdleTimeout(10 * time.Millisecond))
	if !p.Shutdown() {
		t.Fatal("shutdown should succeed")
	}
	var executed int32
	// 直接调用 dispatch 模拟"检查已通过、入队晚到"的竞态任务
	p.dispatch(&task{ctx: context.Background(), fn: func() { atomic.AddInt32(&executed, 1) }})
	if tc := p.TaskCount(); tc != 0 {
		t.Fatalf("Shutdown 后队列不应滞留任务, got %d", tc)
	}
	if atomic.LoadInt32(&executed) != 0 {
		t.Fatalf("Shutdown 后不应执行新提交的任务, got %d", executed)
	}
}

// TestGoPool_TimeoutExit_NoTaskStuck 压测 worker 空闲超时退出与提交的竞争窗口:
// dispatch 读 WaiterCount()>0 时不建新 worker, 但 worker 可能已越过 deadline 即将退出
// (waiter 未减 / works 未减), check-then-act 窗口导致任务滞留队列无人消费
func TestGoPool_TimeoutExit_NoTaskStuck(t *testing.T) {
	t.Parallel()
	for round := 0; round < 200; round++ {
		p := NewGoPool(WithMaxWorkers(1), WithIdleTimeout(time.Millisecond))
		done := make(chan struct{})
		go func() {
			defer close(done)
			var wg sync.WaitGroup
			for j := 0; j < 5; j++ {
				wg.Add(1)
				if err := p.Go(func() { wg.Done() }); err != nil {
					wg.Done()
				}
				// 间隔大于空闲超时, 迫使 worker 在两次提交间超时退出
				time.Sleep(1500 * time.Microsecond)
			}
			wg.Wait()
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatalf("round %d: 任务滞留队列, worker 超时退出窗口吞任务", round)
		}
		p.Shutdown()
	}
}

// TestGoPool_Shutdown_DrainAcceptedTasks 验证 Shutdown 语义为排空:
// 并发提交与 Shutdown 竞争时, 所有被接受(返回 nil)的任务必须全部执行
func TestGoPool_Shutdown_DrainAcceptedTasks(t *testing.T) {
	t.Parallel()
	for round := 0; round < 20; round++ {
		p := NewGoPool(WithMaxWorkers(4), WithIdleTimeout(5*time.Millisecond))
		const n = 64
		var (
			submitted int32
			executed  int32
			mu        sync.Mutex
			accepted  []int
		)
		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			i := i
			go func() {
				err := p.Go(func() {
					atomic.AddInt32(&executed, 1)
					wg.Done()
				})
				if err == nil {
					atomic.AddInt32(&submitted, 1)
					mu.Lock()
					accepted = append(accepted, i)
					mu.Unlock()
				} else {
					wg.Done()
				}
			}()
		}
		p.Shutdown()
		wg.Wait()
		if got := atomic.LoadInt32(&executed); got != atomic.LoadInt32(&submitted) {
			t.Fatalf("round %d: 已接受任务丢失, submitted=%d executed=%d", round, submitted, executed)
		}
	}
}
