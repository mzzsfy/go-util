package logger

import (
    "strings"
    "sync"
    "sync/atomic"
)

// 命名 Log 全局注册表, 支持点分隔的父子继承
var (
    registryMu sync.RWMutex
    registry   = make(map[string]*Log)
)

// Logger 按名称获取或创建命名 Log
// name 用点分隔自动建立父子继承: Get("app.user") 确保 app 和 app.user 都存在, 后者继承前者级别
// 同名重复调用返回同一实例, opts 每次调用都会应用
func Logger(name string, opts ...Option) *Log { return Get(name, opts...) }

//Get 同 Logger
func Get(name string, opts ...Option) *Log {
    if name == "" {
        return New("", opts...)
    }
    registryMu.RLock()
    l, ok := registry[name]
    registryMu.RUnlock()
    if ok {
        for _, opt := range opts {
            opt(l)
        }
        return l
    }

    registryMu.Lock()
    defer registryMu.Unlock()
    if l, ok := registry[name]; ok {
        for _, opt := range opts {
            opt(l)
        }
        return l
    }
    l = createNamedLocked(name)
    for _, opt := range opts {
        opt(l)
    }
    return l
}

// createNamedLocked 创建命名 Log 及其所有中间层级, 调用方需持有 registryMu 写锁
func createNamedLocked(name string) *Log {
    parts := strings.Split(name, ".")
    for i := range parts {
        fullname := strings.Join(parts[:i+1], ".")
        if _, ok := registry[fullname]; ok {
            continue
        }
        var parent *Log
        if i > 0 {
            parent = registry[strings.Join(parts[:i], ".")]
        }
        var resolved int32
        if parent != nil {
            resolved = int32(parent.level())
        } else {
            resolved = atomic.LoadInt32(&defaultLevel)
        }
        registry[fullname] = &Log{
            name:     fullname,
            writer:   loadDefaultWriter(),
            parent:   parent,
            localLv:  levelInherit,
            resolved: resolved,
            gen:      atomic.LoadInt32(&globalGen),
        }
        if _, ok := registry[fullname].writer.(asyncWriter); ok {
            registry[fullname].isAsync = true
        }
    }
    return registry[name]
}

// SetLevelByName 按名称设置 Log 级别, 返回是否存在该 Log
func SetLevelByName(name string, lv Level) bool {
    registryMu.RLock()
    defer registryMu.RUnlock()
    if l, ok := registry[name]; ok {
        l.SetLevel(lv)
        return true
    }
    return false
}

// SetLevel 旧版兼容, 等价于 SetLevelByName
func SetLevel(name string, lv Level) { SetLevelByName(name, lv) }

// SetLevelRecursive 递归设置 Log 及其所有子 Log 的级别
func SetLevelRecursive(name string, lv Level) {
    registryMu.RLock()
    defer registryMu.RUnlock()
    prefix := name + "."
    for key, l := range registry {
        if key == name || strings.HasPrefix(key, prefix) {
            atomic.StoreInt32(&l.localLv, int32(lv))
            atomic.StoreInt32(&l.resolved, int32(lv))
        }
    }
    atomic.AddInt32(&globalGen, 1)
}

// RemoveLogger 移除指定 Log 及其子 Log
// 已持有的引用仍可使用, 但不再被全局管理
func RemoveLogger(name string) {
    registryMu.Lock()
    defer registryMu.Unlock()
    prefix := name + "."
    for key := range registry {
        if key == name || strings.HasPrefix(key, prefix) {
            delete(registry, key)
        }
    }
}

// AllLogger 遍历所有已创建的命名 Log
func AllLogger(visit func(name string)) {
    registryMu.RLock()
    defer registryMu.RUnlock()
    for name := range registry {
        visit(name)
    }
}

// RegistryCount 返回已注册的命名 Log 数量
func RegistryCount() int {
    registryMu.RLock()
    defer registryMu.RUnlock()
    return len(registry)
}
