package storage

import "sync"

//基于goroutine的局部存储,类似于threadLocal,必须调用GOClean清理

var (
    glsMap          = NewMap[int64, Map[string, any]](0)
    glsLock         = sync.RWMutex{}
    KnowHowToUseGls = false
)

func GOGet(key string) (any, bool) {
    return getMap(GoID()).Get(key)
}

func GOSet(key string, value string) {
    getMap(GoID()).Put(key, value)
}

func GODel(key string) {
    getMap(GoID()).Delete(key)
}

func GOClean() {
    glsLock.Lock()
    defer glsLock.Unlock()
    glsMap.Delete(GoID())
}

func getMap(id int64) Map[string, any] {
    if !KnowHowToUseGls {
        panic("Set `KnowHowToUseGls=true` after you know how to use it.You must call GOClean!")
    }
    glsLock.Lock()
    defer glsLock.Unlock()
    value, ok := glsMap.Get(id)
    if !ok {
        value = NewMap[string, any](0)
        glsMap.Put(id, value)
    }
    return value
}
