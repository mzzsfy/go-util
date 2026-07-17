package unsafe

import (
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"
)

// ============================================================
// 问题1: unsafe.Pointer(&value) 是否导致value逃逸到堆?
// ============================================================

//go:noinline
func takePointer(p unsafe.Pointer) {
	// 什么都不做,仅用来观察参数传递是否导致逃逸
	_ = p
}

// 场景A: 对局部变量取 unsafe.Pointer,然后传给外部函数
//go:noinline
func EscapeTest_BasicInt() unsafe.Pointer {
	var v int = 42
	return unsafe.Pointer(&v) // 预期: v逃逸到堆,因为返回了指向它的指针
}

// 场景B: 对局部变量取 unsafe.Pointer,仅在函数内部使用
//go:noinline
func EscapeTest_LocalOnly() uintptr {
	var v int = 42
	p := unsafe.Pointer(&v)
	// 转换为uintptr,不再持有指针,应该不逃逸
	return uintptr(p)
}

// 场景C: 传入参数取 unsafe.Pointer
//go:noinline
func EscapeTest_Arg(v int) unsafe.Pointer {
	return unsafe.Pointer(&v) // v是参数,取地址是否逃逸?
}

// 场景D: noescape技巧 -- 复制运行时的做法
//go:nosplit
//go:nocheckptr
func noescapeLocal(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

//go:noinline
func EscapeTest_NoEscape() uintptr {
	var v int = 42
	_ = noescapeLocal(unsafe.Pointer(&v))
	return uintptr(noescapeLocal(unsafe.Pointer(&v)))
}

// ============================================================
// 问题2: 小值类型(int, small struct)是否强制堆分配?
// ============================================================

type smallStruct struct {
	a int
	b int
}

type pointerStruct struct {
	a *int
	b int
}

//go:noinline
func EscapeTest_SmallStruct() unsafe.Pointer {
	var s smallStruct
	s.a = 1
	s.b = 2
	return unsafe.Pointer(&s) // 小结构体,是否会逃逸?
}

//go:noinline
func EscapeTest_PointerStruct() unsafe.Pointer {
	var s pointerStruct
	x := 42
	s.a = &x
	return unsafe.Pointer(&s) // 含指针的结构体
}

//go:noinline
func EscapeTest_StructLocalOnly() int {
	var s smallStruct
	s.a = 1
	s.b = 2
	p := unsafe.Pointer(&s)
	s2 := (*smallStruct)(p)
	return s2.a + s2.b // 仅局部使用,是否不逃逸?
}

// ============================================================
// 问题4: atomic.StorePointer 存泛型值到数组
// ============================================================

type genericArray[T any] struct {
	arr [4]unsafe.Pointer
}

//go:noinline
func EscapeTest_AtomicStore() {
	var ga genericArray[int]
	var v int = 99
	// 关键: atomic.StorePointer 是否导致 v 逃逸?
	atomic.StorePointer(&ga.arr[0], unsafe.Pointer(&v))
}

// 同上但结构体
//go:noinline
func EscapeTest_AtomicStoreStruct() {
	var ga genericArray[smallStruct]
	var s smallStruct
	s.a = 1
	s.b = 2
	atomic.StorePointer(&ga.arr[0], unsafe.Pointer(&s))
}

// ============================================================
// 问题5: reflect.TypeOf 检查是否含指针
// ============================================================

//go:noinline
func EscapeTest_ReflectCheck() bool {
	var v int = 42
	t := reflect.TypeOf(v)
	// reflect.TypeOf 是否导致v逃逸?
	_ = t
	return t.Kind() == reflect.Int
}

//go:noinline
func hasPointers[T any]() bool {
	var zero T
	return reflect.TypeOf(zero).Kind() == reflect.Pointer ||
		reflect.TypeOf(zero).Kind() == reflect.Slice ||
		reflect.TypeOf(zero).Kind() == reflect.Map ||
		reflect.TypeOf(zero).Kind() == reflect.Chan ||
		reflect.TypeOf(zero).Kind() == reflect.Func ||
		reflect.TypeOf(zero).Kind() == reflect.Interface
}

// ============================================================
// 泛型函数中的逃逸行为
// ============================================================

//go:noinline
func GenericStorePointer[T any](v T) unsafe.Pointer {
	return unsafe.Pointer(&v)
}

//go:noinline
func GenericStoreLocalOnly[T any](v T) T {
	p := unsafe.Pointer(&v)
	result := *(*T)(p)
	return result
}

//go:noinline
func GenericAtomicStore[T any](arr []unsafe.Pointer, v T) {
	atomic.StorePointer(&arr[0], unsafe.Pointer(&v))
}

// ============================================================
// 模拟 channel 内部做法: memmove 复制值
// ============================================================

//go:noinline
func MemmoveStore[T any](dst, src *T) {
	size := unsafe.Sizeof(*src)
	// 使用 typedmemmove 的替代方案: 直接通过指针复制
	// 注意: 这里模拟channel的做法,直接复制值而非存指针
	copy(
		unsafe.Slice((*byte)(unsafe.Pointer(dst)), size),
		unsafe.Slice((*byte)(unsafe.Pointer(src)), size),
	)
}

//go:noinline
func EscapeTest_Memmove() int {
	var dst [16]byte
	var v int = 42
	MemmoveStore[int]((*int)(unsafe.Pointer(&dst[0])), &v)
	return *(*int)(unsafe.Pointer(&dst[0]))
}

// 另一个关键测试: 将值嵌入到 node 结构中(类似当前queue做法)
type node[T any] struct {
	value T
	next  unsafe.Pointer
}

//go:noinline
func EscapeTest_NodeEmbed() *node[int] {
	n := &node[int]{value: 42}
	return n
}

// 对比: 直接用 unsafe.Pointer 存值的地址 vs 嵌入值
//go:noinline
func EscapeTest_NodePointerStore() unsafe.Pointer {
	n := &node[int]{value: 42}
	// 将 value 的地址作为 unsafe.Pointer 存到别处
	return unsafe.Pointer(&n.value)
}

// ============================================================
// 真正的零分配测试: 不逃逸的 unsafe.Pointer 使用
// ============================================================

// 场景: 值存储在预分配的 []byte 中,通过 memmove 复制,不逃逸原始值
//go:noinline
func ZeroAllocStore[T any](buf []byte, v T) {
	var local T = v
	src := unsafe.Pointer(&local)
	size := unsafe.Sizeof(local)
	// 直接 copy 到 buf,local 不需要逃逸
	copy(buf[:size], unsafe.Slice((*byte)(src), size))
}

//go:noinline
func EscapeTest_ZeroAllocStore() int {
	buf := make([]byte, 16)
	var v int = 42
	ZeroAllocStore[int](buf, v)
	return *(*int)(unsafe.Pointer(&buf[0]))
}

// ============================================================
// 确认测试: benchmark 中检查分配
// ============================================================

func BenchmarkEscapeTest_AtomicStore(b *testing.B) {
	var ga genericArray[int]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v int = i
		atomic.StorePointer(&ga.arr[0], unsafe.Pointer(&v))
	}
}

func BenchmarkEscapeTest_NodeEmbed(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := &node[int]{value: i}
		_ = n
	}
}

func BenchmarkEscapeTest_MemmoveStore(b *testing.B) {
	var dst [16]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v int = i
		MemmoveStore[int]((*int)(unsafe.Pointer(&dst[0])), &v)
	}
}

func BenchmarkEscapeTest_ZeroAllocStore(b *testing.B) {
	buf := make([]byte, 16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v int = i
		ZeroAllocStore[int](buf, v)
	}
}
