package logger

import (
    "io"
    "strings"
    "sync"
    "sync/atomic"
    "unicode/utf8"

    "github.com/mzzsfy/go-util/pool"
    "github.com/mzzsfy/go-util/storage"
)

type Option func(*Log) *Log

// levelUnset 哨兵值: 继承父级
const levelUnset int32 = -1

var (
    logPool = pool.NewObjectPool[logger](func() *logger {
        l := &logger{}
        l.local = levelUnset
        l.resolved = int32(InfoLevel)
        return l
    }, func(l *logger) {
        l.local = levelUnset
        l.resolved = int32(InfoLevel)
        l.levelGen = 0
        l.localWriter = nil
        l.showName = nil
        l.fullName = ""
        l.parent = nil
        l.kvs = nil
        l.kvPrefix = nil
        l.derived = false
    })
    globalLog               = storage.NewMap(storage.MapTypeSwiss[string, *logger](32))
    globalLock              = sync.RWMutex{}
    showNameMaxLength int32 = 18
    maxLoggerCount    int32 = 10000 // 最大logger数量, 0=无限制
)

// SetMaxLoggerCount 设置最大logger数量, 0表示无限制
func SetMaxLoggerCount(n int) {
    if n < 0 {
        n = 0
    }
    atomic.StoreInt32(&maxLoggerCount, int32(n))
}

// SetLogNameMaxLength 设置日志名称的最大长度,默认18
func SetLogNameMaxLength(length int) {
    if length < 4 {
        length = 4
    }
    globalLock.Lock()
    atomic.StoreInt32(&showNameMaxLength, int32(length))
    globalLog.Iter(func(key string, value *logger) bool {
        fullName := value.FullName()
        names := strings.Split(fullName, ".")
        newShowName := initName(len(names)-1, names, fullName)
        value.setShowName(newShowName)
        return false
    })
    globalLock.Unlock()
}

// AllLogger 获取所有的logger 名称
func AllLogger() func(t func(string)) {
    return func(t func(string)) {
        globalLock.RLock()
        var keys []string
        globalLog.Iter(func(key string, value *logger) bool {
            keys = append(keys, key)
            return false
        })
        globalLock.RUnlock()
        for _, key := range keys {
            t(key)
        }
    }
}

// Logger 获取一个新的log对象,name规则xx.xx.xxx
func Logger(name string, options ...Option) *Log {
    globalLock.RLock()
    l, ok := globalLog.Get(name)
    if !ok {
        globalLock.RUnlock()
        globalLock.Lock()
        if _, ok := globalLog.Get(name); !ok {
            // 检查logger数量限制
            if limit := atomic.LoadInt32(&maxLoggerCount); limit > 0 && globalLog.Count() >= int(limit) {
                // 超出限制: 创建临时logger但不缓存, 每次调用返回新实例
                // 为避免锁外访问globalLog, 不查找父级, 直接使用默认级别
                globalLock.Unlock()
                nl := logPool.Get()
                names := strings.Split(name, ".")
                nl.setShowName(initName(len(names)-1, names, name))
                nl.setFullName(name)
                // 临时logger不关联父级, 使用默认级别
                atomic.StoreInt32(&nl.resolved, atomic.LoadInt32(&defaultLevel))
                atomic.StoreInt32(&nl.levelGen, atomic.LoadInt32(&levelGeneration))
                nl.derived = true // 标记为可回收
                for _, option := range options {
                    nl = option(nl)
                }
                return nl
            }
            names := strings.Split(name, ".")
            for i := range names {
                fullname := strings.Join(names[:i+1], ".")
                if _, ok := globalLog.Get(fullname); ok {
                    continue
                }
                showName := initName(i, names, fullname)
                l1 := logPool.Get()
                l1.setShowName(showName)
                l1.setFullName(fullname)
                l1.parent = globalLog.GetSimple(strings.Join(names[:i], "."))
                if l1.parent != nil {
                    atomic.StoreInt32(&l1.resolved, int32(l1.parent.resolveLevel()))
                } else {
                    atomic.StoreInt32(&l1.resolved, atomic.LoadInt32(&defaultLevel))
                }
                atomic.StoreInt32(&l1.levelGen, atomic.LoadInt32(&levelGeneration))
                globalLog.Put(fullname, l1)
            }
        }
        l = globalLog.GetSimple(name)
        globalLock.Unlock()
    } else {
        globalLock.RUnlock()
    }
    for _, option := range options {
        l = option(l)
    }
    return l
}

func initName(step int, names []string, fullname string) string {
    maxLen := int(atomic.LoadInt32(&showNameMaxLength))
    showName := fullname
    rLen := utf8.RuneCountInString(showName)
    if rLen <= maxLen {
        if rLen < maxLen {
            return strings.Repeat(" ", maxLen-rLen) + showName
        }
        return showName
    }

    if atomic.LoadInt32(&compressedLogName) != 0 {
        newNames := append([]string(nil), names[:step+1]...)
        for i := 0; i < len(newNames)-1; i++ {
            if runes := []rune(newNames[i]); len(runes) > 0 {
                newNames[i] = string(runes[:1])
            }
            showName = strings.Join(newNames, ".")
            rLen = utf8.RuneCountInString(showName)
            if rLen <= maxLen {
                break
            }
        }
    }

    runes := []rune(showName)
    if len(runes) > maxLen {
        idx := len(runes) - maxLen + 3
        if idx < 0 {
            idx = 0
        }
        if idx < len(runes) {
            if runes[idx] == '.' {
                showName = string(runes[:2]) + ".." + string(runes[idx+1:])
            } else {
                showName = string(runes[0]) + ".." + string(runes[idx:])
            }
        } else {
            showName = string(runes[:maxLen])
        }
    }

    rLen = utf8.RuneCountInString(showName)
    if rLen < maxLen {
        showName = strings.Repeat(" ", maxLen-rLen) + showName
    }
    return showName
}

// SetLevel 通过名称设置指定 logger 的级别
// 传递 LevelUnset 清除本地设置, 继承父级或默认级别
func SetLevel(name string, lv Level) {
    globalLock.RLock()
    if l, ok := globalLog.Get(name); ok {
        l.SetLevel(lv)
    }
    globalLock.RUnlock()
}

// SetLevelRecursive 递归设置 logger 及其所有子 logger 的级别
// 传递 LevelUnset 清除本地设置, 继承父级或默认级别
func SetLevelRecursive(name string, lv Level) {
    globalLock.RLock()
    defer globalLock.RUnlock()
    prefix := name + "."
    globalLog.Iter(func(key string, value *logger) bool {
        if key == name || strings.HasPrefix(key, prefix) {
            value.SetLevel(lv)
        }
        return false
    })
}

// RemoveLogger 移除指定logger及其子logger
// 注意: 移除后如果有代码持有旧引用, 仍可使用但不再被全局管理, 最终会被GC回收
func RemoveLogger(name string) {
    globalLock.Lock()
    prefix := name + "."
    var toRemove []string
    globalLog.Iter(func(key string, value *logger) bool {
        if key == name || strings.HasPrefix(key, prefix) {
            toRemove = append(toRemove, key)
        }
        return false
    })
    for _, key := range toRemove {
        globalLog.Delete(key)
    }
    globalLock.Unlock()
}

// WithWriter 创建使用独立 writer 的 logger
func WithWriter(w io.Writer) Option {
    wtv := newWriterTargetValue(w)
    return func(l *Log) *Log {
        nl := l.derive()
        nl.localWriter = &wtv
        return nl
    }
}

// derive 从当前 logger 派生新实例 (仅继承名称和级别)
// 派生logger不设parent, 避免被Unuse后parent链指向pool重用对象
// 级别在创建时快照, 不参与运行时动态继承
func (l *logger) derive() *logger {
    nl := logPool.Get()
    atomic.StorePointer(&nl.showName, atomic.LoadPointer(&l.showName))
    nl.fullName = l.fullName
    // 不设parent: 派生logger级别在创建时固定, Unuse后不会产生悬挂引用
    nl.local = levelUnset
    atomic.StoreInt32(&nl.resolved, atomic.LoadInt32(&l.resolved))
    atomic.StoreInt32(&nl.levelGen, atomic.LoadInt32(&levelGeneration))
    nl.localWriter = l.localWriter
    nl.derived = true
    return nl
}
