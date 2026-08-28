package concurrent

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Given delayQueue 短延迟
// When 入队后立即出队
// Then 未到期返回 false, 到期后可取到元素
func Test_DelayQueue_NotReadyThenFired(t *testing.T) {
	t.Parallel()
	const delayTime = 50 * time.Millisecond
	q := NewQueue[int](WithTypeDelay[int](delayTime))

	q.Enqueue(1)
	if _, ok := q.Dequeue(); ok {
		t.Fatal("未到期时出队不应成功")
	}

	deadline := time.Now().Add(waitTimeout)
	for {
		if v, ok := q.Dequeue(); ok {
			if v != 1 {
				t.Fatalf("value=%d want=1", v)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("到期后仍未出队")
		}
		time.Sleep(time.Millisecond)
	}
}

// Given delayQueue 多元素
// When 等待超过最大延迟
// Then 全部元素最终可取到(不断言到期序, 存在单消费者串行局限)
func Test_DelayQueue_AllElementsEventually(t *testing.T) {
	t.Parallel()
	q := NewQueue[int](WithTypeDelay[int](30 * time.Millisecond))

	const n = 5
	for i := 0; i < n; i++ {
		q.Enqueue(i)
	}

	seen := make(map[int]bool)
	deadline := time.Now().Add(waitTimeout)
	for len(seen) < n {
		if v, ok := q.Dequeue(); ok {
			if v < 0 || v >= n {
				t.Fatalf("unexpected value %d", v)
			}
			seen[v] = true
			continue
		}
		if time.Now().After(deadline) {
			t.Fatalf("等待超时, 只收到 %v", seen)
		}
		time.Sleep(time.Millisecond)
	}
	if q.Size() != 0 {
		t.Fatalf("size=%d want=0", q.Size())
	}
}

// Given delayQueue 消费者已退出窗口
// When 退出窗口期入队
// Then 双检重启接管, 元素仍能到期被取到
func Test_DelayQueue_EnqueueAfterConsumerIdle(t *testing.T) {
	t.Parallel()
	q := NewQueue[int](WithTypeDelay[int](10 * time.Millisecond))

	// 第一轮: 完整走完消费周期, 消费者退出
	q.Enqueue(1)
	deadline := time.Now().Add(waitTimeout)
	for {
		if _, ok := q.Dequeue(); ok {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("第一轮元素未到期出队")
		}
		time.Sleep(time.Millisecond)
	}
	// 等待消费者确认退出(seg 队列空后 consume 返回)
	for i := 0; i < 100 && q.Size() == 0; i++ {
		time.Sleep(time.Millisecond)
	}

	// 第二轮: 消费者已退出, 入队应触发重启
	q.Enqueue(2)
	deadline = time.Now().Add(waitTimeout)
	for {
		if v, ok := q.Dequeue(); ok {
			if v != 2 {
				t.Fatalf("value=%d want=2", v)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("重启后元素未到期出队")
		}
		time.Sleep(time.Millisecond)
	}
}

// --- BDD: sliding_window allowNumber 合法边界 ---

// Given allowNumber 等于 windowNumber
// When 创建
// Then 不返回错误, 每窗口限流为分配后的值
func Test_SlidingWindow_AllowNumberEqualsWindowNumber(t *testing.T) {
	t.Parallel()
	sw, err := NewSlidingWindow(3000, 3, 3)
	if err != nil {
		t.Fatalf("allowNumber==windowNumber 应合法: %v", err)
	}
	if sw == nil {
		t.Fatal("sw should not be nil")
	}
}

// Given 固定时钟注入
// When 持续请求
// Then 配额有限放行, 时钟不动不放行, 推进超过整窗后恢复放行
func Test_SlidingWindow_FixedClockLimitAndRecovery(t *testing.T) {
	t.Parallel()
	const windowTime = int64(1000)
	const allowNumber = 10
	const windowNumber = 5

	// 固定时钟以构造时刻为基准, 与分片时间戳的初始化对齐
	now := time.Now().UnixMilli()
	sw, err := NewSlidingWindow(windowTime, allowNumber, windowNumber)
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	sw.Now = func() int64 { return now }

	// 阶段A: 时钟固定, 放行应为有限个
	allowedA := 0
	for i := 0; i < 200; i++ {
		if sw.CanDo() {
			allowedA++
		}
	}
	if allowedA == 0 {
		t.Fatal("初始应放行至少一个请求")
	}
	if allowedA >= allowNumber {
		t.Fatalf("时钟不动时放行数=%d 应小于总配额%d", allowedA, allowNumber)
	}

	// 阶段B: 时钟不动, 配额耗尽后持续拒绝
	for i := 0; i < 100; i++ {
		if sw.CanDo() {
			t.Fatal("时钟未推进不应恢复放行")
		}
	}

	// 阶段C: 推进超过整窗, 所有分片过期, 恢复放行
	now += windowTime * 2
	allowedC := 0
	for i := 0; i < 200; i++ {
		if sw.CanDo() {
			allowedC++
		}
	}
	if allowedC == 0 {
		t.Fatal("窗口完全过期后应恢复放行")
	}
}

// --- BDD: 治理 Test_DequeueTimeout(原零断言+泄漏goroutine) ---

// Given 空队列
// When 生产者延迟入队, 消费者带超时出队
// Then 首次超时返回 false, 后续命中返回元素, 无参数版阻塞命中
func Test_DequeueTimeout_Semantics(t *testing.T) {
	t.Parallel()
	q := BlockQueueWrapper[int](newSegQueue[int]())

	// 空队列: 超时过期
	if v, ok := q.DequeueBlock(50 * time.Millisecond); ok {
		t.Fatalf("空队列超时出队不应成功, 得到 %v", v)
	}

	// 生产者延迟入队, 超时窗口大于延迟, 必命中
	go func() {
		time.Sleep(10 * time.Millisecond)
		q.Enqueue(1)
	}()
	got, ok := q.DequeueBlock(waitTimeout)
	if !ok || got != 1 {
		t.Fatalf("DequeueBlock=%v,%v want 1,true", got, ok)
	}

	// 无参数阻塞版
	go func() {
		time.Sleep(10 * time.Millisecond)
		q.Enqueue(2)
	}()
	got, ok = q.DequeueBlock()
	if !ok || got != 2 {
		t.Fatalf("DequeueBlock=%v,%v want 2,true", got, ok)
	}
}

// --- BDD: 治理 Test_LkQueue(原零断言) ---

// Given 多种队列实现
// When 串行入队出队
// Then 数量守恒且值保序
func Test_Queue_SerialIntegrity(t *testing.T) {
	t.Parallel()
	for _, m := range tryDequeueMakers {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			q := m.make()
			n := m.serialCap
			for i := 0; i < n; i++ {
				q.Enqueue(i)
			}
			if q.Size() != n {
				t.Fatalf("size=%d want=%d", q.Size(), n)
			}
			for want := 0; want < n; want++ {
				got, ok := q.Dequeue()
				if !ok || got != want {
					t.Fatalf("dequeue=%v,%v want %d,true", got, ok, want)
				}
			}
			if q.Size() != 0 {
				t.Fatalf("size=%d want=0", q.Size())
			}
		})
	}
}

// Given 并发生产消费
// When 多种队列实现
// Then 总量守恒, 不丢不重
func Test_Queue_ConcurrentIntegrity(t *testing.T) {
	t.Parallel()
	for _, m := range tryDequeueMakers {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			const producers, perProducer = 4, 500
			const total = producers * perProducer
			q := m.make()

			var consumed int64
			var wg sync.WaitGroup
			// 单消费者循环出队, 空转让出
			wg.Add(1)
			go func() {
				defer wg.Done()
				for atomic.LoadInt64(&consumed) < total {
					if _, ok := q.Dequeue(); ok {
						atomic.AddInt64(&consumed, 1)
					} else {
						time.Sleep(time.Microsecond)
					}
				}
			}()
			for p := 0; p < producers; p++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < perProducer; i++ {
						q.Enqueue(1)
					}
				}()
			}
			// 消费完成信号兜底, 防消费者卡死挂起测试
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(30 * time.Second):
				t.Fatalf("并发吞吐超时: consumed=%d want=%d", atomic.LoadInt64(&consumed), total)
			}
			if got := atomic.LoadInt64(&consumed); got != total {
				t.Fatalf("consumed=%d want=%d", got, total)
			}
		})
	}
}

// Given errors.Is 哨兵
// When allowNumber 为边界值
// Then 边界恰好合法, 越界返回对应哨兵错误
func Test_SlidingWindow_AllowNumberBoundary(t *testing.T) {
	t.Parallel()
	// 恰好等于窗口数: 合法
	if _, err := NewSlidingWindow(3000, 3, 3); err != nil {
		t.Fatalf("边界值应合法: %v", err)
	}
	// 小于窗口数: 哨兵错误
	_, err := NewSlidingWindow(3000, 2, 3)
	if !errors.Is(err, ErrAllowNumber) {
		t.Fatalf("allowNumber<windowNumber 应返回 ErrAllowNumber, 得到 %v", err)
	}
}
