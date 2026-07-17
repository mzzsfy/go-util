package unsafe

import (
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"
)

// ============================================================
// 补充测试: 更细致的逃逸场景
// ============================================================

// 场景E: noescape 技巧 + atomic.StorePointer
// 这是关键: 用 noescape 包装后, atomic.StorePointer 是否还导致逃逸?
//go:noinline
func EscapeTest_NoEscapeAtomicStore() {
	var ga genericArray[int]
	var v int = 99
	p := noescapeLocal(unsafe.Pointer(&v))
	// noescape 已经切断逃逸分析链,但 atomic.StorePointer 仍会把指针值存到堆上
	// 关键区别: v本身不会逃逸,但p指向的地址实际上是栈上的v
	// 如果v在栈上,存储栈指针到堆上的 arr 中是危险的!
	atomic.StorePointer(&ga.arr[0], p)
}

// 场景F: 用 reflect 检查 T 是否含指针,并据此选择存储策略
//go:noinline
func isPointerFree[T any]() bool {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return isPointerFreeType(t)
}

func isPointerFreeType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map,
		reflect.Chan, reflect.Func, reflect.Interface,
		reflect.UnsafePointer:
		return false
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if !isPointerFreeType(t.Field(i).Type) {
				return false
			}
		}
	}
	return true
}

// 场景G: 使用 unsafe.Sizeof + typedmemmove 的替代方案
// 这里用 reflect.TypeOf 检查后,不含指针的类型直接 copy
//go:noinline
func SafeCopy[T any](dst, src *T) {
	size := unsafe.Sizeof(*src)
	// 对于不含指针的类型, 直接 copy 是安全的
	// 对于含指针的类型, 需要 reflectlite + write barrier
	copy(
		unsafe.Slice((*byte)(unsafe.Pointer(dst)), size),
		unsafe.Slice((*byte)(unsafe.Pointer(src)), size),
	)
}

// 用 reflect.TypeOf 判断的零分配存储
//go:noinline
func ReflectTypeCheck() {
	_ = isPointerFree[int]()
	_ = isPointerFree[string]()
	_ = isPointerFree[smallStruct]()
	_ = isPointerFree[pointerStruct]()
}

// ============================================================
// 模拟 channel: 直接 memmove 到预分配缓冲区
// ============================================================

// 模拟 hchan 的环形缓冲区
type ChanSim[T any] struct {
	buf  []byte // 预分配的字节缓冲区
	head uint32
	tail uint32
	size uintptr // unsafe.Sizeof(T)
	cap  uint32
}

func NewChanSim[T any](cap int) *ChanSim[T] {
	var zero T
	elemSize := unsafe.Sizeof(zero)
	return &ChanSim[T]{
		buf:  make([]byte, int(elemSize)*cap),
		size: elemSize,
		cap:  uint32(cap),
	}
}

//go:noinline
func (c *ChanSim[T]) Send(v T) {
	idx := c.tail % c.cap
	offset := uintptr(idx) * c.size
	// 直接 copy 值到 buf, 不存指针
	src := unsafe.Pointer(&v)
	copy(c.buf[offset:offset+c.size], unsafe.Slice((*byte)(src), c.size))
	c.tail++
}

//go:noinline
func (c *ChanSim[T]) Recv() T {
	idx := c.head % c.cap
	offset := uintptr(idx) * c.size
	var result T
	dst := unsafe.Pointer(&result)
	copy(unsafe.Slice((*byte)(dst), c.size), c.buf[offset:offset+c.size])
	c.head++
	return result
}

// 测试 ChanSim 是否零分配
//go:noinline
func EscapeTest_ChanSim() int {
	ch := NewChanSim[int](4)
	ch.Send(42)
	v := ch.Recv()
	return v
}

// ============================================================
// 泛型 + atomic.StorePointer 的实际分配测试
// ============================================================

// 当前 queue_free_lock_link_array.go 的做法:
// SetV 中用 atomic.StorePointer(&l.arr[idx], unsafe.Pointer(v))
// v 是 *T 类型, T 的值嵌入在 node 中
// 这意味着每次 Enqueue 都需要分配一个 node, T 的值随之到堆上
//go:noinline
func EscapeTest_CurrentQueuePattern() {
	var arr [4]unsafe.Pointer
	n := &node[int]{value: 42}
	// n 本身需要堆分配 (因为它需要比函数活得更久)
	atomic.StorePointer(&arr[0], unsafe.Pointer(n))
}

// 改进方案: 用 []byte + memmove 存储 T 的值, 不需要分配 node
//go:noinline
func EscapeTest_MemmovePattern() int {
	// 预分配缓冲区
	buf := make([]byte, 8) // sizeof(int)
	var v int = 42
	src := unsafe.Pointer(&v)
	copy(buf, unsafe.Slice((*byte)(src), unsafe.Sizeof(v)))
	return *(*int)(unsafe.Pointer(&buf[0]))
}

// ============================================================
// Benchmark: 对比不同存储策略
// ============================================================

func BenchmarkCurrentQueuePattern(b *testing.B) {
	var arr [4]unsafe.Pointer
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		n := &node[int]{value: i}
		atomic.StorePointer(&arr[0], unsafe.Pointer(n))
	}
}

func BenchmarkMemmovePattern(b *testing.B) {
	buf := make([]byte, 8)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var v int = i
		src := unsafe.Pointer(&v)
		copy(buf, unsafe.Slice((*byte)(src), unsafe.Sizeof(v)))
	}
}

func BenchmarkChanSimPattern(b *testing.B) {
	ch := NewChanSim[int](4)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ch.Send(i)
		_ = ch.Recv()
	}
}

// 小结构体对比
func BenchmarkCurrentQueuePatternStruct(b *testing.B) {
	var arr [4]unsafe.Pointer
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		n := &node[smallStruct]{value: smallStruct{a: i, b: i + 1}}
		atomic.StorePointer(&arr[0], unsafe.Pointer(n))
	}
}

func BenchmarkMemmovePatternStruct(b *testing.B) {
	buf := make([]byte, 16) // sizeof(smallStruct)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := smallStruct{a: i, b: i + 1}
		src := unsafe.Pointer(&v)
		copy(buf, unsafe.Slice((*byte)(src), unsafe.Sizeof(v)))
	}
}

// ============================================================
// 测试 reflect.TypeOf 是否导致泛型变量逃逸
// ============================================================

//go:noinline
func GenericReflectType[T any](v T) reflect.Type {
	// reflect.TypeOf 使用 any(v) 转换, 这是否导致 v 逃逸?
	return reflect.TypeOf(v)
}

//go:noinline
func GenericReflectTypeNoEscape[T any](v T) bool {
	// 如果 T 不含指针, any(v) 不会逃逸
	// reflect.TypeOf 内部调用了 abi.TypeOf, 它用 abi.NoEscape 保护
	_ = reflect.TypeOf(v)
	return true
}
