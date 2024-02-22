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
    globalLog  = storage.NewMap[string, *logger]()
    globalLock = sync.RWMutex{}
    // ShowNameMaxLength 默认日志名称的最大长度
    ShowNameMaxLength = 18
)

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
                //xxx.xxx.xxx.xxx => x.x.xxx.xxx
                if len([]rune(showName)) > ShowNameMaxLength {
                    for i := range newNames[:len(newNames)-1] {
                        newNames[i] = string([]rune(newNames[i])[:1])
                        showName = strings.Join(newNames, ".")
                        if len(showName) <= ShowNameMaxLength {
                            break
                        }
                    }
                    //x.x.xxx.xxx => ..x.xxx.xxx
                    showName1 := []rune(showName)
                    if len(showName1) > ShowNameMaxLength {
                        showName = ".." + string(showName1[len(showName1)-16:])
                    }
                }
                if len([]rune(showName)) < ShowNameMaxLength {
                    showName = strings.Repeat(" ", ShowNameMaxLength-len([]rune(showName))) + showName
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
