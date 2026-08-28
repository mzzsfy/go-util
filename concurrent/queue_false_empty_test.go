package concurrent

import (
	"sync/atomic"
	"testing"
	"time"
)

// 假空复现: 生产者占位(tail推进)后未发布seq, 单次Dequeue必须等待发布而非返回假空
func Test_SegDequeueNotFalseEmpty(t *testing.T) {
	t.Parallel()
	q := newSegQueue[int]().(*segQueue[int])
	pos := atomic.AddUint64(&q.tailPos, 1) - 1
	seg := q.ensureWriteSegment(pos >> segBits)
	idx := pos & segMask
	// 模拟生产者在seq发布前被长时间调度走
	go func() {
		time.Sleep(20 * time.Millisecond)
		seg.vals[idx] = 42
		atomic.StoreUint64(&seg.seqs[idx], pos+1)
	}()
	type result struct {
		v  int
		ok bool
	}
	done := make(chan result, 1)
	go func() {
		v, ok := q.Dequeue()
		done <- result{v, ok}
	}()
	select {
	case r := <-done:
		if !r.ok {
			t.Fatal("Dequeue自旋耗尽返回假空, 生产者已占位未发布")
		}
		if r.v != 42 {
			t.Fatalf("v=%d want=42", r.v)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Dequeue在3s内未等到元素发布")
	}
}

func Test_RingDequeueNotFalseEmpty(t *testing.T) {
	t.Parallel()
	q := newRingQueue[int](16).(*ringQueue[int])
	pos := atomic.AddUint64(&q.tail, 1) - 1
	idx := pos & q.mask
	go func() {
		time.Sleep(20 * time.Millisecond)
		q.vals[idx] = 42
		atomic.StoreUint64(&q.seqs[idx], pos+1)
	}()
	type result struct {
		v  int
		ok bool
	}
	done := make(chan result, 1)
	go func() {
		v, ok := q.Dequeue()
		done <- result{v, ok}
	}()
	select {
	case r := <-done:
		if !r.ok {
			t.Fatal("Dequeue返回假空, 生产者已占位未发布")
		}
		if r.v != 42 {
			t.Fatalf("v=%d want=42", r.v)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Dequeue在3s内未等到元素发布")
	}
}
