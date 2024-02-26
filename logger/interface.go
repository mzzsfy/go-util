package logger

import (
    "context"
    "io"
)

type Log interface {
    FullName() string
    Level() Level
    SetLevel(*Level) Log
    T(message string, args ...any) Log
    TF(message string, f func() []any) Log
    D(message string, args ...any) Log
    DF(message string, f func() []any) Log
    I(message string, args ...any) Log
    IF(message string, f func() []any) Log
    W(message string, args ...any) Log
    WF(message string, f func() []any) Log
    E(message string, args ...any) Log
    EF(message string, f func() []any) Log
    L(lv Level, message string, args ...any) Log
    LF(lv Level, message string, f func() []any) Log
    // WithPlugin 产生一个新的,使用该Plugin的Log
    WithPlugin(Plugin) Log
    Plugin() []Plugin
    // WithContext 产生一个新的,使用该Context的Log
    WithContext(context.Context) Log
    Context() context.Context
    // UnUse 标记当前log不再使用
    UnUse()
}

type Plugin interface{}

type Buffer interface {
    io.Writer
    io.StringWriter
    io.ByteWriter
    Bytes() []byte
    String() string
    Len() int
    Cap() int
    Reset()
}

type PluginAddSuffix interface {
    Plugin
    // AddSuffix 添加后缀
    //"2024-01-01 01:01:00.832[                test]I: test"
    // 变为
    //"2024-01-01 01:01:00.832[                test]I xxxxx: test"
    AddSuffix(Level, Buffer, Log, Plugin)
}

type PluginWrite interface {
    Plugin
    // BeforeWrite 在实际写入日志前调用,可以修改日志内容
    // "2024-01-01 01:01:00.832[                test]I: 这是可修改部分"
    BeforeWrite(Level, Buffer, Log, Plugin)
}

//type PluginChangeTarget interface {
//    Plugin
//    ChangeTarget(Level, Log, Plugin) io.Writer
//}
