package concurrent

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// 分段MPMC队列,动态大小
// 每个segment是独立的ring,通过链表连接实现动态增长
// 算法参考Dmitry Vyukov

const (
	segBits = 9
	segSize = 1 << segBits
	segMask = segSize - 1
)

// segPlaceholder 段分配占位哨兵: 先占位后分配, 保证段边界处仅一次分配,
// 读者见占位视作未就绪, 写者见占位等待发布
var segAllocating byte
var segPlaceholder = unsafe.Pointer(&segAllocating)

type segment[T any] struct {
	// seqs必须为首字段: 32位平台仅分配结构的首字保证8字节对齐, seq需64位原子访问
	seqs [segSize]uint64
	vals [segSize]T
	id   uint64
	next unsafe.Pointer
}

type segQueue[T any] struct {
	headPos uint64
	_       [cpuCacheKillerPaddingLength]byte
	tailPos uint64
	_       [cpuCacheKillerPaddingLength]byte
	headSeg unsafe.Pointer
	_       [cpuCacheKillerPaddingLength]byte
	tailSeg unsafe.Pointer
}

func newSegQueue[T any]() Queue[T] {
	s := newSegment[T](0)
	sp := unsafe.Pointer(s)
	return &segQueue[T]{
		headSeg: sp,
		tailSeg: sp,
	}
}

// WithTypeSegment 使用分段队列,动态大小,MPMC安全
func WithTypeSegment[T any]() Opt[T] {
	return func(opt *opt[T]) {
		opt.Type = newSegQueue[T]
	}
}

func (q *segQueue[T]) Enqueue(v T) {
	pos := atomic.AddUint64(&q.tailPos, 1) - 1
	segID := pos >> segBits
	seg := (*segment[T])(atomic.LoadPointer(&q.tailSeg))
	if seg.id != segID {
		seg = q.ensureWriteSegment(segID)
	}
	idx := pos & segMask
	for {
		if atomic.LoadUint64(&seg.seqs[idx]) == pos {
			seg.vals[idx] = v
			atomic.StoreUint64(&seg.seqs[idx], pos+1)
			return
		}
		runtime.Gosched()
	}
}

func (q *segQueue[T]) Dequeue() (T, bool) {
	// 无自旋上限: 竞争/生产者未发布时等待而非假空, 空队列由h>=tailPos判定
	for spin := 0; ; spin++ {
		h := atomic.LoadUint64(&q.headPos)
		// 内联findReadSegment快速路径
		seg := (*segment[T])(atomic.LoadPointer(&q.headSeg))
		if seg.id != h>>segBits {
			var ok bool
			seg, ok = q.findReadSegment(h >> segBits)
			if !ok {
				var zero T
				return zero, false
			}
		}
		idx := h & segMask
		seq := atomic.LoadUint64(&seg.seqs[idx])
		diff := int64(seq) - int64(h+1)
		if diff == 0 {
			if atomic.CompareAndSwapUint64(&q.headPos, h, h+1) {
				v := seg.vals[idx]
				var zero T
				seg.vals[idx] = zero
				atomic.StoreUint64(&seg.seqs[idx], h+segSize)
				if (h & segMask) == segMask {
					q.advanceHead()
				}
				return v, true
			}
		}
		if diff < 0 && spin > 16 && h >= atomic.LoadUint64(&q.tailPos) {
			var zero T
			return zero, false
		}
		if spin > 0 && spin%8 == 0 {
			runtime.Gosched()
		}
	}
}

func (q *segQueue[T]) TryDequeue() (T, bool) {
	h := atomic.LoadUint64(&q.headPos)
	seg := (*segment[T])(atomic.LoadPointer(&q.headSeg))
	if seg.id != h>>segBits {
		var ok bool
		seg, ok = q.findReadSegment(h >> segBits)
		if !ok {
			var zero T
			return zero, false
		}
	}
	idx := h & segMask
	seq := atomic.LoadUint64(&seg.seqs[idx])
	if int64(seq)-int64(h+1) != 0 {
		var zero T
		return zero, false
	}
	if !atomic.CompareAndSwapUint64(&q.headPos, h, h+1) {
		var zero T
		return zero, false
	}
	v := seg.vals[idx]
	var zero T
	seg.vals[idx] = zero
	atomic.StoreUint64(&seg.seqs[idx], h+segSize)
	if (h & segMask) == segMask {
		q.advanceHead()
	}
	return v, true
}

func (q *segQueue[T]) Size() int {
	return int(atomic.LoadUint64(&q.tailPos) - atomic.LoadUint64(&q.headPos))
}

func (q *segQueue[T]) findReadSegment(targetID uint64) (*segment[T], bool) {
	seg := (*segment[T])(atomic.LoadPointer(&q.headSeg))
	if seg.id == targetID {
		return seg, true
	}
	if seg.id > targetID {
		return nil, false
	}
	for seg.id < targetID {
		nextPtr := atomic.LoadPointer(&seg.next)
		// 占位中视作未就绪, 由调用方按空处理并重试
		if nextPtr == nil || nextPtr == segPlaceholder {
			return nil, false
		}
		seg = (*segment[T])(nextPtr)
	}
	if seg.id != targetID {
		return nil, false
	}
	return seg, true
}

func (q *segQueue[T]) advanceHead() {
	head := (*segment[T])(atomic.LoadPointer(&q.headSeg))
	targetID := atomic.LoadUint64(&q.headPos) >> segBits
	if head.id >= targetID {
		return
	}
	for head.id < targetID {
		nextPtr := atomic.LoadPointer(&head.next)
		if nextPtr == nil || nextPtr == segPlaceholder {
			break
		}
		head = (*segment[T])(nextPtr)
	}
	old := atomic.LoadPointer(&q.headSeg)
	atomic.CompareAndSwapPointer(&q.headSeg, old, unsafe.Pointer(head))
}

func (q *segQueue[T]) ensureWriteSegment(targetID uint64) *segment[T] {
	seg := (*segment[T])(atomic.LoadPointer(&q.tailSeg))
	for {
		if seg.id == targetID {
			return seg
		}
		if seg.id > targetID {
			return q.findWriteSegment(targetID)
		}
		nextPtr := atomic.LoadPointer(&seg.next)
		if nextPtr == segPlaceholder {
			// 他者正在分配, 等待发布
			runtime.Gosched()
			continue
		}
		if nextPtr == nil {
			if !atomic.CompareAndSwapPointer(&seg.next, nil, segPlaceholder) {
				// 他者抢先占位或已发布, 重读
				continue
			}
			// 占位成功者独占分配, 消除冲突时的段垃圾
			ns := newSegment[T](seg.id + 1)
			atomic.StorePointer(&seg.next, unsafe.Pointer(ns))
			atomic.CompareAndSwapPointer(&q.tailSeg, unsafe.Pointer(seg), unsafe.Pointer(ns))
			seg = ns
			continue
		}
		seg = (*segment[T])(nextPtr)
	}
}

func (q *segQueue[T]) findWriteSegment(targetID uint64) *segment[T] {
	for {
		seg := (*segment[T])(atomic.LoadPointer(&q.headSeg))
		for seg.id < targetID {
			nextPtr := atomic.LoadPointer(&seg.next)
			if nextPtr == nil || nextPtr == segPlaceholder {
				runtime.Gosched()
				break
			}
			seg = (*segment[T])(nextPtr)
		}
		if seg.id == targetID {
			return seg
		}
		runtime.Gosched()
	}
}

func newSegment[T any](id uint64) *segment[T] {
	s := &segment[T]{id: id}
	base := id << segBits
	for i := uint64(0); i < segSize; i++ {
		s.seqs[i] = base + i
	}
	return s
}
