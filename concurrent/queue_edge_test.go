package concurrent

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// enqueueTimeout 入队阻塞判定上限: 超过视为永久阻塞(满队列自旋语义)
const enqueueTimeout = 2 * time.Second

// waitTimeout 异步事件等待上限, 余量充足避免慢机误报
const waitTimeout = 5 * time.Second

// tryDequeueMakers 三种队列实现的构造器, 用于对拍 TryDequeue 行为
// serialCap 为串行场景安全批量: chanQueue 容量固定, 超容量入队会阻塞
var tryDequeueMakers = []struct {
	name      string
	make      func() Queue[int]
	block     bool // 是否实现 TryDequeuer
	serialCap int  // 串行测试可一次性入队的元素数
}{
	{"seg", func() Queue[int] { return newSegQueue[int]() }, true, 1000},
	{"ring", func() Queue[int] { return newRingQueue[int](8) }, true, 8},
	{"chan", func() Queue[int] { return newChanQueue[int](8) }, true, 8},
}

// Given 空队列
// When TryDequeue
// Then 立即返回 false 且不阻塞
func Test_TryDequeue_EmptyQueue(t *testing.T) {
	t.Parallel()
	for _, m := range tryDequeueMakers {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			q := m.make()
			if _, ok := q.(TryDequeuer[int]); !ok {
				t.Fatalf("%s 应实现 TryDequeuer", m.name)
			}
			done := make(chan struct{})
			go func() {
				defer close(done)
				td := q.(TryDequeuer[int])
				if v, ok := td.TryDequeue(); ok {
					t.Errorf("空队列 TryDequeue 不应成功, 得到 %v", v)
				}
			}()
			select {
			case <-done:
			case <-time.After(enqueueTimeout):
				t.Fatal("空队列 TryDequeue 不应阻塞")
			}
		})
	}
}

// Given 非空队列
// When TryDequeue
// Then 返回队头元素, FIFO 顺序保持
func Test_TryDequeue_NonEmptyFIFO(t *testing.T) {
	t.Parallel()
	for _, m := range tryDequeueMakers {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			q := m.make()
			td := q.(TryDequeuer[int])
			for i := 0; i < 4; i++ {
				q.Enqueue(i)
			}
			for want := 0; want < 4; want++ {
				got, ok := td.TryDequeue()
				if !ok || got != want {
					t.Fatalf("TryDequeue = %v,%v want %d,true", got, ok, want)
				}
			}
			if _, ok := td.TryDequeue(); ok {
				t.Fatal("取完后 TryDequeue 应返回 false")
			}
		})
	}
}

// Given 多生产者并发入队
// When 消费者并发 TryDequeue
// Then 每个元素恰好被取走一次
func Test_TryDequeue_ConcurrentExactlyOnce(t *testing.T) {
	t.Parallel()
	for _, m := range tryDequeueMakers {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			const producers, perProducer = 4, 250
			const total = producers * perProducer
			q := m.make()
			td := q.(TryDequeuer[int])

			var consumed int64
			var prodWG, consWG sync.WaitGroup
			for p := 0; p < producers; p++ {
				prodWG.Add(1)
				go func() {
					defer prodWG.Done()
					for i := 0; i < perProducer; i++ {
						q.Enqueue(1)
					}
				}()
			}
			stop := make(chan struct{})
			for c := 0; c < 4; c++ {
				consWG.Add(1)
				go func() {
					defer consWG.Done()
					for {
						select {
						case <-stop:
							return
						default:
						}
						if _, ok := td.TryDequeue(); ok {
							atomic.AddInt64(&consumed, 1)
						}
					}
				}()
			}
			prodWG.Wait()
			// 生产完毕后等待队列排空, deadline 兜底防挂死
			deadline := time.Now().Add(waitTimeout)
			for atomic.LoadInt64(&consumed) < total && q.Size() > 0 {
				if time.Now().After(deadline) {
					close(stop)
					t.Fatalf("排空超时: consumed=%d want=%d size=%d", consumed, total, q.Size())
				}
				time.Sleep(time.Millisecond)
			}
			close(stop)
			consWG.Wait()
			if consumed != total {
				t.Fatalf("consumed=%d want=%d", consumed, total)
			}
		})
	}
}

// Given cap 非二次幂
// When 构造 ring 队列并入队取整后的容量数
// Then 容量被向上取整到 2 的幂, 取整容量内入队不阻塞
func Test_Ring_CapacityRoundsUpToPow2(t *testing.T) {
	t.Parallel()
	// cap=3 取整为 4
	q := newRingQueue[int](3)
	for i := 0; i < 4; i++ {
		v := i
		done := make(chan struct{})
		go func() {
			defer close(done)
			q.Enqueue(v)
		}()
		select {
		case <-done:
		case <-time.After(enqueueTimeout):
			t.Fatalf("第 %d 个入队不应阻塞, 容量取整应为 4", i+1)
		}
	}
	if q.Size() != 4 {
		t.Fatalf("size=%d want=4", q.Size())
	}
	for want := 0; want < 4; want++ {
		got, ok := q.Dequeue()
		if !ok || got != want {
			t.Fatalf("dequeue=%v,%v want %d,true", got, ok, want)
		}
	}
}

// Given ring 队列已满
// When 再入队
// Then Enqueue 自旋等待而非覆盖, 出队腾位后入队完成
func Test_Ring_FullBlocksUntilSpace(t *testing.T) {
	t.Parallel()
	const capVal = 2
	q := newRingQueue[int](capVal)
	for i := 0; i < capVal; i++ {
		q.Enqueue(i)
	}

	enqueued := make(chan struct{})
	go func() {
		q.Enqueue(capVal)
		close(enqueued)
	}()
	select {
	case <-enqueued:
		t.Fatal("满队列入队应阻塞等待腾位")
	case <-time.After(20 * time.Millisecond):
		// 预期: 仍在自旋等待
	}

	got, ok := q.Dequeue()
	if !ok || got != 0 {
		t.Fatalf("dequeue=%v,%v want 0,true", got, ok)
	}
	select {
	case <-enqueued:
	case <-time.After(waitTimeout):
		t.Fatal("腾位后入队应完成")
	}
	if got, ok := q.Dequeue(); !ok || got != 1 {
		t.Fatalf("dequeue=%v,%v want 1,true", got, ok)
	}
	if got, ok := q.Dequeue(); !ok || got != capVal {
		t.Fatalf("dequeue=%v,%v want %d,true", got, ok, capVal)
	}
}

// Given chanQueue buffer 非法
// When 构造
// Then buffer 归一为 1, 第二个入队阻塞
func Test_ChanQueue_MinBufferClamp(t *testing.T) {
	t.Parallel()
	q := newChanQueue[int](0)
	q.Enqueue(1)
	if q.Size() != 1 {
		t.Fatalf("size=%d want=1", q.Size())
	}

	enqueued := make(chan struct{})
	go func() {
		q.Enqueue(2)
		close(enqueued)
	}()
	select {
	case <-enqueued:
		t.Fatal("容量1时第二个入队应阻塞")
	case <-time.After(20 * time.Millisecond):
		// 预期阻塞
	}
	if _, ok := q.Dequeue(); !ok {
		t.Fatal("dequeue should succeed")
	}
	select {
	case <-enqueued:
	case <-time.After(waitTimeout):
		t.Fatal("腾位后入队应完成")
	}
}

// Given chanQueue 空队列带超时出队
// When 超时小于入队延迟
// Then 到期返回 false; 延迟后入队的场景命中返回元素
func Test_ChanQueue_DequeueBlock_Timeout(t *testing.T) {
	t.Parallel()
	t.Run("过期返回false", func(t *testing.T) {
		t.Parallel()
		q := newChanQueue[int](1).(*chanQueue[int])
		start := time.Now()
		if v, ok := q.DequeueBlock(30 * time.Millisecond); ok {
			t.Fatalf("空队列超时出队不应成功, 得到 %v", v)
		}
		if elapsed := time.Since(start); elapsed < 25*time.Millisecond {
			t.Fatalf("过早返回, elapsed=%v", elapsed)
		}
	})
	t.Run("等待中命中", func(t *testing.T) {
		t.Parallel()
		q := newChanQueue[int](1).(*chanQueue[int])
		go func() {
			time.Sleep(10 * time.Millisecond)
			q.Enqueue(42)
		}()
		got, ok := q.DequeueBlock(waitTimeout)
		if !ok || got != 42 {
			t.Fatalf("DequeueBlock=%v,%v want 42,true", got, ok)
		}
	})
	t.Run("无参数阻塞命中", func(t *testing.T) {
		t.Parallel()
		q := newChanQueue[int](1).(*chanQueue[int])
		go func() {
			time.Sleep(10 * time.Millisecond)
			q.Enqueue(7)
		}()
		if got, ok := q.DequeueBlock(); !ok || got != 7 {
			t.Fatalf("DequeueBlock=%v,%v want 7,true", got, ok)
		}
	})
}

// Given blockQueue 包装 seg 队列
// When 空队列带超时出队
// Then 到期返回 false 且 waiter 归零
func Test_BlockQueue_TimeoutExpire(t *testing.T) {
	t.Parallel()
	q := BlockQueueWrapper[int](newSegQueue[int]())
	if v, ok := q.DequeueBlock(30 * time.Millisecond); ok {
		t.Fatalf("空队列超时出队不应成功, 得到 %v", v)
	}
	if w := q.WaiterCount(); w != 0 {
		t.Fatalf("超时后 waiter=%d want=0", w)
	}
}

// Given blockQueue 包装 seg 队列
// When 已有元素时带超时出队
// Then 快速命中, 不经等待
func Test_BlockQueue_HitWithoutWait(t *testing.T) {
	t.Parallel()
	q := BlockQueueWrapper[int](newSegQueue[int]())
	q.Enqueue(9)
	if got, ok := q.DequeueBlock(waitTimeout); !ok || got != 9 {
		t.Fatalf("DequeueBlock=%v,%v want 9,true", got, ok)
	}
}

// Given 已包装 BlockQueue 的队列
// When 再次 BlockQueueWrapper
// Then 原样返回同一实例
func Test_BlockQueueWrapper_Idempotent(t *testing.T) {
	t.Parallel()
	q := BlockQueueWrapper[int](newSegQueue[int]())
	if BlockQueueWrapper[int](q) != q {
		t.Fatal("已实现 BlockQueue 时应原样返回")
	}
}

// Given blockQueue 多消费者
// When 并发 DequeueBlock 与生产
// Then 每个元素恰好被消费一次
func Test_BlockQueue_ConcurrentDrain(t *testing.T) {
	t.Parallel()
	const producers, perProducer = 2, 200
	const total = producers * perProducer
	q := BlockQueueWrapper[int](newSegQueue[int]())

	var consumed int64
	var consWG sync.WaitGroup
	stop := make(chan struct{})
	for c := 0; c < 3; c++ {
		consWG.Add(1)
		go func() {
			defer consWG.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, ok := q.DequeueBlock(50 * time.Millisecond); ok {
					atomic.AddInt64(&consumed, 1)
				}
			}
		}()
	}
	var prodWG sync.WaitGroup
	for p := 0; p < producers; p++ {
		prodWG.Add(1)
		go func() {
			defer prodWG.Done()
			for i := 0; i < perProducer; i++ {
				q.Enqueue(1)
			}
		}()
	}
	prodWG.Wait()

	deadline := time.Now().Add(waitTimeout)
	for atomic.LoadInt64(&consumed) < total {
		if time.Now().After(deadline) {
			close(stop)
			t.Fatalf("排空超时: consumed=%d want=%d", consumed, total)
		}
		time.Sleep(time.Millisecond)
	}
	close(stop)
	consWG.Wait()
	if consumed != total {
		t.Fatalf("consumed=%d want=%d", consumed, total)
	}
}
