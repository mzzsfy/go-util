package storage

import (
    "github.com/mzzsfy/go-util/unsafe"
    "math"
    "runtime"
    "sync"
    "sync/atomic"
)

var GoID = unsafe.GoID

//基于goroutine的局部存储,类似于threadLocal,必须调用GOClean清理,明确后需要设置 KnowHowToUseGls=true

var (
    glsMap          = NewMap(MapTypeSwiss[int64, Map[uint32, any]](0))
    glsLock         = sync.RWMutex{}
    KnowHowToUseGls = false
)

type GlsError struct {
    GoIds []int64
}

func (g GlsError) Error() string {
    return "gls发生泄露,请检查!"
}

func (g GlsError) String() string {
    return "gls发生泄露,请检查!"
}

var keyIdGen = uint32(0)

// Key 提供一个类型安全的key
//不使用struct以提供特殊操作的可能性
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
    if !KnowHowToUseGls {
        panic("Set `KnowHowToUseGls=true` after you know how to use it.You must call GlsClean!")
    }
    glsLock.RLock()
    id := GoID()
    m, ok := glsMap.Get(id)
    if !ok {
        glsLock.RUnlock()
        glsLock.Lock()
        m = NewMap(MapTypeSwiss[uint32, any](2))
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

func GlsClean() {
    check()
    glsLock.RLock()
    id := GoID()
    _, ok := glsMap.Get(id)
    glsLock.RUnlock()
    if !ok {
        return
    }
    glsLock.Lock()
    glsMap.Delete(id)
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
        panic("key too much, overflow! You should define keys in global, not in functions")
    }
    return Key[T](id)
}

var testI = 0
var testInterval = 10

//检查gls是否泄露,每次检查都会增加检查间隔,直到检查间隔大于1000000,所以对性能几乎没影响
func check() {
    testI++
    if testI > testInterval {
        if runtime.NumGoroutine() < glsMap.Count() {
            var ids []int64
            glsLock.RLock()
            glsMap.Iter(func(k int64, v Map[uint32, any]) (stop bool) {
                ids = append(ids, k)
                return false
            })
            glsLock.RUnlock()
            panic(GlsError{GoIds: ids})
        }
        testI = 0
        if testInterval > 1_000_000 {
            testInterval = 1_000_001
        } else {
            testInterval = int(float32(testInterval) * 1.5)
        }
    }
}
