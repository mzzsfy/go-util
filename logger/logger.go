package logger

import (
    "bytes"
    "context"
    "fmt"
    "github.com/mzzsfy/go-util/helper"
    "github.com/mzzsfy/go-util/pool"
    "io"
    "strings"
)

var (
    _ Log = (*logger)(nil)

    // PrintBlankTag 是否打印空白tag
    PrintBlankTag bool
    // PrintYearInfo 打印年份信息,0打印4位,1,打印2位,2不打印
    PrintYearInfo = 1
    // CompressedLogName 是否压缩日志名称
    CompressedLogName = true
    globalPlugin      []Plugin
)

func AddGlobalPlugin(plugin ...Plugin) {
    globalPlugin = append(globalPlugin, plugin...)
}
func GlobalPlugins() []Plugin {
    r := make([]Plugin, len(globalPlugin))
    copy(r, globalPlugin)
    return r
}
func CleanGlobalPlugin() {
    globalPlugin = nil
}

type logger struct {
    level    *Level
    tag      rune
    showName string
    fullName string
    parent   *logger
    plugin   []Plugin
    context  context.Context
    //writer   io.Writer
}

func (l *logger) FullName() string {
    return l.fullName
}

func (l *logger) target() io.Writer {
    //if l.writer != nil {
    //    return l.writer
    //}
    //if l.parent != nil {
    //    return l.parent.target()
    //}
    return writerTarget
}
func (l *logger) Level() Level {
    if l.level != nil && *l.level >= 0 {
        return *l.level
    }
    if l.parent != nil {
        return l.parent.Level()
    }
    return *defaultLevel
}
func (l *logger) SetLevel(lv *Level) Log {
    if lv != nil {
        if l.level == nil {
            l.level = new(Level)
        }
        *l.level = *lv
    } else {
        *l.level = -1
    }
    return l
}

// TF Trace 懒加载
func (l *logger) TF(message string, f func() []any) Log { return l.LF(TraceLevel, message, f) }

// D debug
func (l *logger) D(message string, args ...any) Log { return l.L(DebugLevel, message, args...) }

// DF debug 懒加载
func (l *logger) DF(message string, f func() []any) Log { return l.LF(DebugLevel, message, f) }

// I info
func (l *logger) I(message string, args ...any) Log { return l.L(InfoLevel, message, args...) }

// IF info 懒加载
func (l *logger) IF(message string, f func() []any) Log { return l.LF(InfoLevel, message, f) }

// W warn
func (l *logger) W(message string, args ...any) Log { return l.L(WarnLevel, message, args...) }

// WF warn 懒加载
func (l *logger) WF(message string, f func() []any) Log { return l.LF(WarnLevel, message, f) }

// E error
func (l *logger) E(message string, args ...any) Log { return l.L(ErrorLevel, message, args...) }

// EF error 懒加载
func (l *logger) EF(message string, f func() []any) Log { return l.LF(ErrorLevel, message, f) }

// L 打印日志
func (l *logger) L(lv Level, message string, args ...any) Log {
    if l.Level() <= lv {
        l.doLog(lv, message, args)
    }
    return l
}

// LF 打印日志
func (l *logger) LF(lv Level, message string, f func() []any) Log {
    if l.Level() <= lv {
        l.doLog(lv, message, f())
    }
    return l
}

func (l *logger) Plugin() []Plugin {
    return l.plugin
}

// WithPlugin 产生一个新的Log,当输出日志前,会尝试调用该方法,多次调用只保留最后一个
func (l *logger) WithPlugin(p Plugin) Log {
    if p == nil {
        return l
    }
    nl := logPool.Get()
    *nl = *l
    nl.plugin = append(nl.plugin, p)
    return nl
}

func (l *logger) WithContext(ctx context.Context) Log {
    if ctx == nil && l.context == nil {
        return l
    }
    nl := logPool.Get()
    *nl = *l
    nl.context = ctx
    return nl
}

func (l *logger) UnUse() {
    if l.context == nil && l.plugin == nil {
        return
    }
    if Logger(l.FullName()) == l {
        return
    }
    logPool.Put(l)
}

func (l *logger) Context() context.Context {
    if l.context != nil {
        return l.context
    } else if l.parent != nil {
        return l.parent.Context()
    }
    return nil
}

func (l *logger) doLog(lv Level, message string, args []any) {
    target := l.target()
    w, aw := target.(helper.AsyncWriter)
    format := bfPool.Get()
    if !aw {
        defer bfPool.Put(format)
    }
    if len(args) != 0 {
        if strings.Contains(message, bigParenthesisString) || !strings.Contains(message, percentSignString) {
            doLogFormat1(format, message, args)
        } else {
            doLogFormat2(format, message, args)
        }
    } else {
        format.Write(helper.StringToBytes(message))
    }
    l.beforeWrite(lv, format)
    for i := 0; i < len(globalPlugin); i++ {
        if p, ok := globalPlugin[i].(PluginWrite); ok {
            p.BeforeWrite(lv, format, l, p)
        }
    }
    if format.Len() > 0 {
        start := FormatStart(l, lv)
        if !aw {
            defer bfPool.Put(start)
        }
        start.Write(format.Bytes())
        start.WriteByte('\n')
        if aw {
            w.WriterAsync(start.Bytes(), func() {
                bfPool.Put(start)
                bfPool.Put(format)
            })
        } else {
            target.Write(helper.StringToBytes(start.String()))
        }
    }
}

func (l *logger) beforeWrite(lv Level, format Buffer) {
    for _, plugin := range l.plugin {
        if p, ok := plugin.(PluginWrite); ok {
            p.BeforeWrite(lv, format, l, plugin)
        }
    }
    if l.parent != nil {
        l.parent.beforeWrite(lv, format)
    }
}

var bfPool = func() *pool.BytePool {
    p := pool.NewSimpleBytesPool()
    p.SetMaxCap(128)
    return p
}()

func FormatStart(l *logger, lv Level) *pool.Bytes {
    s := bfPool.Get()
    AppendNowTime(s)
    s.WriteByte('[')
    AppendLoggerName(l, s)
    s.WriteByte(']')
    AppendLevel(lv, s)
    for _, plugin := range globalPlugin {
        if p, ok := plugin.(PluginAddSuffix); ok {
            p.AddSuffix(lv, s, l, plugin)
        }
    }
    s.WriteByte(':')
    s.WriteByte(' ')
    return s
}

func AppendLevel(lv Level, s Buffer) {
    s.Write(helper.StringToBytes(lv.String()))
}

func AppendLoggerName(l *logger, s Buffer) {
    if l.tag > 0 {
        if l.tag < 128 {
            s.WriteByte(byte(l.tag))
        } else {
            s.Write(helper.StringToBytes(string(l.tag)))
        }
    } else if PrintBlankTag {
        s.WriteByte(' ')
    }
    s.Write(helper.StringToBytes(l.showName))
}

var (
    percentSignString    = "%"
    bigParenthesisString = "{}"
    bigParenthesis       = helper.StringToBytes(bigParenthesisString)
)

//{}占位符风格
func doLogFormat1(s Buffer, format string, args []any) {
    //todo: 修改实现,性能有较大提升空间
    l := len(args)
    split := bytes.Split(helper.StringToBytes(format), bigParenthesis)
    sl := len(split) - 1
    for i := 0; i < sl; i++ {
        s.Write(split[i])
        if l > i {
            appendAny(s, args[i])
        } else {
            s.Write(bigParenthesis)
        }
    }
    s.Write(split[sl])
    if l > sl {
        if e, ok := args[l-1].(error); ok {
            args = args[:l-1]
            l = len(args)
            for i := sl; i < l; i++ {
                s.WriteByte(' ')
                appendAny(s, args[i])
            }
            s.WriteByte(' ')
            s.Write(helper.StringToBytes(fmt.Sprintf("%+v", e)))
        } else {
            for i := sl; i < l; i++ {
                s.WriteByte(' ')
                appendAny(s, args[i])
            }
        }
    }
}

//Format风格
func doLogFormat2(s Buffer, format string, args []any) {
    count := strings.Count(format, percentSignString)
    var err error
    if len(args) > count {
        l := len(args)
        if l > 0 {
            if e, ok := args[l-1].(error); ok {
                err = e
                args = args[:l-1]
            }
        }
    }
    s.Write(helper.StringToBytes(fmt.Sprintf(format, args...)))
    if err != nil {
        s.WriteByte(' ')
        s.Write(helper.StringToBytes(fmt.Sprintf("%+v", err)))
    }
}
