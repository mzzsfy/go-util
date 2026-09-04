package concurrent

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func zipAny(vars ...any) []any {
	return vars
}

// Given delayQueue
// When 未到期时出队
// Then 出队成功时刻距入队至少 delay, 到期后轮询可取到全部元素(不断言到期序)
func TestDelayQueue_Dequeue(t *testing.T) {
	t.Parallel()
	// 判据锚定出队成功时刻: 入队与探测之间被调度抢占跨过到期时刻属正确行为
	const delayTime = 100 * time.Millisecond
	enqueueAt := time.Now()
	queue := newDelayQueue[int](delayTime)
	for i := 1; i <= 5; i++ {
		queue.Enqueue(i)
	}
	taken := 0
	if v, ok := queue.Dequeue(); ok {
		if elapsed := time.Since(enqueueAt); elapsed < delayTime {
			t.Fatalf("未到期时出队不应成功, 得到 %v, elapsed=%s", v, elapsed)
		}
		if v != 1 {
			t.Fatalf("value=%d want=1", v)
		}
		taken = 1
	}

	deadline := time.Now().Add(waitTimeout)
	got := 0
	for got < 5-taken {
		if _, ok := queue.Dequeue(); ok {
			got++
			continue
		}
		if time.Now().After(deadline) {
			t.Fatalf("到期元素未全部出队: got=%d want=%d", got+taken, 5)
		}
		time.Sleep(time.Millisecond)
	}
}

// Given 空队列
// When 生产者延迟入队, 消费者带超时出队
// Then 超时语义与阻塞语义按预期工作
func Test_DequeueTimeout(t *testing.T) {
	t.Parallel()
	queue := BlockQueueWrapper(newSegQueue[int]())

	// 空队列: 超时过期返回 false
	if v, ok := queue.DequeueBlock(300 * time.Millisecond); ok {
		t.Fatal("空队列超时出队不应成功, 得到", v)
	}

	// 生产者入队后: 无参数阻塞版命中
	queue.Enqueue(1)
	got, ok := queue.DequeueBlock()
	if !ok || got != 1 {
		t.Fatalf("DequeueBlock = %v,%v want 1,true", got, ok)
	}
	_ = zipAny
}

// Given 多种队列实现
// When 并发生产并发消费
// Then 数量守恒
func Test_LkQueue(t *testing.T) {
	t.Parallel()
	num := 100000
	for _, o := range []struct {
		name string
		opt  Opt[int]
	}{
		{"chan", WithTypeChan[int](128)},
		{"seg", WithTypeSegment[int]()},
	} {
		t.Run(o.name, func(t *testing.T) {
			queue := NewQueue[int](o.opt)
			var consumed int64
			var wg sync.WaitGroup
			wg.Add(1)
			// 并发消费, 防止有界队列满时阻塞生产者
			go func() {
				defer wg.Done()
				for atomic.LoadInt64(&consumed) < int64(num) {
					if _, ok := queue.Dequeue(); ok {
						atomic.AddInt64(&consumed, 1)
					} else {
						runtime.Gosched()
					}
				}
			}()
			for i := 0; i < num; i++ {
				queue.Enqueue(1)
			}
			// 超时兜底: 消费停滞时快速失败而非挂死
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(30 * time.Second):
				t.Fatalf("消费停滞: consumed=%d want=%d", atomic.LoadInt64(&consumed), num)
			}
			if got := atomic.LoadInt64(&consumed); got != int64(num) {
				t.Fatalf("consumed=%d want=%d", got, num)
			}
		})
	}
}
