package storage

import (
    "github.com/mzzsfy/go-util/concurrent"
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

    glsMap = NewMap(MapTypeSwissConcurrent[int64, Map[uint64, any]]())

    glsLock   concurrent.RwLocker = concurrent.NoLock{}
    checkLock concurrent.RwLocker = &sync.RWMutex{}

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
    // GetById 获取指定goid的key对应的值,提供跨携程访问窗口
    GetById(goid int64) (value T, exist bool)
    Set(T)
    // Delete 删除key,如果autoClean为true,则删除后,如果当前goid没有其他key,则自动清理
    Delete(autoClean ...bool)
}

func getSubMap(goid int64, init bool) Map[uint64, any] {
    glsLock.RLock()
    m, ok := glsMap.Get(goid)
    glsLock.RUnlock()
    if !ok {
        if init {
            m = glsSubMapPool.Get().(Map[uint64, any])
            glsLock.Lock()
            glsMap.Put(goid, m)
            glsLock.Unlock()
        } else {
            return nil
        }
    }
    return m
}

type KeySimple[T any] uint64

func (k KeySimple[T]) Get() (value T, exist bool) {
    return k.GetById(GoID())
}

func (k KeySimple[T]) GetById(goid int64) (value T, exist bool) {
    m := getSubMap(goid, false)
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
    m := getSubMap(id, true)
    m.Put(uint64(k), any(value))
    check()
}

func (k KeySimple[T]) Delete(c ...bool) {
    id := GoID()
    m := getSubMap(id, false)
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
    return k.GetById(GoID())
}

func (k *KeyFn[T]) GetById(goid int64) (value T, exist bool) {
    exist = true
    m := getSubMap(goid, true)
    get, o := m.Get(k.key)
    if o {
        value = get.(T)
    }
    get = k.fn()
    m.Put(k.key, get)
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
    m := getSubMap(goid, false)
    if m == nil {
        return
    }
    glsLock.Lock()
    glsMap.Delete(goid)
    glsLock.Unlock()

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

var testI = 0
var testInterval = 10

//检查gls是否泄露,每次检查都会增加检查间隔,直到检查间隔大于1000000,所以对性能几乎没影响
func check() {
    testI++
    if testI >= testInterval {
        cpuAddNum := runtime.NumCPU() + 64
        if runtime.NumGoroutine() < glsMap.Count()+cpuAddNum {
            runtime.Gosched()
            if runtime.NumGoroutine() < glsMap.Count()+cpuAddNum {
                runtime.GC()
                runtime.Gosched()
                checkLock.Lock()
                defer checkLock.Unlock()
                runtime.GC()
                numGoroutine := runtime.NumGoroutine()
                if numGoroutine < glsMap.Count()+cpuAddNum {
                    var ids []int64
                    glsMap.Iter(func(k int64, v Map[uint64, any]) (stop bool) {
                        ids = append(ids, k)
                        return false
                    })
                    panic(GlsError{
                        NumGoroutine: numGoroutine,
                        GlsGoIds:     ids,
                    })
                }
            }
        }
        testI = 0
        if testInterval > 1_000_000 {
            testInterval = 1_000_001
        } else {
            testInterval = int(float32(testInterval) * 1.5)
        }
    }
}
