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

    //glsMap = NewMap(MapTypeSwiss[int64, Map[uint32, any]](0))
    //glsMap = NewMap(MapTypeGo[int64, Map[uint32, any]](0))
    //glsMap          = NewMap(MapTypeConcurrentWrapper(MapTypeGo[int64, Map[uint32, any]](0)))
    glsMap = NewMap(MapTypeSwissConcurrent[int64, Map[uint32, any]]())
    //glsLock         concurrent.RwLocker = &sync.RWMutex{}
    glsLock concurrent.RwLocker = noLock{}
    //glsSubMapPool                       = sync.Pool{New: func() any { return NewMap(MapTypeSwiss[uint32, any](1)) }}
    glsSubMapPool = sync.Pool{New: func() any { return NewMap(MapTypeArray[uint32, any](2)) }}
)

type noLock struct{}

func (l noLock) Lock()         {}
func (l noLock) Unlock()       {}
func (l noLock) TryLock() bool { return true }

func (l noLock) RLock()         {}
func (l noLock) RUnlock()       {}
func (l noLock) TryRLock() bool { return true }

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

var keyIdGen = uint32(0)

// Key 提供一个类型安全的key
// 不使用struct以提供特殊操作的可能性
type Key[T any] uint32

func (k Key[T]) Get() (value T, exist bool) {
    glsLock.RLock()
    m, ok := glsMap.Get(GoID())
    glsLock.RUnlock()
    if !ok {
        return
    }
    get, o := m.Get(uint32(k))
    if o {
        value = get.(T)
        exist = true
    }
    return
}

func (k Key[T]) GetSimple() (value T) {
    get, _ := k.Get()
    return get
}

func (k Key[T]) MustGet() (value T) {
    get, ok := k.Get()
    if !ok {
        panic("key not set value")
    }
    return get
}

func (k Key[T]) Set(value T) {
    if !knowHowToUseGls {
        panic("call `KnowHowToUseGls()` after you know how to use it! (You must call GlsClean() after code)")
    }
    glsLock.RLock()
    id := GoID()
    m, ok := glsMap.Get(id)
    if !ok {
        glsLock.RUnlock()
        m = glsSubMapPool.Get().(Map[uint32, any])
        glsLock.Lock()
        glsMap.Put(id, m)
        glsLock.Unlock()
    } else {
        glsLock.RUnlock()
    }
    m.Put(uint32(k), any(value))
    check()
}

func (k Key[T]) Delete() {
    glsLock.RLock()
    value, ok := glsMap.Get(GoID())
    glsLock.RUnlock()
    if !ok {
        return
    }
    value.Delete(uint32(k))
}

func (k Key[T]) GlsClean() { GlsClean() }
func KnowHowToUseGls() {
    knowHowToUseGls = true
}
func GlsClean() {
    check()
    glsLock.RLock()
    id := GoID()
    m, ok := glsMap.Get(id)
    glsLock.RUnlock()
    if !ok {
        return
    }
    glsLock.Lock()
    glsMap.Delete(id)
    m.Clean()
    glsSubMapPool.Put(m)
    glsLock.Unlock()
}

// NewGlsItem returns a new unique key.
//
// var item = NewGlsItem[int]() //at global
// value,exist := item.Get()
func NewGlsItem[T any]() Key[T] {
    id := atomic.AddUint32(&keyIdGen, 1)
    if id == 0 {
        keyIdGen = math.MaxUint32 //无限 panic

        // 使用全局定义key不可能超过2^32个
        // 2^32 keys should be enough for everyone
        panic("gls key too much, overflow! You should define keys in global, not in functions")
    }
    return Key[T](id)
}

var testI = 0
var testInterval = 10

//检查gls是否泄露,每次检查都会增加检查间隔,直到检查间隔大于1000000,所以对性能几乎没影响
func check() {
    testI++
    if testI > testInterval {
        if testI == testInterval {
            if runtime.NumGoroutine() < glsMap.Count() {
                runtime.Gosched()
                if runtime.NumGoroutine() < glsMap.Count() {
                    glsLock.Lock()
                    if testI > testInterval && runtime.NumGoroutine() < glsMap.Count() {
                        var ids []int64
                        glsMap.Iter(func(k int64, v Map[uint32, any]) (stop bool) {
                            ids = append(ids, k)
                            return false
                        })
                        glsLock.Unlock()
                        panic(GlsError{
                            NumGoroutine: runtime.NumGoroutine(),
                            GlsGoIds:     ids,
                        })
                    } else {
                        glsLock.Unlock()
                    }
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
