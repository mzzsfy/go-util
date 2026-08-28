package concurrent

import (
	"runtime"
	"sync/atomic"
)

// MPMC有界队列,预分配ring buffer,零alloc
// 算法参考Dmitry Vyukov
type ringQueue[T any] struct {
	// 32位平台仅首字段保证8字节对齐, head/tail/seqs需64位原子访问
	head uint64
	_    [cpuCacheKillerPaddingLength]byte
	tail uint64
	_    [cpuCacheKillerPaddingLength]byte
	mask uint64
	seqs []uint64
	vals []T
}

// WithTypeRing 使用固定容量环形队列,零分配,MPMC安全
// cap会被向上取整到最近的2的幂
func WithTypeRing[T any](cap int) Opt[T] {
	return func(opt *opt[T]) {
		opt.Type = func() Queue[T] {
			return newRingQueue[T](cap)
		}
	}
}

func newRingQueue[T any](cap int) Queue[T] {
	// 容量必须是2的幂
	if cap < 1 {
		cap = 1
	}
	cap = nextPow2(cap)
	seqs := make([]uint64, cap)
	for i := range seqs {
		seqs[i] = uint64(i)
	}
	return &ringQueue[T]{
		seqs: seqs,
		vals: make([]T, cap),
		mask: uint64(cap - 1),
	}
}

func nextPow2(v int) int {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	return v + 1
}

func (q *ringQueue[T]) Size() int {
	return int(atomic.LoadUint64(&q.tail) - atomic.LoadUint64(&q.head))
}

func (q *ringQueue[T]) Enqueue(v T) {
	for {
		t := atomic.LoadUint64(&q.tail)
		idx := t & q.mask
		seq := atomic.LoadUint64(&q.seqs[idx])
		diff := int64(seq) - int64(t)
		if diff == 0 {
			// slot可用,尝试占位
			if atomic.CompareAndSwapUint64(&q.tail, t, t+1) {
				q.vals[idx] = v
				atomic.StoreUint64(&q.seqs[idx], t+1)
				return
			}
		} else if diff < 0 {
			// 队列满,让出CPU
			runtime.Gosched()
		}
	}
}

func (q *ringQueue[T]) Dequeue() (T, bool) {
	// 无自旋上限: 竞争/生产者未发布时等待而非假空, 空队列由h>=tail判定
	for spin := 0; ; spin++ {
		h := atomic.LoadUint64(&q.head)
		idx := h & q.mask
		seq := atomic.LoadUint64(&q.seqs[idx])
		diff := int64(seq) - int64(h+1)
		if diff == 0 {
			if atomic.CompareAndSwapUint64(&q.head, h, h+1) {
				v := q.vals[idx]
				var zero T
				q.vals[idx] = zero
				atomic.StoreUint64(&q.seqs[idx], h+q.mask+1)
				return v, true
			}
		} else if diff < 0 && h >= atomic.LoadUint64(&q.tail) {
			var zero T
			return zero, false
		}
		if spin > 0 && spin%8 == 0 {
			runtime.Gosched()
		}
	}
}

// TryDequeue 单次CAS尝试,失败立即返回
func (q *ringQueue[T]) TryDequeue() (T, bool) {
	h := atomic.LoadUint64(&q.head)
	idx := h & q.mask
	seq := atomic.LoadUint64(&q.seqs[idx])
	if int64(seq)-int64(h+1) != 0 {
		var zero T
		return zero, false
	}
	if !atomic.CompareAndSwapUint64(&q.head, h, h+1) {
		var zero T
		return zero, false
	}
	v := q.vals[idx]
	var zero T
	q.vals[idx] = zero
	atomic.StoreUint64(&q.seqs[idx], h+q.mask+1)
	return v, true
}
