package storage

import (
	"github.com/mzzsfy/go-util/unsafe"
	"math"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
)

var GoID = unsafe.GoID

//基于goroutine的局部存储,类似于threadLocal,必须调用GOClean清理,明确后需要调用 KnowHowToUseGls()

var (
	knowHowToUseGls = false

	glsMap sync.Map // int64 -> Map[uint64, any]

	glsEntryCount int64 // glsMap 中的活跃条目数, 原子维护, 替代 Range 计数

	checkLock sync.RWMutex

	glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeArray[uint64, any](2)) }}
)

type GlsError struct {
	NumGoroutine int
	GlsGoIds     []int64
}

func (g GlsError) Error() string {
	return g.String()
}

func (g GlsError) String() string {
	return "gls发生泄露,请检查!goroutine数量:" + strconv.Itoa(g.NumGoroutine) + ", gls数量:" + strconv.Itoa(len(g.GlsGoIds))
}

var keyIdGen = uint64(0)

// Key 提供一个类型安全的key
type Key[T any] interface {
	Get() (value T, exist bool)
	// GetById 只读获取指定goid的key对应的值,提供跨携程访问窗口
	// 不会为目标goid创建存储, goid不存在或未Set时返回 (零值, false)
	GetById(goid int64) (value T, exist bool)
	Set(T)
	// Delete 删除key,如果autoClean为true,则删除后,如果当前goid没有其他key,则自动清理
	Delete(autoClean ...bool)
}

// getSubMap 快速读路径, 可被编译器内联
func getSubMap(goid int64) Map[uint64, any] {
	v, ok := glsMap.Load(goid)
	if !ok {
		return nil
	}
	return v.(Map[uint64, any])
}

// getOrInitSubMap 读路径未命中时创建子表, 仅首次访问每 goroutine 触发
func getOrInitSubMap(goid int64) Map[uint64, any] {
	if m := getSubMap(goid); m != nil {
		return m
	}
	m := glsSubMapPool.Get().(Map[uint64, any])
	actual, loaded := glsMap.LoadOrStore(goid, m)
	if !loaded {
		atomic.AddInt64(&glsEntryCount, 1)
	}
	am := actual.(Map[uint64, any])
	if am != m {
		glsSubMapPool.Put(m)
	}
	return am
}

type KeySimple[T any] uint64

func (k KeySimple[T]) Get() (value T, exist bool) {
	return k.GetById(GoID())
}

func (k KeySimple[T]) GetById(goid int64) (value T, exist bool) {
	m := getSubMap(goid)
	if m == nil {
		return
	}
	get, o := m.Get(uint64(k))
	if o {
		value = get.(T)
		exist = true
	}
	return
}

func (k KeySimple[T]) GetSimple() (value T) {
	get, _ := k.Get()
	return get
}

func (k KeySimple[T]) Set(value T) {
	if !knowHowToUseGls {
		panic("call `KnowHowToUseGls()` after you know how to use it! (You must call Clean() after code)")
	}
	id := GoID()
	m := getOrInitSubMap(id)
	m.Put(uint64(k), any(value))
	check()
}

func (k KeySimple[T]) Delete(c ...bool) {
	id := GoID()
	m := getSubMap(id)
	if m != nil {
		m.Delete(uint64(k))
		if len(c) > 0 && c[0] && m.Count() == 0 {
			GlsClean()
		}
	}
}

type KeyFn[T any] struct {
	key uint64
	fn  func() T
}

func (k *KeyFn[T]) Get() (value T, exist bool) {
	id := GoID()
	if m := getSubMap(id); m != nil {
		if get, o := m.Get(k.key); o {
			return get.(T), true
		}
	}
	// 默认值工厂仅在自身 goid 上触发延迟初始化
	m := getOrInitSubMap(id)
	v := k.fn()
	m.Put(k.key, any(v))
	return v, true
}

// GetById 只读查询, 与 KeySimple.GetById 语义一致:
// 不会为任意 goid 创建子表(否则伪造/已退出的 goid 会造成长期泄漏)
func (k *KeyFn[T]) GetById(goid int64) (value T, exist bool) {
	m := getSubMap(goid)
	if m == nil {
		return
	}
	get, o := m.Get(k.key)
	if o {
		value = get.(T)
		exist = true
	}
	return
}

func (k *KeyFn[T]) Set(t T) {
	KeySimple[T](k.key).Set(t)
}

func (k *KeyFn[T]) Delete(b ...bool) {
	KeySimple[T](k.key).Delete(b...)
}

func KnowHowToUseGls() {
	knowHowToUseGls = true
}
func GlsCleanWithId(goid int64) {
	v, loaded := glsMap.LoadAndDelete(goid)
	if !loaded {
		return
	}
	atomic.AddInt64(&glsEntryCount, -1)
	m := v.(Map[uint64, any])
	m.Clean()
	glsSubMapPool.Put(m)
}
func GlsClean() {
	GlsCleanWithId(GoID())
}

// NewGlsItem returns a new unique key.
//
// var item = NewGlsItem[int]() //at global
// value,exist := item.Get()
func NewGlsItem[T any]() Key[T] {
	id := atomic.AddUint64(&keyIdGen, 1)
	if id == 0 {
		keyIdGen = math.MaxUint64 //无限 panic

		// 使用全局定义key不可能超过2^64个
		// 2^64 keys should be enough for everyone
		panic("gls key too much, overflow! You should define keys in global, don't declare a new one every time")
	}
	return KeySimple[T](id)
}

func NewGlsItemWithFunc[T any](defaultValue func() T) Key[T] {
	return &KeyFn[T]{
		key: uint64(NewGlsItem[T]().(KeySimple[T])),
		fn:  defaultValue,
	}
}

func NewGlsItemWithDefault[T any](defaultValue T) Key[T] {
	return NewGlsItemWithFunc(func() T { return defaultValue })
}

// 泄漏检查间隔, 概率性触发: 每次 Set 以 1/checkInterval 的概率执行检查
var checkInterval int64 = 10

// 检查gls是否泄露,每次检查都会增加检查间隔,直到检查间隔大于1000000,所以对性能几乎没影响
func check() {
	interval := atomic.LoadInt64(&checkInterval)
	if fastrand()%uint32(interval) != 0 {
		return
	}
	glsN := int(atomic.LoadInt64(&glsEntryCount))
	if runtime.NumGoroutine() < glsN {
		runtime.Gosched()
		cpuAddNum := runtime.NumCPU() + 32
		glsN = int(atomic.LoadInt64(&glsEntryCount))
		if runtime.NumGoroutine()+cpuAddNum < glsN {
			runtime.GC()
			runtime.Gosched()
			checkLock.Lock()
			runtime.GC()
			numGoroutine := runtime.NumGoroutine()
			glsN = int(atomic.LoadInt64(&glsEntryCount))
			if numGoroutine+cpuAddNum < glsN {
				var ids []int64
				glsMap.Range(func(k, _ any) bool {
					ids = append(ids, k.(int64))
					// 必须遍历完整表, 提前终止会导致泄漏报告只含部分 goid
					return true
				})
				checkLock.Unlock()
				panic(GlsError{
					NumGoroutine: numGoroutine,
					GlsGoIds:     ids,
				})
			}
			checkLock.Unlock()
		}
	}
	if interval > 1_000_000 {
		atomic.StoreInt64(&checkInterval, 1_000_001)
	} else {
		atomic.StoreInt64(&checkInterval, interval+interval/2)
	}
}
