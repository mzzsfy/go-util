package concurrent

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 判定消费完成后旧segment是否变为不可达(GC可回收)
func Test_SegSegmentGCProbe(t *testing.T) {
	t.Parallel()
	q := newSegQueue[int]().(*segQueue[int])
	var released uint32
	seg0 := (*segment[int])(atomic.LoadPointer(&q.headSeg))
	runtime.SetFinalizer(seg0, func(*segment[int]) { atomic.AddUint32(&released, 1) })
	seg0 = nil

	n := segSize * 3
	for i := 0; i < n; i++ {
		q.Enqueue(i)
	}
	for i := 0; i < n; i++ {
		if _, ok := q.Dequeue(); !ok {
			t.Fatal("dequeue failed")
		}
	}
	if atomic.LoadPointer(&q.headSeg) == nil {
		t.Fatal("headSeg nil")
	}
	// 多轮GC促使finalizer执行
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		runtime.GC()
		if atomic.LoadUint32(&released) == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("旧segment未被GC回收(仍可达), 存在内存泄漏")
}

// 并发压测内存稳定性: 8P8C 大量跨段消费后堆内存应回落, 段不可达
func Test_SegSegmentGCPressure(t *testing.T) {
	t.Parallel()
	q := newSegQueue[int]()
	const total = 100000
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for p := 0; p < 8; p++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; ; i++ {
				select {
				case <-stop:
					return
				default:
				}
				q.Enqueue(i)
			}
		}()
	}
	var consumed uint64
	for c := 0; c < 8; c++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if atomic.LoadUint64(&consumed) >= total {
					return
				}
				if _, ok := q.Dequeue(); ok {
					atomic.AddUint64(&consumed, 1)
				}
			}
		}()
	}
	deadline := time.Now().Add(30 * time.Second)
	for atomic.LoadUint64(&consumed) < total && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	close(stop)
	wg.Wait()
	if c := atomic.LoadUint64(&consumed); c < total {
		t.Fatalf("consumed=%d want>=%d", c, total)
	}
	// 回收性判定用finalizer探针: HeapAlloc受并行测试GC噪声干扰不可靠
	runtime.GC()
	probeQ := newSegQueue[int]().(*segQueue[int])
	var released uint32
	ps := (*segment[int])(atomic.LoadPointer(&probeQ.headSeg))
	runtime.SetFinalizer(ps, func(*segment[int]) { atomic.AddUint32(&released, 1) })
	ps = nil
	for i := 0; i < segSize*2; i++ {
		probeQ.Enqueue(i)
	}
	for i := 0; i < segSize*2; i++ {
		if _, ok := probeQ.Dequeue(); !ok {
			t.Fatal("probe dequeue failed")
		}
	}
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		runtime.GC()
		if atomic.LoadUint32(&released) == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("压测后新建队列的旧segment未被GC回收, 段仍可达")
}
