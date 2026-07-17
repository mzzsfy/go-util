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

type segSlot[T any] struct {
	seq   uint64
	value T
}

type segment[T any] struct {
	id    uint64
	next  unsafe.Pointer
	slots [segSize]segSlot[T]
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
	s := &seg.slots[pos&segMask]
	for {
		if atomic.LoadUint64(&s.seq) == pos {
			s.value = v
			atomic.StoreUint64(&s.seq, pos+1)
			return
		}
		runtime.Gosched()
	}
}

func (q *segQueue[T]) Dequeue() (T, bool) {
	for spin := 0; spin < 64; spin++ {
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
		s := &seg.slots[h&segMask]
		seq := atomic.LoadUint64(&s.seq)
		diff := int64(seq) - int64(h+1)
		if diff == 0 {
			if atomic.CompareAndSwapUint64(&q.headPos, h, h+1) {
				v := s.value
				var zero T
				s.value = zero
				atomic.StoreUint64(&s.seq, h+segSize)
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
	var zero T
	return zero, false
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
	s := &seg.slots[h&segMask]
	seq := atomic.LoadUint64(&s.seq)
	if int64(seq)-int64(h+1) != 0 {
		var zero T
		return zero, false
	}
	if !atomic.CompareAndSwapUint64(&q.headPos, h, h+1) {
		var zero T
		return zero, false
	}
	v := s.value
	var zero T
	s.value = zero
	atomic.StoreUint64(&s.seq, h+segSize)
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
		next := (*segment[T])(atomic.LoadPointer(&seg.next))
		if next == nil {
			return nil, false
		}
		seg = next
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
		next := (*segment[T])(atomic.LoadPointer(&head.next))
		if next == nil {
			break
		}
		head = next
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
		next := (*segment[T])(atomic.LoadPointer(&seg.next))
		if next != nil {
			seg = next
			continue
		}
		// 始终创建seg.id+1,保证链表连续不断裂
		ns := newSegment[T](seg.id + 1)
		nsp := unsafe.Pointer(ns)
		if atomic.CompareAndSwapPointer(&seg.next, nil, nsp) {
			atomic.CompareAndSwapPointer(&q.tailSeg, unsafe.Pointer(seg), nsp)
			seg = ns
			continue
		}
		seg = (*segment[T])(atomic.LoadPointer(&seg.next))
	}
}

func (q *segQueue[T]) findWriteSegment(targetID uint64) *segment[T] {
	for {
		seg := (*segment[T])(atomic.LoadPointer(&q.headSeg))
		for seg.id < targetID {
			next := (*segment[T])(atomic.LoadPointer(&seg.next))
			if next == nil {
				runtime.Gosched()
				break
			}
			seg = next
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
		s.slots[i].seq = base + i
	}
	return s
}
