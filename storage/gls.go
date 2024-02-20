package storage

import (
    "runtime"
    "sync"
)

//基于goroutine的局部存储,类似于threadLocal,必须调用GOClean清理,明确后需要设置 KnowHowToUseGls=true

var (
    glsMap          = NewMap[int64, Map[string, any]](0)
    glsLock         = sync.RWMutex{}
    KnowHowToUseGls = false
)

type GlsError struct {
    GoIds []int64
}

func (g GlsError) Error() string {
    return "gls发生泄露,请检查"
}

func (g GlsError) String() string {
    return "gls发生泄露,请检查"
}

func GOSet(key string, value string) {
    if !KnowHowToUseGls {
        panic("Set `KnowHowToUseGls=true` after you know how to use it.You must call GOClean!")
    }
    glsLock.RLock()
    id := GoID()
    m, ok := glsMap.Get(id)
    if !ok {
        glsLock.RUnlock()
        glsLock.Lock()
        m = NewMap[string, any](0)
        glsMap.Put(id, m)
        glsLock.Unlock()
    } else {
        glsLock.RUnlock()
    }
    m.Put(key, value)
    check()
}

func GOGet(key string) (any, bool) {
    glsLock.RLock()
    value, ok := glsMap.Get(GoID())
    glsLock.RUnlock()
    if !ok {
        return nil, false
    }
    return value.Get(key)
}

func GODel(key string) {
    glsLock.RLock()
    value, ok := glsMap.Get(GoID())
    glsLock.RUnlock()
    if !ok {
        return
    }
    value.Delete(key)
}

func GOClean() {
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

var testI = 0
var testInterval = 10

func check() {
    testI++
    if testI > testInterval {
        if runtime.NumGoroutine() < glsMap.Count() {
            var ids []int64
            glsLock.RLock()
            glsMap.Iter(func(k int64, v Map[string, any]) (stop bool) {
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
