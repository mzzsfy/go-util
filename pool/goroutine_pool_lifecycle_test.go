package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Given 未设置 panicHandler
// When 提交会 panic 的任务
// Then panic 被吞掉, worker 存活, 后续任务正常执行
func TestGoPool_PanicWithoutHandler(t *testing.T) {
	t.Parallel()

	p := NewGoPool(WithIdleTimeout(50 * time.Millisecond))
	if err := p.Go(func() { panic("no-handler") }); err != nil {
		t.Fatalf("submit panic task failed: %v", err)
	}

	// panic 任务执行完毕后提交正常任务, 事件同步等待
	done := make(chan struct{})
	if err := p.Go(func() { close(done) }); err != nil {
		t.Fatalf("submit normal task failed: %v", err)
	}
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("panic 后 worker 应存活, 正常任务应被执行")
	}

	if !p.Shutdown() {
		t.Fatal("shutdown should succeed")
	}
}

// Given WithIdleTimeout 传入非法值
// When 创建并使用池
// Then 非法值被忽略(保持默认), 池正常工作
func TestGoPool_WithIdleTimeout_InvalidIgnored(t *testing.T) {
	t.Parallel()

	p := NewGoPool(WithIdleTimeout(0), WithIdleTimeout(-1*time.Second))
	var count int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		if err := p.Go(func() { atomic.AddInt32(&count, 1); wg.Done() }); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}
	wg.Wait()
	if got := atomic.LoadInt32(&count); got != 10 {
		t.Fatalf("executed=%d want=10", got)
	}
	if !p.Shutdown() {
		t.Fatal("shutdown should succeed")
	}
}

// Given 池经历完整生命周期
// When Shutdown 后提交被拒, Restart 后再次提交
// Then 各阶段语义正确, 任务全部执行
func TestGoPool_FullLifecycle_Reuse(t *testing.T) {
	t.Parallel()

	p := NewGoPool(WithName("lifecycle"))

	// 阶段1: 正常提交
	var phase1 int32
	var wg1 sync.WaitGroup
	wg1.Add(5)
	for i := 0; i < 5; i++ {
		if err := p.Go(func() { atomic.AddInt32(&phase1, 1); wg1.Done() }); err != nil {
			t.Fatalf("phase1 submit failed: %v", err)
		}
	}
	wg1.Wait()
	if got := atomic.LoadInt32(&phase1); got != 5 {
		t.Fatalf("phase1 executed=%d want=5", got)
	}

	// 阶段2: Shutdown 排空后拒绝新任务
	if !p.Shutdown() {
		t.Fatal("first shutdown should succeed")
	}
	if err := p.CtxGo(context.Background(), func() {}); err != ErrPoolClosed {
		t.Fatalf("after shutdown expect ErrPoolClosed, got %v", err)
	}
	if p.Shutdown() {
		t.Fatal("second shutdown should return false")
	}

	// 阶段3: Restart 复活
	if !p.Restart() {
		t.Fatal("restart should succeed after shutdown")
	}

	// 阶段4: 复活后提交正常执行
	var phase4 int32
	var wg4 sync.WaitGroup
	wg4.Add(3)
	for i := 0; i < 3; i++ {
		if err := p.Go(func() { atomic.AddInt32(&phase4, 1); wg4.Done() }); err != nil {
			t.Fatalf("phase4 submit failed: %v", err)
		}
	}
	wg4.Wait()
	if got := atomic.LoadInt32(&phase4); got != 3 {
		t.Fatalf("phase4 executed=%d want=3", got)
	}

	if !p.Shutdown() {
		t.Fatal("final shutdown should succeed")
	}
}

// Given 运行中调用 Restart
// When 池未关闭
// Then 返回 false 且池继续可用
func TestGoPool_RestartWhileRunning_NoSideEffect(t *testing.T) {
	t.Parallel()

	p := NewGoPool()
	if p.Restart() {
		t.Fatal("restart on running pool should return false")
	}

	done := make(chan struct{})
	if err := p.Go(func() { close(done) }); err != nil {
		t.Fatalf("submit after failed restart failed: %v", err)
	}
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("task should complete after failed restart")
	}
	if !p.Shutdown() {
		t.Fatal("shutdown should succeed")
	}
}
