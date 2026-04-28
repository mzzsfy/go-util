package pool_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mzzsfy/go-util/pool"
)

// TestGoPool_CtxGo_ContextPropagation 验证 CtxGo 传递的 context 到达 panicHandler
func TestGoPool_CtxGo_ContextPropagation(t *testing.T) {
	t.Parallel()

	type ctxKey string
	const key ctxKey = "mykey"

	// 创建带自定义 key 的 context
	ctx := context.WithValue(context.Background(), key, "myval")

	// channel 用于接收 panicHandler 拿到的 context
	ctxCh := make(chan context.Context, 1)

	p := pool.NewGopool(
		pool.WithIdleTimeout(10*time.Millisecond),
		pool.WithPanicHandler(func(v any, handlerCtx context.Context) {
			select {
			case ctxCh <- handlerCtx:
			default:
			}
		}),
	)

	// 提交一个会 panic 的任务, 传递带 key 的 context
	err := p.CtxGo(ctx, func() {
		panic("test-panic-ctx")
	})
	if err != nil {
		t.Fatalf("CtxGo 提交失败: %v", err)
	}

	// 等待 panicHandler 被调用
	select {
	case handlerCtx := <-ctxCh:
		// 验证 handler 收到的 ctx 包含自定义 key
		val := handlerCtx.Value(key)
		if val != "myval" {
			t.Fatalf("期望 context value 'myval', 实际 %v", val)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("等待 panicHandler 超时")
	}

	p.Shutdown()
}

// TestGoPool_IdleTimeout_WorkerExit 验证空闲超时后 worker 退出
func TestGoPool_IdleTimeout_WorkerExit(t *testing.T) {
	t.Parallel()

	p := pool.NewGopool(pool.WithIdleTimeout(50 * time.Millisecond))

	// 提交一个任务并等待完成
	var wg sync.WaitGroup
	wg.Add(1)
	if err := p.Go(func() { wg.Done() }); err != nil {
		t.Fatalf("Go 提交失败: %v", err)
	}
	wg.Wait()

	// 等待空闲超时 (50ms) + 余量
	time.Sleep(150 * time.Millisecond)

	// 所有 worker 应已退出
	if wc := p.WorkerCount(); wc != 0 {
		t.Fatalf("期望 WorkerCount() == 0, 实际 %d", wc)
	}
}

// TestGoPool_Shutdown_SecondCallReturnsFalse 验证第二次 Shutdown 返回 false
func TestGoPool_Shutdown_SecondCallReturnsFalse(t *testing.T) {
	t.Parallel()

	p := pool.NewGopool()

	// 第一次 Shutdown 应返回 true
	if !p.Shutdown() {
		t.Fatal("第一次 Shutdown() 应返回 true")
	}

	// 第二次 Shutdown 应返回 false
	if p.Shutdown() {
		t.Fatal("第二次 Shutdown() 应返回 false")
	}
}

// TestGoPool_TaskCount 验证 TaskCount 反映队列积压
func TestGoPool_TaskCount(t *testing.T) {
	t.Parallel()

	p := pool.NewGopool(pool.WithMaxWorks(1), pool.WithIdleTimeout(50*time.Millisecond))

	// 用通道阻塞唯一的 worker
	block := make(chan struct{})

	var wg sync.WaitGroup
	// 第一个任务: 阻塞等待信号
	wg.Add(1)
	if err := p.Go(func() {
		<-block
		wg.Done()
	}); err != nil {
		t.Fatalf("提交阻塞任务失败: %v", err)
	}

	// 短暂等待 worker 启动并开始执行阻塞任务
	time.Sleep(20 * time.Millisecond)

	// 提交额外的排队任务
	const extraTasks = 10
	for i := 0; i < extraTasks; i++ {
		wg.Add(1)
		if err := p.Go(func() { wg.Done() }); err != nil {
			// 如果提交失败, 手动 Done 避免死锁
			wg.Done()
		}
	}

	// 短暂等待任务入队
	time.Sleep(10 * time.Millisecond)

	// 此时应有排队任务积压
	if tc := p.TaskCount(); tc == 0 {
		t.Fatal("期望 TaskCount() > 0, 实际 0")
	}

	// 释放阻塞任务, 等待全部完成
	close(block)
	wg.Wait()

	// 所有任务完成后队列应为空
	if tc := p.TaskCount(); tc != 0 {
		t.Fatalf("期望 TaskCount() == 0, 实际 %d", tc)
	}

	p.Shutdown()
}

// TestGoPool_WithMaxWorks_ZeroOrNegative 验证非法输入不改变默认值
func TestGoPool_WithMaxWorks_ZeroOrNegative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    int
	}{
		{name: "zero", n: 0},
		{name: "negative", n: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WithMaxWorks 传入 0 或负数时应保持默认 1024
			// 验证方式: 提交多个任务, 确保它们都能正常完成
			p := pool.NewGopool(pool.WithMaxWorks(tt.n), pool.WithIdleTimeout(50*time.Millisecond))

			const totalTasks = 20
			var executed int32
			var wg sync.WaitGroup
			wg.Add(totalTasks)
			for i := 0; i < totalTasks; i++ {
				if err := p.Go(func() {
					atomic.AddInt32(&executed, 1)
					wg.Done()
				}); err != nil {
					t.Fatalf("提交任务失败: %v", err)
				}
			}

			wg.Wait()
			if got := atomic.LoadInt32(&executed); got != totalTasks {
				t.Fatalf("期望执行 %d 个任务, 实际 %d", totalTasks, got)
			}

			p.Shutdown()
		})
	}
}
