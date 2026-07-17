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

func Test_DequeueTimeout(t *testing.T) {
	t.Parallel()
	queue := BlockQueueWrapper[int](BlockQueueWrapper(newSegQueue[int]()))
	go func() {
		for {
			i := 1
			time.Sleep(time.Millisecond * 500)
			queue.Enqueue(i)
		}
	}()
	t.Log(time.Now(), zipAny(queue.DequeueBlock(time.Millisecond*300)))
	t.Log(time.Now(), zipAny(queue.DequeueBlock(time.Millisecond*300)))
	t.Log(time.Now(), zipAny(queue.DequeueBlock()))
	t.Log(time.Now(), zipAny(queue.DequeueBlock(time.Millisecond*300)))
	t.Log(time.Now(), zipAny(queue.DequeueBlock()))
	t.Log(time.Now(), zipAny(queue.DequeueBlock()))
}

func Test_LkQueue(t *testing.T) {
	t.Parallel()
	num := 1000000
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
			// 并发消费,防止有界队列满时阻塞
			wg.Add(1)
			go func() {
				defer wg.Done()
				for atomic.LoadInt64(&consumed) < int64(num) {
					_, ok := queue.Dequeue()
					if ok {
						atomic.AddInt64(&consumed, 1)
					} else {
						runtime.Gosched()
					}
				}
			}()
			for i := 0; i < num; i++ {
				queue.Enqueue(1)
			}
			wg.Wait()
			if atomic.LoadInt64(&consumed) != int64(num) {
				t.Fatal("消费数据量不正确", atomic.LoadInt64(&consumed), num)
			}
		})
	}
}

func TestDelayQueue_Dequeue(t *testing.T) {
	t.Parallel()
	queue := newDelayQueue[int](time.Millisecond * 100)
	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)
	queue.Enqueue(4)
	queue.Enqueue(5)
	dequeue, b := queue.Dequeue()
	if b {
		t.Fatal("dequeue1", dequeue)
	}
	time.Sleep(time.Millisecond * 10)
	dequeue, b = queue.Dequeue()
	if b {
		t.Fatal("dequeue2", dequeue)
	}
	time.Sleep(time.Millisecond * 100)
	dequeue, b = queue.Dequeue()
	if !b {
		t.Fatal("dequeue3", dequeue)
	}
	t.Log(queue.Dequeue())
	t.Log(queue.Dequeue())
	t.Log(queue.Dequeue())
	t.Log(queue.Dequeue())
	t.Log(queue.Dequeue())
}
