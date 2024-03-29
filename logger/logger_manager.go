package logger

import (
    "context"
    "github.com/mzzsfy/go-util/pool"
    "github.com/mzzsfy/go-util/storage"
    "strings"
    "sync"
)

type Option func(*logger) *logger

var (
    logPool = pool.NewObjectPool[logger](func() *logger {
        level := new(Level)
        *level = -1
        return &logger{
            level: level,
        }
    }, func(l *logger) {
        l.level = nil
        l.tag = 0
        l.showName = ""
        l.fullName = ""
        l.parent = nil
        l.plugin = nil
        l.context = nil
    })
    globalLog  = storage.NewMap(storage.MapTypeSwiss[string, *logger](32))
    globalLock = sync.RWMutex{}
    // showNameMaxLength 默认日志名称的最大长度
    showNameMaxLength = 18
)

// SetLogNameMaxLength 设置日志名称的最大长度,默认18
func SetLogNameMaxLength(length int) {
    globalLock.Lock()
    globalLock.Unlock()
    showNameMaxLength = length
    globalLog.Clean()
}

// AllLogger 获取所有的logger 名称,建议搭配 seq包使用
// seq.From(logger.AllLogger()).XXX
func AllLogger() func(t func(string)) {
    return func(t func(string)) {
        globalLock.RLock()
        defer globalLock.RUnlock()
        globalLog.Iter(func(key string, value *logger) bool {
            globalLock.RUnlock()
            defer globalLock.Lock()
            t(key)
            return false
        })
    }
}

// Logger 获取一个新的log对象,name规则xx.xx.xxx
func Logger(name string, options ...Option) Log {
    globalLock.RLock()
    l, ok := globalLog.Get(name)
    if !ok {
        globalLock.RUnlock()
        globalLock.Lock()
        if _, ok := globalLog.Get(name); !ok {
            names := strings.Split(name, ".")
            for i := range names {
                fullname := strings.Join(names[:i+1], ".")
                if _, ok := globalLog.Get(fullname); ok {
                    continue
                }
                var newNames []string
                for j := 0; j < i+1; j++ {
                    newNames = append(newNames, names[j])
                }
                showName := fullname
                if CompressedLogName {
                    //xxx.xxx.xxx.xxx => x.x.xxx.xxx
                    if len([]rune(showName)) > showNameMaxLength {
                        for i := range newNames[:len(newNames)-1] {
                            newNames[i] = string([]rune(newNames[i])[:1])
                            showName = strings.Join(newNames, ".")
                            if len(showName) <= showNameMaxLength {
                                break
                            }
                        }
                        //x.x.xxx.xxx => x..xxx.xxx
                        showName1 := []rune(showName)
                        if len(showName1) > showNameMaxLength {
                            idx := len(showName1) - showNameMaxLength + 3
                            if showName[idx] == '.' {
                                showName = string(fullname[:2]) + ".." + string(showName1[idx+1:])
                            } else {
                                showName = string(showName1[0]) + ".." + string(showName1[idx:])
                            }
                        }
                    }
                } else {
                    showName1 := []rune(showName)
                    if len(showName1) > showNameMaxLength {
                        idx := len(showName1) - showNameMaxLength + 3
                        if showName[idx] == '.' {
                            showName = string(showName1[:2]) + ".." + string(showName1[idx+1:])
                        } else {
                            showName = string(showName1[0]) + ".." + string(showName1[idx:])
                        }
                    }
                }
                if len([]rune(showName)) < showNameMaxLength {
                    showName = strings.Repeat(" ", showNameMaxLength-len([]rune(showName))) + showName
                }
                l1 := logPool.Get()
                l1.showName = showName
                l1.fullName = fullname
                l1.parent = globalLog.GetSimple(strings.Join(names[:i], "."))
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

func WithSetPlugin(plugin ...Plugin) Option {
    return func(l *logger) *logger {
        l.plugin = append(l.plugin, plugin...)
        return l
    }
}

func WithSetContext(ctx context.Context) Option {
    return func(l *logger) *logger {
        l.context = ctx
        return l
    }
}

// WithTag 在名字前添加一个符号,如@,*,#等,用来做醒目区分
func WithTag(tag rune) Option {
    return func(l *logger) *logger {
        nl := logPool.Get()
        *nl = *l
        nl.tag = tag
        return nl
    }
}

func WithParentPlugin() Option {
    return func(l *logger) *logger {
        if l.parent == nil {
            return l
        }
        nl := logPool.Get()
        *nl = *l
        nl.plugin = append(nl.plugin, l.parent.plugin...)
        return nl
    }
}

func WithPlugin(plugin ...Plugin) Option {
    return func(l *logger) *logger {
        if len(plugin) == 0 {
            return l
        }
        nl := logPool.Get()
        *nl = *l
        nl.plugin = append(nl.plugin, plugin...)
        return nl
    }
}
